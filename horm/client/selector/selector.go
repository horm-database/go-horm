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

// Package selector determines how database chooses a backend node by service name. It contains service
// discovery, load balance and circuit breaker.
package selector

import (
	"context"
	"time"

	"github.com/horm-database/common/naming"
)

// Selector is the interface that defines the selector.
type Selector interface {
	Select(serviceName string, opts *Options) (*naming.Node, error) // gets a backend node by service name.
	Report(node *naming.Node, cost time.Duration, err error) error  // reports request status.
}

var (
	selectors = make(map[string]Selector)
)

// Register registers a named Selector, such as l5, cmlb and tseer.
func Register(name string, s Selector) {
	selectors[name] = s
}

// Get gets a named Selector.
func Get(name string) Selector {
	s := selectors[name]
	return s
}

// Options defines the call options.
type Options struct {
	// Ctx is the corresponding context to request.
	Ctx context.Context
	// Key is the hash key of stateful routing.
	Key string
	// Replicas is the replicas of a single node for stateful routing. It's optional, and used to
	// address hash ring.
	Replicas int
	// EnvKey is the environment key.
	EnvKey string
	// SourceServiceName is the caller service name.
	SourceServiceName string
	// SourceEnvName is the caller environment name.
	SourceEnvName string
	// SourceMetadata is the caller metadata used to match routing.
	SourceMetadata map[string]string
	// DestinationEnvName is the callee environment name which is used to get node in the specific
	// environment.
	DestinationEnvName string
	// DestinationMetadata is the callee metadata used to match routing.
	DestinationMetadata map[string]string
	// LoadBalanceType is the load balance type.
	LoadBalanceType string

	// EnvTransfer is the environment of upstream server.
	EnvTransfer string
}
