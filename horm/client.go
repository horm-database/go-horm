// Copyright (c) 2024 The horm-database Authors. All rights reserved.
// This file Author:  CaoHao <18500482693@163.com> .
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package horm mysql 查询语句封装包
package horm

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/horm-database/common/consts"
	"github.com/horm-database/common/crypto"
	"github.com/horm-database/common/errs"
	"github.com/horm-database/common/json"
	"github.com/horm-database/common/proto"
	"github.com/horm-database/common/snowflake"
	"github.com/horm-database/common/types"
	"github.com/horm-database/common/util"
	"github.com/horm-database/go-horm/horm/client"
)

var GlobalClient Client // 全局查询语句执行客户端

// Client 查询语句执行客户端
type Client interface {
	Exec(ctx context.Context, q *Query, retReceiver ...interface{}) (isNil bool, err error)
	PExec(ctx context.Context, q *Query) error
	CompExec(ctx context.Context, q *Query, retReceiver interface{}) error
}

// NewClient 创建查询语句执行客户端
// param: name 配置名
// param: opts 参数配置
func NewClient(name string, opts ...Option) Client {
	return &cli{
		name: name,
		opts: opts,
		c:    client.DefaultClient,
	}
}

// cli 查询语句执行客户端 Client 实现
type cli struct {
	name string
	opts []Option
	c    *client.Client
}

// SetGlobalClient 设置全局查询语句执行客户端。
// param: name 调用名，取 orm.yaml 配置的 server.caller.name
// param: opts 参数配置
func SetGlobalClient(name string, opts ...Option) {
	GlobalClient = NewClient(name, opts...)
}

// Exec 单执行单元 result 接收结果的指针，可以不传，最多一个
func (o *cli) Exec(ctx context.Context, q *Query, retReceiver ...interface{}) (isNil bool, err error) {
	header, result, err := o.exec(ctx, consts.QueryModeSingle, q)
	if err != nil {
		return false, err
	}

	if header.Err != nil && header.Err.Code != 0 {
		return false, header.Err.ToError()
	}

	if header.IsNil {
		return true, nil
	}

	if len(retReceiver) == 0 && (q.Unit.Op == consts.OpExists || q.Unit.Op == consts.OpHExists) {
		var exists bool
		err = q.GetCoder().Decode(q.ResultType, result, []interface{}{&exists})
		if err != nil {
			return false, errs.Newf(errs.ErrClientDecode, "[request_id=%d] %v, result=[%s]",
				q.RequestID, err, types.ToString(result))
		}
		if exists {
			return false, nil
		} else {
			return true, nil
		}
	}

	err = q.GetCoder().Decode(q.ResultType, result, retReceiver)
	if err != nil {
		return false, errs.Newf(errs.ErrClientDecode, "[request_id=%d] %v, result=[%s]",
			q.RequestID, err, types.ToString(result))
	}

	return header.IsNil, nil
}

// PExec 执行并行查询（多个执行单元并发，没有嵌套子查询）
func (o *cli) PExec(ctx context.Context, q *Query) error {
	header, result, err := o.exec(ctx, consts.QueryModeParallel, q)
	if err != nil {
		return err
	}

	rspData := map[string]interface{}{}
	err = json.Api.Unmarshal(result, &rspData)
	if err != nil {
		return errs.Newf(errs.ErrClientDecode, "[request_id=%d] "+
			"result decode to ret receiver error: %v, resp=[%s]", q.RequestID, err, string(result))
	}

	query := q.GetHead()
	for {
		if query == nil {
			return nil
		}

		if header.RspErrs != nil {
			rspErr, ok := header.RspErrs[query.Key]
			if ok && rspErr != nil && rspErr.Code != 0 {
				if query.RespError != nil {
					*query.RespError = &errs.Error{
						Type: errs.EType(rspErr.Type),
						Code: int(rspErr.Code),
						Msg:  rspErr.Msg,
					}
				}

				query = query.next
				continue
			}
		}

		if header.RspNils != nil {
			isNil, ok := header.RspNils[query.Key]
			if ok && isNil && query.IsNil != nil {
				*query.IsNil = true
			}
		}

		if ret, ok := rspData[query.Key]; ok {
			err = query.GetCoder().Decode(query.ResultType, ret, query.Receiver)
			if err != nil && query.RespError != nil {
				*query.RespError = errs.Newf(errs.ErrClientDecode,
					"[request_id=%d] %v, result=[%s]", query.RequestID, err, types.ToString(result))
			}
		}

		query = query.next
	}
}

