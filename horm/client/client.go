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

// Package client is horm-go client,
// including network transportation, resolving, routing etc.
package client

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/horm-database/common/codec"
	"github.com/horm-database/common/errs"
	"github.com/horm-database/common/metrics"
	"github.com/horm-database/common/naming"
	"github.com/horm-database/common/proto"
	"github.com/horm-database/common/types"
)

// DefaultClient 默认通用客户端（thread-safe）
var DefaultClient = &Client{}

type Client struct{}

type ReqParam struct {
	WorkspaceID int
	Encryption  int8
	Token       string
	Target      string
	Location    struct {
		Region string
		Zone   string
		Compus string
	}
}

// Invoke 调用服务端接口
func (c *Client) Invoke(ctx context.Context, head *proto.RequestHeader,
	reqBody []byte, reqParam *ReqParam) (*proto.ResponseHeader, []byte, error) {
	ctx, msg := codec.NewMessage(ctx)
	defer codec.RecycleMessage(msg)

	timeout := types.GetMillisecond(int(head.Timeout))

	msg.WithCallRPCName("server.access.api/Query")
	msg.WithClientReqHead(head)
	msg.WithRequestID(head.RequestId)
	msg.WithTraceID(head.TraceId)
	msg.WithRequestTimeout(timeout)
	msg.WithFrameCodec(reqParam)

	// get options
	opts, err := c.getOptions(msg, reqParam.Target, timeout)
	if err != nil {
		return nil, nil, err
	}

	if opts.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	return invoke(ctx, reqBody, opts)
}

func (c *Client) getOptions(msg *codec.Msg, target string, timeout time.Duration) (*Options, error) {
	opts := defaultOptions.clone()

	// set service info options
	opts.SelectOptions.SourceServiceName = msg.CallerServiceName()
	opts.SelectOptions.SourceEnvName = msg.Env()

	opts.Target = target
	opts.EndPoint = ""
	opts.Timeout = timeout

	if err := opts.parseTarget(); err != nil {
		return nil, errs.New(errs.ErrClientRoute, err.Error())
	}

	return opts, nil
}

// 编码请求数据，并发起网络请求
func roundTrip(ctx context.Context, reqBody []byte, opts *Options) (*proto.ResponseHeader, []byte, error) {
	msg := codec.Message(ctx)

	// check if codec is empty, after updating msg
	if opts.Codec == nil {
		metrics.ClientCodecEmpty.Incr()
		return nil, nil, errs.New(errs.ErrClientEncode, "client: codec empty")
	}

	reqBuf, err := createReqBuf(msg, reqBody, opts)
	if err != nil {
		return nil, nil, err
	}

	// call backend service
	respBuf, err := opts.Transport.RoundTrip(ctx, reqBuf, opts)
	if err != nil {
		return nil, nil, err
	}

	respHeader, respBodyBuf, err := opts.Codec.Decode(msg, respBuf)
	if err != nil {
		return nil, nil, errs.New(errs.ErrClientDecode, "client codec Decode: "+err.Error())
	}

	if respHeader.Err != nil {
		return nil, nil, respHeader.Err.ToError()
	}

	reqHeader, err := getRequestHead(msg)
	if err != nil {
		return respHeader, nil, err
	}

	// 请求返回 request_id 不一致，则返回异常
	if respHeader.RequestId != reqHeader.RequestId {
		return respHeader, nil, errs.Newf(errs.ErrRequestIDNotMatch,
			"response request_id %d different from request request_id %d", respHeader.RequestId, reqHeader.RequestId)
	}

	// 请求返回 query_mode 不一致，会导致数据解析异常，说明 Exec、PExec、CompExec 用法有误
	if respHeader.QueryMode != reqHeader.QueryMode {
		return respHeader, nil, errs.Newf(errs.ErrQueryModeNotMatch,
			"response query mode %d different from request query mode %d", respHeader.QueryMode, reqHeader.QueryMode)
	}

	if msg.ClientRespError() != nil {
		return respHeader, nil, msg.ClientRespError()
	}

	return respHeader, respBodyBuf, nil
}

func createReqBuf(msg *codec.Msg, reqBody []byte, opts *Options) ([]byte, error) {
	reqBuf, err := opts.Codec.Encode(msg, reqBody)
	if err != nil {
		return nil, errs.New(errs.ErrClientEncode, "client codec Encode: "+err.Error())
	}

	return reqBuf, nil
}

func invoke(ctx context.Context, reqBody []byte, opts *Options) (*proto.ResponseHeader, []byte, error) {
	msg := codec.Message(ctx)

	// select a node of the backend service
	node, err := selectNode(ctx, opts)
	if err != nil {
		return nil, nil, err
	}

	resolveRemoteAddr(msg, node.Network, node.Address)

	// start to process the next filter and report
	begin := time.Now()
	respHeader, result, err := roundTrip(ctx, reqBody, opts)
	cost := time.Since(begin)

	if e, ok := err.(*errs.Error); ok &&
		e.Type == errs.ETypeSystem && (e.Code == errs.ErrClientConnect ||
		e.Code == errs.ErrClientTimeout || e.Code == errs.ErrClientNet) {
		e.Msg = fmt.Sprintf("%s, cost:%s", e.Msg, cost)
		opts.Selector.Report(node, cost, err)
	} else {
		opts.Selector.Report(node, cost, err)
	}

	// back pass the node info
	if addr := msg.RemoteAddr(); addr != nil {
		opts.Node.set(node, addr.String(), cost)
	} else {
		opts.Node.set(node, node.Address, cost)
	}

	return respHeader, result, err
}

// selects a backend node by selector related options and sets the msg.
func selectNode(ctx context.Context, opts *Options) (*naming.Node, error) {
	opts.SelectOptions.Ctx = ctx

	node, err := getNode(opts)
	if err != nil {
		metrics.SelectNodeFail.Incr()
		return nil, err
	}

	// selector might block for a while, need to check if ctx is still available
	if ctx.Err() == context.Canceled {
		return nil, errs.New(errs.ErrClientCanceled, "selector canceled after Select: "+ctx.Err().Error())
	}

	if ctx.Err() == context.DeadlineExceeded {
		return nil, errs.New(errs.ErrClientTimeout, "selector timeout after Select: "+ctx.Err().Error())
	}

	opts.LoadNodeConfig(node)

	return node, nil
}

// select node
func getNode(opts *Options) (*naming.Node, error) {
	node, err := opts.Selector.Select(opts.EndPoint, &opts.SelectOptions)
	if err != nil {
		return nil, errs.New(errs.ErrClientRoute, "client Select: "+err.Error())
	}

	if node.Address == "" {
		return nil, errs.New(errs.ErrClientRoute, fmt.Sprintf("client Select: node address empty:%+v", node))
	}
	return node, nil
}

func resolveRemoteAddr(msg *codec.Msg, network string, address string) {
	if msg.RemoteAddr() != nil {
		return
	}

	switch network {
	case "tcp", "tcp4", "tcp6":
		// 检查地址是否可以解析为 ip
		host, _, err := net.SplitHostPort(address)
		if err != nil || net.ParseIP(host) == nil {
			return
		}
	}

	var addr net.Addr
	switch network {
	case "tcp", "tcp4", "tcp6":
		addr, _ = net.ResolveTCPAddr(network, address)
	default:
		addr, _ = net.ResolveTCPAddr("tcp4", address)
	}

	msg.WithRemoteAddr(addr)
}
