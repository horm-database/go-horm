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

package horm

import (
	"context"

	"github.com/horm-database/common/consts"
	"github.com/horm-database/common/errs"
	"github.com/horm-database/common/proto"
	"github.com/horm-database/go-horm/horm/codec"
)

// Query 请求语句
type Query struct {
	Unit        *proto.Unit    // 请求单元
	first       *Query         // 首查询
	last        *Query         // 并行查询上一个查询
	next        *Query         // 并行查询下一个查询
	sub         *Query         // 子查询
	parent      *Query         // 父查询
	trans       *Query         // 事务语句
	Client      Client         // 客户端
	Error       error          // Query 语句错误
	Key         string         // 语句 key
	Receiver    []interface{}  // 结果接收
	IsNil       *bool          // 是否包含空值
	RespError   *error         // 返回错误
	Compress    bool           // 是否压缩 false - 不压缩 true - 压缩
	ResultType  consts.RetType // 返回数据类型
	Coder       codec.Codec    // 数据编解码器
	RequestID   uint64         // 请求 id
	TraceID     string         // 请求 trace_id
	RequestBody []byte         // 请求体
}

// Reset 语句初始化
func (s *Query) Reset() *Query {
	s.Unit = &proto.Unit{Size: -1} //未传值
	s.first = s
	s.last = nil
	s.next = nil
	s.sub = nil
	s.parent = nil
	s.trans = nil
	s.Client = nil
	s.Error = nil
	s.Key = ""
	s.Receiver = nil
	s.IsNil = nil
	s.RespError = nil
	s.Compress = false
	s.ResultType = 0
	s.Coder = nil
	s.RequestID = 0
	s.TraceID = ""
	s.RequestBody = []byte{}

	return s
}

// NewQuery 创建新执行语句
// param: name 为执行语句名称
func NewQuery(name string) *Query {
	var unit = &proto.Unit{Size: -1} //未传值

	s := &Query{Unit: unit}
	s.first = s

	if len(name) > 0 {
		s.Name(name)
	}

	return s
}

// NewTransaction 创建一个事务语句
// param: trans 为该事务所有执行语句，所有语句必须全部成功或失败。注意：仅支持带事务功能的数据库回滚。
func NewTransaction(transName string, trans *Query) *Query {
	var unit = &proto.Unit{Name: transName, Size: -1} //未传值

	s := &Query{Unit: unit, trans: trans.GetHead()}
	s.first = s

	s.Op("transaction")

	return s
}

// Next 创建下一条并发执行语句
// param: name 为执行语句名称
func (s *Query) Next(name string) *Query {
	next := NewQuery(name)
	next.first = s.first

	next.last = s
	s.next = next
	return next
}

// AddNext 将输入的语句串联到本语句并行执行上。
// param: nexts 需要被串联的并行执行的所有语句
func (s *Query) AddNext(nexts ...*Query) *Query {
	if len(nexts) == 0 {
		return nil
	}

	last := s
	for _, v := range nexts {
		next := v.GetHead()
		next.first = s.first
		next.last = last
		last.next = next
		last = next

		for {
			if last.next == nil {
				break
			}
			last = next.next
		}
	}

	return last
}

// AddSub 新增嵌套子查询语句
func (s *Query) AddSub(sub *Query) *Query {
	s.sub = sub.first
	sub.parent = s
	return s
}

// GetHead 获取头部
func (s *Query) GetHead() *Query {
	var p = s
	for {
		if p.parent == nil {
			return p.first
		}
		p = p.parent
	}
}

// WithCoder 更换编解码器
func (s *Query) WithCoder(coder codec.Codec) *Query {
	s.Coder = coder
	return s
}

// GetCoder 获取编解码器
func (s *Query) GetCoder() codec.Codec {
	if s.Coder != nil {
		return s.Coder
	}

	//返回默认编码器
	return codec.DefaultCodec
}

// WithReceiver 接收返回，主要用于并发查询
func (s *Query) WithReceiver(isNil *bool, err *error, receiver ...interface{}) *Query {
	s.IsNil = isNil
	s.RespError = err
	s.Receiver = receiver
	return s
}

// Exec 执行单个操作单元
func (s *Query) Exec(ctx context.Context, retReceiver ...interface{}) (isNil bool, err error) {
	head := s.GetHead()
	if head.Client != nil {
		return head.Client.Exec(ctx, head, retReceiver...)
	} else if GlobalClient != nil {
		return GlobalClient.Exec(ctx, head, retReceiver...)
	}
	return false, errs.Newf(errs.ErrClientNotInit, "client is not init")
}

// PExec 并行执行多个操作单元
func (s *Query) PExec(ctx context.Context) error {
	head := s.GetHead()
	if head.Client != nil {
		return head.Client.PExec(ctx, head)
	} else if GlobalClient != nil {
		return GlobalClient.PExec(ctx, head)
	}
	return errs.Newf(errs.ErrClientNotInit, "client is not init")
}

// CompExec 复合查询
func (s *Query) CompExec(ctx context.Context, retReceiver interface{}) error {
	head := s.GetHead()
	if head.Client != nil {
		return head.Client.CompExec(ctx, head, retReceiver)
	} else if GlobalClient != nil {
		return GlobalClient.CompExec(ctx, head, retReceiver)
	}
	return errs.Newf(errs.ErrClientNotInit, "client is not init")
}

// WithClient 设置查询客户端
func (s *Query) WithClient(c Client) *Query {
	s.Client = c
	return s
}

// SetCompress 压缩，调用该方法表示数据将通过Gzip压缩传递
func (s *Query) SetCompress() *Query {
	s.Compress = true
	return s
}

// WithRequestID 设置 request_id
func (s *Query) WithRequestID(id uint64) *Query {
	s.RequestID = id
	return s
}

// WithTraceID 设置 trace_id
func (s *Query) WithTraceID(id string) *Query {
	s.TraceID = id
	return s
}
