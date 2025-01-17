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

// Package pool provides the connection pool.
package pool

import (
	"context"
	"time"
)

const (
	defaultDialTimeout   = 200 * time.Millisecond
	defaultIdleTimeout   = 50 * time.Second
	defaultMaxIdle       = 65536
	defaultCheckInterval = 3 * time.Second
)

// Options indicates pool configuration.
type Options struct {
	MinIdle         int           // 最小闲置连接数量
	MaxIdle         int           // 最大闲置连接数量，0 代表不做闲置，框架默认值 2048
	MaxActive       int           // 最大活跃连接数量，0 代表不做限制
	Wait            bool          // 活跃连接达到最大数量时，是否等待
	IdleTimeout     time.Duration // 空闲连接超时时间，默认值 50s
	MaxConnLifetime time.Duration // 连接的最大生命周期
	DialTimeout     time.Duration // 建立连接超时时间，默认值 200ms
	ForceClose      bool
	Checker         HealthChecker
}

func getDialCtx(ctx context.Context, dialTimeout time.Duration) (context.Context, context.CancelFunc) {
	// ctx 不为空，而且设置了超时时间，则返回 ctx
	if ctx != nil {
		_, ok := ctx.Deadline()
		if ok {
			return ctx, nil
		}
	}

	if dialTimeout == 0 {
		dialTimeout = defaultDialTimeout
	}

	return context.WithTimeout(context.Background(), dialTimeout)

}

// DialOptions request parameters.
type DialOptions struct {
	Network   string
	Address   string
	LocalAddr string
	Timeout   time.Duration
}
