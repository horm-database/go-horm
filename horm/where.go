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
	"fmt"

	"github.com/horm-database/common/consts"
)

func (w Where) OrWhere(where Where, comment ...string) Where {
	key := consts.OR
	if len(comment) > 0 {
		key = fmt.Sprintf("%s #%s", consts.OR, comment[0])
	}

	if w == nil {
		return Where{
			key: where,
		}
	}

	w[key] = where
	return w
}

func (w Where) AndWhere(where Where, comment ...string) Where {
	key := consts.AND
	if len(comment) > 0 {
		key = fmt.Sprintf("%s #%s", consts.AND, comment[0])
	}

	if w == nil {
		return Where{
			key: where,
		}
	}

	w[key] = where
	return w
}

func (w Where) NotWhere(where Where, comment ...string) Where {
	key := consts.NOT
	if len(comment) > 0 {
		key = fmt.Sprintf("%s #%s", consts.NOT, comment[0])
	}

	if w == nil {
		return Where{
			key: where,
		}
	}

	w[key] = where
	return w
}
