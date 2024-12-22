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

package selector

import (
	"errors"
	"strings"
	"time"

	"github.com/horm-database/common/naming"
	"github.com/horm-database/common/util"
)

func init() {
	Register("ip", NewIPSelector())  // ip://ip:port
	Register("dns", NewIPSelector()) // dns://domain:port
}

// ipSelector is a selector based on ip list.
type ipSelector struct {
	safeRand *util.SafeRand
}

// NewIPSelector creates a new ipSelector.
func NewIPSelector() *ipSelector {
	return &ipSelector{
		safeRand: util.NewSafeRand(time.Now().UnixNano()),
	}
}

// Select implements Selector.Select. ServiceName may have multiple IP, such as ip1:port1,ip2:port2.
func (s *ipSelector) Select(serviceName string, opts *Options) (node *naming.Node, err error) {
	if serviceName == "" {
		return nil, errors.New("serviceName empty")
	}

	addr, err := s.chooseOne(serviceName)
	if err != nil {
		return nil, err
	}
	return &naming.Node{ServiceName: serviceName, Address: addr}, nil
}

func (s *ipSelector) chooseOne(serviceName string) (string, error) {
	num := strings.Count(serviceName, ",") + 1
	if num == 1 {
		return serviceName, nil
	}

	var addr string
	r := s.safeRand.Intn(num)
	for i := 0; i <= r; i++ {
		j := strings.IndexByte(serviceName, ',')
		if j < 0 {
			addr = serviceName
			break
		}
		addr, serviceName = serviceName[:j], serviceName[j+1:]
	}
	return addr, nil
}

// Report reports nothing.
func (s *ipSelector) Report(*naming.Node, time.Duration, error) error {
	return nil
}