// CompExec 执行复合查询（包含嵌套子查询）
func (o *cli) CompExec(ctx context.Context, q *Query, retReceiver interface{}) error {
	header, result, err := o.exec(ctx, consts.QueryModeCompound, q)
	if err != nil {
		return err
	}

	if header.Err != nil && header.Err.Code != 0 {
		return &errs.Error{
			Type: errs.EType(header.Err.Type),
			Code: int(header.Err.Code),
			Msg:  header.Err.Msg,
		}
	}

	err = json.Api.Unmarshal(result, retReceiver)
	if err != nil {
		return errs.Newf(errs.ErrClientDecode, "[request_id=%d] "+
			"result decode to ret receiver error: %v, resp=[%s]", q.RequestID, err, string(result))
	}

	return nil
}

func (o *cli) exec(ctx context.Context, mode uint32, q *Query) (*proto.ResponseHeader, []byte, error) {
	q = q.GetHead()

	units, err := createUnits(q)
	if err != nil {
		return nil, nil, err
	}

	q.RequestBody, err = json.Api.Marshal(units)
	if err != nil {
		return nil, nil, errs.New(errs.ErrClientEncode, "client unit marshal error: "+err.Error())
	}

	opts := getOptions(o.name)

	if len(o.opts) > 0 {
		for _, opt := range o.opts {
			opt(opts)
		}
	}

	timeout := opts.Timeout
	if deadline, ok := ctx.Deadline(); ok { // 如果 context 超时时间比数据库设置的超时时间要短，则取 context 超时时间。
		leftMillisecond := uint32(deadline.Sub(time.Now()) / time.Millisecond)
		if leftMillisecond < opts.Timeout {
			timeout = leftMillisecond
		}
	}

	head := proto.RequestHeader{}
	head.Version = Version
	head.QueryMode = mode
	head.Timestamp = uint64(time.Now().UnixMilli())
	head.Timeout = timeout

	if opts.Caller != "" {
		head.Caller = opts.Caller
	} else {
		head.Caller = opts.Name
	}

	head.Callee = "server.access.api/Query"
	head.Appid = opts.Appid
	head.Ip = opts.LocalIP
	head.AuthRand = uint32(rand.Intn(99999999))

	if head.Ip == "" {
		head.Ip = util.GetLocalIP()
	}

	if q.Compress { // 压缩
		head.Compress = consts.Compression
	}

	if q.RequestID != 0 {
		head.RequestId = q.RequestID
	} else {
		head.RequestId = snowflake.GenerateID()
		q.RequestID = head.RequestId
	}

	if q.TraceID != "" {
		head.TraceId = q.TraceID
	}

	// 签名
	md5Str := fmt.Sprintf("%d%s%d%d%d%s%d%d%s%d%d%d", head.Appid, opts.Secret,
		head.RequestType, head.QueryMode, head.RequestId, head.TraceId, head.Timestamp,
		head.Timeout, head.Caller, head.Compress, head.AuthRand, head.Version)

	head.Sign = crypto.MD5Str(md5Str)

	reqParam := client.ReqParam{
		WorkspaceID: opts.WorkspaceID,
		Encryption:  opts.Encryption,
		Token:       opts.Token,
		Target:      opts.Target,
	}

	reqParam.Location.Region = opts.Location.Region
	reqParam.Location.Zone = opts.Location.Zone
	reqParam.Location.Compus = opts.Location.Compus

	return o.c.Invoke(ctx, &head, q.RequestBody, &reqParam)
}
