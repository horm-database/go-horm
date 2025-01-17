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

package pool

import (
	"context"
	"errors"
	"io"
	"net"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/horm-database/common/metrics"
	"github.com/horm-database/go-horm/horm/codec"
)

var globalBuffer = make([]byte, 1)

// DefaultConnectionPool is the default connection pool
var DefaultConnectionPool = NewConnectionPool()

// connection pool error message.
var (
	ErrPoolLimit  = errors.New("connection pool limit")  // number of connections exceeds the limit error.
	ErrPoolClosed = errors.New("connection pool closed") // connection pool closed error.
	ErrConnClosed = errors.New("pool closed")            // connection closed.
	ErrNoDeadline = errors.New("dial no deadline")       // has no deadline set.
)

// HealthChecker idle connection health check function.
// The function supports quick check and comprehensive check.
// Quick check is called when an idle connection is obtained,
// and only checks whether the connection status is abnormal.
// The function returns true to indicate that the connection is available normally.
type HealthChecker func(pc *PoolConn, isFast bool) bool

// NewConnectionPool creates a connection pool.
func NewConnectionPool() *Pool {
	// Default value, tentative, need to debug to determine the specific value.
	opts := &Options{
		MaxIdle:     defaultMaxIdle,
		IdleTimeout: defaultIdleTimeout,
		DialTimeout: defaultDialTimeout,
	}

	return &Pool{
		opts:            opts,
		connectionPools: new(sync.Map),
	}
}

// Pool connection pool factory, maintains connection pools corresponding to all addresses,
// and connection pool option information.
type Pool struct {
	opts            *Options
	connectionPools *sync.Map
}

type dialFunc = func(ctx context.Context) (net.Conn, error)

func (p *Pool) getDialFunc(network string, address string) dialFunc {
	dialOpts := &DialOptions{
		Network: network,
		Address: address,
	}

	return func(ctx context.Context) (net.Conn, error) {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		d, ok := ctx.Deadline()
		if !ok {
			return nil, ErrNoDeadline
		}

		opts := *dialOpts
		opts.Timeout = time.Until(d)

		var localAddr net.Addr
		if opts.LocalAddr != "" {
			var err error
			localAddr, err = net.ResolveTCPAddr(opts.Network, opts.LocalAddr)
			if err != nil {
				return nil, err
			}
		}

		dialer := &net.Dialer{
			Timeout:   opts.Timeout,
			LocalAddr: localAddr,
		}

		return dialer.Dial(opts.Network, opts.Address)
	}
}

// GetConn is used to get the connection from the connection pool.
func (p *Pool) GetConn(ctx context.Context, network string, address string) (*PoolConn, error) {
	ctx, cancel := getDialCtx(ctx, p.opts.DialTimeout)
	if cancel != nil {
		defer cancel()
	}

	key := getNodeKey(network, address)
	if v, ok := p.connectionPools.Load(key); ok {
		return v.(*ConnectionPool).Get(ctx)
	}

	newPool := &ConnectionPool{
		Dial:            p.getDialFunc(network, address),
		MinIdle:         p.opts.MinIdle,
		MaxIdle:         p.opts.MaxIdle,
		MaxActive:       p.opts.MaxActive,
		Wait:            p.opts.Wait,
		MaxConnLifetime: p.opts.MaxConnLifetime,
		IdleTimeout:     p.opts.IdleTimeout,
		forceClosed:     p.opts.ForceClose,
	}

	newPool.checker = newPool.defaultChecker
	if p.opts.Checker != nil {
		newPool.checker = p.opts.Checker
	}

	// Avoid the problem of writing concurrently to the pool map during initialization.
	v, ok := p.connectionPools.LoadOrStore(key, newPool)
	if !ok {
		go newPool.checkRoutine(defaultCheckInterval)
		newPool.initialConnections(newPool.MinIdle)
		return newPool.Get(ctx)
	}
	return v.(*ConnectionPool).Get(ctx)
}

// ConnectionPool is the connection pool.
type ConnectionPool struct {
	Dial            func(context.Context) (net.Conn, error) // initialize the connection.
	MinIdle         int                                     // Minimum number of idle connections.
	MaxIdle         int                                     // Maximum number of idle connections, 0 means no limit.
	MaxActive       int                                     // Maximum number of active connections, 0 means no limit.
	IdleTimeout     time.Duration                           // idle connection timeout.
	MaxConnLifetime time.Duration                           // maximum lifetime of the connection.

	Wait        bool          // whether to wait when the maximum number of active connections is reached.
	mu          sync.Mutex    // control concurrent locks.
	checker     HealthChecker // idle connection health check function.
	closed      bool          // whether the connection pool has been closed.
	active      int           // current number of active connections.
	ch          chan struct{} // when Wait is true, used to limit the number of connections.
	once        sync.Once     // indicates whether ch has been initialized.
	idle        connList      // idle connection list.
	forceClosed bool          // force close the connection, suitable for streaming scenarios.
}

