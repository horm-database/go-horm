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
		return nil, errs.New(errs.RetClientConnectFail,
			fmt.Sprintf("transport: network %s not support", opts.Network))
	}
}

// tcpRoundTrip sends tcp request. It supports send, sendAndRcv, keepalive and multiplex.
func (c *transport) tcpRoundTrip(ctx context.Context, reqData []byte, opts *Options) ([]byte, error) {
	if opts.Pool == nil {
		return nil, errs.New(errs.RetClientConnectFail, "tcp transport: connection pool empty")
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
		return nil, errs.New(errs.RetClientCanceled, "tcp transport canceled before Write: "+ctx.Err().Error())
	}
	if ctx.Err() == context.DeadlineExceeded {
		return nil, errs.New(errs.RetClientTimeout, "tcp transport timeout before Write: "+ctx.Err().Error())
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
		return nil, errs.New(errs.RetClientCanceled, "transport canceled before tcp dial: "+ctx.Err().Error())
	}
	if ctx.Err() == context.DeadlineExceeded {
		return nil, errs.New(errs.RetClientTimeout, "transport timeout before tcp dial: "+ctx.Err().Error())
	}

	d, ok := ctx.Deadline()

	// connection pool.
	tcpConn, err := opts.Pool.GetConn(ctx, opts.Network, opts.Address)

	if err != nil {
		return nil, errs.New(errs.RetClientConnectFail, "tcp transport connection pool: "+err.Error())
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
				return errs.New(errs.RetClientTimeout, "tcp transport Write: "+err.Error())
			}
			return errs.New(errs.RetClientNetErr, "tcp transport Write: "+err.Error())
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
			return nil, errs.New(errs.RetClientTimeout, "tcp transport ReadFrame: "+err.Error())
		}
		return nil, errs.New(errs.RetClientReadFrameFail, "tcp transport ReadFrame: "+err.Error())
	}

	return rspData, nil
}
