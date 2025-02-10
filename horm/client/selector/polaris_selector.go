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
	"fmt"
	"time"

	"github.com/horm-database/common/naming"

	"github.com/polarismesh/polaris-go"
)

func init() {
	Register("polaris", &polarisSelector{}) // polaris://rpc.workspace.api
}

// polarisSelector is a selector based on ip list.
type polarisSelector struct {
	consumer polaris.ConsumerAPI
}

// Select implements Selector.Select.
func (s *polarisSelector) Select(serviceName string, opts *Options) (node *naming.Node, err error) {
	if s.consumer == nil {
		s.consumer, err = polaris.NewConsumerAPI()
		if err != nil {
			return nil, err
		}
	}

	if serviceName == "" {
		return nil, errors.New("serviceName empty")
	}

	getOneRequest := &polaris.GetOneInstanceRequest{}
	getOneRequest.Namespace = "workspace"
	getOneRequest.Service = serviceName

	oneInstResp, err := s.consumer.GetOneInstance(getOneRequest)
	if err != nil {
		return nil, err
	}

	instance := oneInstResp.GetInstance()
	if instance == nil {
		return nil, errors.New("not find any instance from polaris server")
	}

	return &naming.Node{
		ServiceName: serviceName,
		Address:     fmt.Sprintf("%s:%d", instance.GetHost(), instance.GetPort()),
	}, nil
}

// Report reports nothing.
func (s *polarisSelector) Report(*naming.Node, time.Duration, error) error {
	return nil
}