func (p *ConnectionPool) initialConnections(count int) {
	if count <= 0 {
		return
	}

	go func() {
		mu := sync.Mutex{}
		connections := make([]*PoolConn, 0, count)
		wg := sync.WaitGroup{}
		wg.Add(count)
		for i := 0; i < count; i++ {
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), defaultDialTimeout)
				defer cancel()
				conn, err := p.get(ctx, true)
				if err != nil {
					wg.Done()
					return
				}
				mu.Lock()
				connections = append(connections, conn)
				mu.Unlock()
				wg.Done()
			}()
		}
		wg.Wait()

		for _, conn := range connections {
			// puts connection back into the connection pool.
			conn.Close()
		}
	}()
}

// Get gets the connection from the connection pool.
func (p *ConnectionPool) Get(ctx context.Context) (pc *PoolConn, err error) {
	pc, err = p.get(ctx, false)
	if err != nil {
		metrics.ConnectionPoolGetConnectionErr.Incr()
		return nil, err
	}
	return pc, nil
}

// Close releases the connection.
func (p *ConnectionPool) Close() error {
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil
	}
	p.closed = true
	p.active -= p.idle.count
	pc := p.idle.head
	p.idle.count = 0
	p.idle.head, p.idle.tail = nil, nil
	if p.ch != nil {
		close(p.ch)
	}
	p.mu.Unlock()
	for ; pc != nil; pc = pc.next {
		pc.Conn.Close()
		pc.closed = true
	}
	return nil
}

// initializeCh After reaching the upper limit of the number of connections,
// if you need to block, you need to initialize p.ch for synchronization.
func (p *ConnectionPool) initializeCh() {
	p.once.Do(
		func() {
			p.ch = make(chan struct{}, p.MaxActive)
			if p.closed {
				close(p.ch)
			} else {
				for i := 0; i < p.MaxActive; i++ {
					p.ch <- struct{}{}
				}
			}
		},
	)
}

