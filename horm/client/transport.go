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

package client

import (
	"context"
	"fmt"
	"net"

	"github.com/horm-database/common/codec"
	"github.com/horm-database/common/errs"
	"github.com/horm-database/go-horm/horm/client/pool"
)

// DefaultClientTransport is the default client transport.
var DefaultClientTransport = &transport{}

// transport is the implementation details of transport, such as tcp/udp roundTrip.
type transport struct{}

// RoundTrip sends client requests.
func (c *transport) RoundTrip(ctx context.Context,
	req []byte, opts *Options) (rsp []byte, err error) {
	opts.Pool = pool.DefaultConnectionPool

	switch opts.Network {
	case "tcp", "tcp4", "tcp6", "unix":
		return c.tcpRoundTrip(ctx, req, opts)
	default:
		return nil, errs.New(errs.ErrClientConnect,
			fmt.Sprintf("transport: network %s not support", opts.Network))
	}
}

// tcpRoundTrip sends tcp request. It supports send, sendAndRcv, keepalive and multiplex.
func (c *transport) tcpRoundTrip(ctx context.Context, reqData []byte, opts *Options) ([]byte, error) {
	if opts.Pool == nil {
		return nil, errs.New(errs.ErrClientConnect, "tcp transport: connection pool empty")
	}

	tcpConn, err := c.dialTCP(ctx, opts)
	if err != nil {
		return nil, err
	}

	// TCP connection is exclusively multiplexed. Close determines whether connection should be put
	// back into the connection pool to be reused.
	defer tcpConn.Close()

	msg := codec.Message(ctx)
	msg.WithRemoteAddr(tcpConn.RemoteAddr())
	msg.WithLocalAddr(tcpConn.LocalAddr())

	if ctx.Err() == context.Canceled {
		return nil, errs.New(errs.ErrClientCanceled, "tcp transport canceled before Write: "+ctx.Err().Error())
	}
	if ctx.Err() == context.DeadlineExceeded {
		return nil, errs.New(errs.ErrClientTimeout, "tcp transport timeout before Write: "+ctx.Err().Error())
	}

	if err := c.tcpWriteFrame(tcpConn, reqData); err != nil {
		return nil, err
	}
	return c.tcpReadFrame(tcpConn)
}

// dialTCP establishes a TCP connection.
func (c *transport) dialTCP(ctx context.Context, opts *Options) (*pool.PoolConn, error) {
	// If ctx has canceled or timeout, just return.
	if ctx.Err() == context.Canceled {
		return nil, errs.New(errs.ErrClientCanceled, "transport canceled before tcp dial: "+ctx.Err().Error())
	}
	if ctx.Err() == context.DeadlineExceeded {
		return nil, errs.New(errs.ErrClientTimeout, "transport timeout before tcp dial: "+ctx.Err().Error())
	}

	d, ok := ctx.Deadline()

	// connection pool.
	tcpConn, err := opts.Pool.GetConn(ctx, opts.Network, opts.Address)

	if err != nil {
		return nil, errs.New(errs.ErrClientConnect, "tcp transport connection pool: "+err.Error())
	}

	if ok {
		_ = tcpConn.SetDeadline(d)
	}

	return tcpConn, nil
}

// tcpWriteReqData writes the tcp frame.
func (c *transport) tcpWriteFrame(conn net.Conn, reqData []byte) error {
	// Send package in a loop.
	sentNum := 0
	num := 0
	var err error
	for sentNum < len(reqData) {
		num, err = conn.Write(reqData[sentNum:])
		if err != nil {
			if e, ok := err.(net.Error); ok && e.Timeout() {
				return errs.New(errs.ErrClientTimeout, "tcp transport Write: "+err.Error())
			}
			return errs.New(errs.ErrClientNet, "tcp transport Write: "+err.Error())
		}
		sentNum += num
	}
	return nil
}

// tcpReadFrame reads the tcp frame.
func (c *transport) tcpReadFrame(conn *pool.PoolConn) ([]byte, error) {
	rspData, err := conn.ReadFrame()
	if err != nil {
		if e, ok := err.(net.Error); ok && e.Timeout() {
			return nil, errs.New(errs.ErrClientTimeout, "tcp transport ReadFrame: "+err.Error())
		}
		return nil, errs.New(errs.ErrClientReadFrame, "tcp transport ReadFrame: "+err.Error())
	}

	return rspData, nil
}
