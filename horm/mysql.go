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
	"reflect"

	"github.com/horm-database/common/errs"
	"github.com/horm-database/common/proto/sql"
	"github.com/horm-database/common/types"
)

// Join 表 join 语句
func (s *Query) Join(table string, relation ...interface{}) *Query {
	return s.realJoin(table, "", relation...)
}

// LeftJoin Left JOIN 表 left join 语句
func (s *Query) LeftJoin(table string, relation ...interface{}) *Query {
	return s.realJoin(table, "LEFT", relation...)
}

// RightJoin Right JOIN 表 right join 语句
func (s *Query) RightJoin(table string, relation ...interface{}) *Query {
	return s.realJoin(table, "RIGHT", relation...)
}

// InnerJoin Inner JOIN 表 inner join 语句
func (s *Query) InnerJoin(table string, relation ...interface{}) *Query {
	return s.realJoin(table, "INNER", relation...)
}

// FullJoin Full JOIN 表 full join 语句
func (s *Query) FullJoin(table string, relation ...interface{}) *Query {
	return s.realJoin(table, "FULL", relation...)
}

func (s *Query) realJoin(table string, joinType string, relation ...interface{}) *Query {
	join := sql.Join{
		Type:  joinType,
		Table: table,
	}

	if len(relation) > 0 {
		rela := relation[0]
		v := reflect.ValueOf(rela)
		if v.Kind() == reflect.String {
			join.Using = []string{rela.(string)}
		} else if types.IsArray(v) {
			switch relations := rela.(type) {
			case Using:
				join.Using = relations
			case []string:
				join.Using = relations
			default:
				s.Error = errs.Newf(errs.ErrReqParamInvalid, "the third param`s type must be []string if it is array")
				return s
			}
		} else if v.Kind() == reflect.Map {
			switch relations := rela.(type) {
			case On:
				join.On = relations
			case map[string]string:
				join.On = relations
			default:
				s.Error = errs.Newf(errs.ErrReqParamInvalid,
					"the third param`s type must be map[string]string if it is map")
				return s
			}
		}
	}

	s.Unit.Join = append(s.Unit.Join, &join)

	return s
}
