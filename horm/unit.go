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
	"github.com/horm-database/common/proto"
)

func createUnits(q *Query) ([]*proto.Unit, error) {
	units := make([]*proto.Unit, 0)

	err := addUnit(&units, q)
	if err != nil {
		return nil, err
	}

	return units, nil
}

func addUnit(units *[]*proto.Unit, q *Query) error {
	if q.Error != nil {
		return q.Error
	}

	if q.Unit.Size < 0 {
		q.Unit.Size = 0
	}

	*units = append(*units, q.Unit)

	if q.sub != nil {
		q.Unit.Sub = make([]*proto.Unit, 0)
		err := addUnit(&q.Unit.Sub, q.sub)
		if err != nil {
			return err
		}
	}

	if q.trans != nil {
		q.Unit.Trans = make([]*proto.Unit, 0)
		err := addUnit(&q.Unit.Trans, q.trans)
		if err != nil {
			return err
		}
	}

	if q.next != nil {
		err := addUnit(units, q.next)
		if err != nil {
			return err
		}
	}

	return nil
}