// get the connection from the connection pool.
func (p *ConnectionPool) get(ctx context.Context, forceNew bool) (*PoolConn, error) {
	if p.Wait && p.MaxActive > 0 {
		p.initializeCh()
		if ctx == nil {
			<-p.ch
		} else {
			select {
			case <-p.ch:
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	if !forceNew {
		// try to get an idle connection.
		if pc := p.getIdleConn(); pc != nil {
			return pc, nil
		}
	}

	// get new connection.
	return p.getNewConn(ctx)
}

func (p *ConnectionPool) getIdleConn() *PoolConn {
	p.mu.Lock()
	for p.idle.head != nil {
		pc := p.idle.head
		p.idle.popHead()
		p.mu.Unlock()
		if p.checker(pc, true) {
			return pc
		}
		pc.Conn.Close()
		pc.closed = true
		p.mu.Lock()
		p.active--
	}
	p.mu.Unlock()
	return nil
}

func (p *ConnectionPool) getNewConn(ctx context.Context) (*PoolConn, error) {
	// If the connection pool has been closed, return an error directly.
	p.mu.Lock()
	if p.closed {
		p.mu.Unlock()
		return nil, ErrPoolClosed
	}

	// Check if the connection pool limit is exceeded.
	if p.overLimit() {
		p.mu.Unlock()
		return nil, ErrPoolLimit
	}

	p.active++
	p.mu.Unlock()
	c, err := p.dial(ctx)
	if err != nil {
		c = nil
		p.mu.Lock()
		p.active--
		if p.ch != nil && !p.closed {
			p.ch <- struct{}{}
		}
		p.mu.Unlock()
		return nil, err
	}
	metrics.ConnectionPoolGetNewConnection.Incr()
	return p.newPoolConn(c), nil
}

func (p *ConnectionPool) newPoolConn(c net.Conn) *PoolConn {
	pc := &PoolConn{
		Conn:       c,
		created:    time.Now(),
		pool:       p,
		forceClose: p.forceClosed,
	}

	pc.fr = codec.NewFramer(codec.NewReader(pc))
	return pc
}

func (p *ConnectionPool) checkHealthOnce() {
	p.mu.Lock()
	n := p.idle.count
	for i := 0; i < n && p.idle.head != nil; i++ {
		pc := p.idle.head
		p.idle.popHead()
		p.mu.Unlock()
		if p.checker(pc, false) {
			p.mu.Lock()
			p.idle.pushTail(pc)
		} else {
			p.mu.Lock()
			pc.Conn.Close()
			pc.closed = true
			p.active--
		}
	}
	p.mu.Unlock()
}

func (p *ConnectionPool) checkRoutine(interval time.Duration) {
	for {
		time.Sleep(interval)
		p.mu.Lock()
		closed := p.closed
		p.mu.Unlock()
		if closed {
			return
		}
		p.checkHealthOnce()

		// Check if the minimum number of idle connections is met.
		p.checkMinIdle()
	}
}

func (p *ConnectionPool) checkMinIdle() {
	if p.MinIdle <= 0 {
		return
	}
	p.mu.Lock()
	idle := p.idle.count
	p.mu.Unlock()
	if idle < p.MinIdle {
		p.initialConnections(p.MinIdle - idle)
	}
}

// defaultChecker is the default idle connection check method,
// returning true means the connection is available normally.
func (p *ConnectionPool) defaultChecker(pc *PoolConn, isFast bool) bool {
	// Check whether the connection status is abnormal:
	// closed, network exception or sticky packet processing exception.
	if pc.isRemoteError() {
		return false
	}

	// Based on performance considerations, the quick check only does the RemoteErr check.
	if isFast {
		return true
	}

	// Check if the connection has exceeded the maximum idle time, if so close the connection.
	if p.IdleTimeout > 0 && pc.t.Add(p.IdleTimeout).Before(time.Now()) {
		metrics.ConnectionPoolIdleTimeout.Incr()
		return false
	}

	// Check if the connection is still alive.
	if p.MaxConnLifetime > 0 && pc.created.Add(p.MaxConnLifetime).Before(time.Now()) {
		metrics.ConnectionPoolLifetimeExceed.Incr()
		return false
	}

	return true
}

// overLimit The current number of active connections is greater than the maximum limit,
// if Wait = false, an error will be returned directly.
func (p *ConnectionPool) overLimit() bool {
	if !p.Wait && p.MaxActive > 0 && p.active >= p.MaxActive {
		metrics.ConnectionPoolOverLimit.Incr()
		return true
	}

	return false
}

// dial establishes a connection.
func (p *ConnectionPool) dial(ctx context.Context) (net.Conn, error) {
	if p.Dial != nil {
		return p.Dial(ctx)
	}
	return nil, errors.New("must pass Dial to pool")
}

// put tries to release the connection to the connection pool.
func (p *ConnectionPool) put(pc *PoolConn, forceClose bool) error {
	if pc.closed {
		return nil
	}
	p.mu.Lock()
	if pc.closed {
		p.mu.Unlock()
		return nil
	}
	if !p.closed && !forceClose {
		pc.t = time.Now()
		p.idle.pushHead(pc)
		if p.idle.count > p.MaxIdle {
			pc = p.idle.tail
			p.idle.popTail()
		} else {
			pc = nil
		}
	}
	if pc != nil {
		p.mu.Unlock()
		pc.closed = true
		pc.Conn.Close()
		p.mu.Lock()
		p.active--
	}
	if p.ch != nil && !p.closed {
		p.ch <- struct{}{}
	}
	p.mu.Unlock()
	return nil
}

// PoolConn is the connection in the connection pool.
type PoolConn struct {
	net.Conn
	fr         *codec.Framer
	t          time.Time
	created    time.Time
	next, prev *PoolConn
	pool       *ConnectionPool
	closed     bool
	forceClose bool
}

// ReadFrame reads the frame.
func (pc *PoolConn) ReadFrame() ([]byte, error) {
	if pc.closed {
		return nil, ErrConnClosed
	}
	if pc.fr == nil {
		pc.pool.put(pc, true)
		return nil, errors.New("framer not set")
	}
	data, err := pc.fr.ReadFrame()
	if err != nil {
		// ReadFrame failure may be socket Read interface timeout failure
		// or the unpacking fails, in both cases the connection should be closed.
		pc.pool.put(pc, true)
		return nil, err
	}

	return data, err
}

// isRemoteError tries to receive a byte to detect whether the peer has actively closed the connection.
// If the peer returns an io.EOF error, it is indicated that the peer has been closed.
// Idle connections should not read body, if the body is read, it means the upper layer's
// sticky packet processing is not done, the connection should also be discarded.
// return true if there is an error in the connection.
func (pc *PoolConn) isRemoteError() bool {
	err := checkConnErrUnblock(pc.Conn, globalBuffer)
	if err != nil {
		metrics.ConnectionPoolRemoteErr.Incr()
		return true
	}
	return false
}

// reset resets the connection state.
func (pc *PoolConn) reset() {
	if pc == nil {
		return
	}
	pc.Conn.SetDeadline(time.Time{})
}

// Write sends body on the connection.
func (pc *PoolConn) Write(b []byte) (int, error) {
	if pc.closed {
		return 0, ErrConnClosed
	}
	n, err := pc.Conn.Write(b)
	if err != nil {
		pc.pool.put(pc, true)
	}
	return n, err
}

// Read reads body on the connection.
func (pc *PoolConn) Read(b []byte) (int, error) {
	if pc.closed {
		return 0, ErrConnClosed
	}
	n, err := pc.Conn.Read(b)
	if err != nil {
		pc.pool.put(pc, true)
	}
	return n, err
}

// Close overrides the Close method of net.Conn and puts it back into the connection pool.
func (pc *PoolConn) Close() error {
	if pc.closed {
		return ErrConnClosed
	}
	pc.reset()
	return pc.pool.put(pc, pc.forceClose)
}

// GetRawConn gets raw connection in PoolConn.
func (pc *PoolConn) GetRawConn() net.Conn {
	return pc.Conn
}

// connList maintains idle connections and uses stacks to maintain connections.
//
// The stack method has an advantage over the queue. When the request volume is relatively small but the request
// distribution is still relatively uniform, the queue method will cause the occupied connection to be delayed.
type connList struct {
	count      int
	head, tail *PoolConn
}

func (l *connList) pushHead(pc *PoolConn) {
	pc.next = l.head
	pc.prev = nil
	if l.count == 0 {
		l.tail = pc
	} else {
		l.head.prev = pc
	}
	l.count++
	l.head = pc
}

func (l *connList) popHead() {
	pc := l.head
	l.count--
	if l.count == 0 {
		l.head, l.tail = nil, nil
	} else {
		pc.next.prev = nil
		l.head = pc.next
	}
	pc.next, pc.prev = nil, nil
}

func (l *connList) pushTail(pc *PoolConn) {
	pc.next = nil
	pc.prev = l.tail
	if l.count == 0 {
		l.head = pc
	} else {
		l.tail.next = pc
	}
	l.count++
	l.tail = pc
}

func (l *connList) popTail() {
	pc := l.tail
	l.count--
	if l.count == 0 {
		l.head, l.tail = nil, nil
	} else {
		pc.prev.next = nil
		l.tail = pc.prev
	}
	pc.next, pc.prev = nil, nil
}

func getNodeKey(network, address string) string {
	const underline = "_"
	var key strings.Builder
	key.Grow(len(network) + len(address) + 1)
	key.WriteString(network)
	key.WriteString(underline)
	key.WriteString(address)
	return key.String()
}

func checkConnErrUnblock(conn net.Conn, buf []byte) error {
	sysConn, ok := conn.(syscall.Conn)
	if !ok {
		return nil
	}
	rawConn, err := sysConn.SyscallConn()
	if err != nil {
		return err
	}

	var sysErr error
	var n int
	err = rawConn.Read(func(fd uintptr) bool {
		// Go sets the socket to non-blocking mode by default, and calling syscall can return directly.
		// Refer to the Go source code: sysSocket() function under src/net/sock_cloexec.go
		n, sysErr = syscall.Read(int(fd), buf)
		// Return true, the blocking and waiting encapsulated by
		// the net library will not be executed, and return directly.
		return true
	})

	if err != nil {
		return err
	}

	// connection is closed, return io.EOF.
	if n == 0 && sysErr == nil {
		metrics.ConnectionPoolRemoteEOF.Incr()
		return io.EOF
	}
	// Idle connections should not read body.
	if n > 0 {
		return errors.New("unexpected read from socket")
	}
	// Return to EAGAIN or EWOULDBLOCK if the idle connection is in normal state.
	if sysErr == syscall.EAGAIN || sysErr == syscall.EWOULDBLOCK {
		return nil
	}
	return sysErr
}
