package client

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/horm-database/common/codec"
	"github.com/horm-database/common/naming"
	"github.com/horm-database/go-horm/horm/client/pool"
	"github.com/horm-database/go-horm/horm/client/selector"
)

// Options are clientside options.
type Options struct {
	Timeout time.Duration // timeout

	// Target is address of backend service: protocol://endpoint, like: ip://ip:port, dns://api.bjfanyi.com:8080, polaris://rpc.workspace.query.api
	Target   string
	EndPoint string // same as service name if target is not set

	Selector      selector.Selector
	SelectOptions selector.Options

	Node *onceNode // for getting node info

	// transport info
	Transport *transport
	Address   string     // IP:Port. Note: address has been resolved from naming service.
	Network   string     // tcp/udp
	Pool      *pool.Pool // client connection pool
	Msg       *codec.Msg

	Codec *clientCodec
}

var (
	defaultOptions *Options // Key: callee
)

func init() {
	defaultOptions = NewOptions()
	defaultOptions.Network = "tcp"
	defaultOptions.Codec = defaultCodec
}

// NewOptions creates a new Options
func NewOptions() *Options {
	return &Options{
		Transport: DefaultClientTransport,
		Selector:  selector.NewIPSelector(),
	}
}

func (opts *Options) clone() *Options {
	if opts == nil {
		return NewOptions()
	}
	o := *opts
	return &o
}

func (opts *Options) parseTarget() error {
	if opts.Target == "" {
		return nil
	}

	// Target should be like: selector://endpoint
	substr := "://"
	index := strings.Index(opts.Target, substr)
	if index == -1 {
		return fmt.Errorf("client: target %s schema invalid, format must be protocol://endpoint", opts.Target)
	}
	opts.Selector = selector.Get(opts.Target[:index])
	if opts.Selector == nil {
		return fmt.Errorf("client: selector %s not exist", opts.Target[:index])
	}
	opts.EndPoint = opts.Target[index+len(substr):]
	if opts.EndPoint == "" {
		return fmt.Errorf("client: target %s endpoint empty, format must be selector://endpoint", opts.Target)
	}

	return nil
}

// LoadNodeConfig loads node config from config center.
func (opts *Options) LoadNodeConfig(node *naming.Node) {
	opts.Address = node.Address
	opts.Codec = defaultCodec

	if node.Network != "" {
		opts.Network = node.Network
	}
}

type onceNode struct {
	*naming.Node
	once sync.Once
}

func (n *onceNode) set(node *naming.Node, address string, cost time.Duration) {
	if n == nil {
		return
	}
	n.once.Do(func() {
		*n.Node = *node
		n.Node.Address = address
		n.Node.CostTime = cost
	})
}
