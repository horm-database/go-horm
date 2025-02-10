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
	"github.com/horm-database/common/consts"
	"github.com/horm-database/common/errs"
	"github.com/horm-database/common/types"
)

// Eq equal mean where key1=value1 AND key2=value2 ...
// input must be key, value, key, value, key, value ...
func (s *Query) Eq(key string, value interface{}, kvs ...interface{}) *Query {
	if s.Unit.Where == nil {
		s.Unit.Where = map[string]interface{}{}
	}

	s.Unit.Where[key] = value

	if len(kvs) > 0 {
		if len(kvs)%2 != 0 {
			s.Error = errs.Newf(errs.ErrReqParamInvalid, "more input must be a key-value pair")
			return s
		}

		var ok bool

		for k, tmp := range kvs {
			if k%2 == 0 { // this is key
				key, ok = tmp.(string)
				if !ok {
					s.Error = errs.Newf(errs.ErrReqParamInvalid, "more input key must be string")
					return s
				}
			} else { // this is value
				s.Unit.Where[key] = tmp
			}
		}
	}

	if s.Unit.Size == -1 {
		s.Unit.Size = 0
	}
	return s
}

func (s *Query) Not(key string, value interface{}) *Query {
	if s.Unit.Where == nil {
		s.Unit.Where = map[string]interface{}{key + consts.OPNot: value}
	} else {
		s.Unit.Where[key+consts.OPNot] = value
	}
	return s
}

func (s *Query) Lt(key string, value interface{}) *Query {
	if s.Unit.Where == nil {
		s.Unit.Where = map[string]interface{}{key + consts.OPLt: value}
	} else {
		s.Unit.Where[key+consts.OPLt] = value
	}
	return s
}

func (s *Query) Gt(key string, value interface{}) *Query {
	if s.Unit.Where == nil {
		s.Unit.Where = map[string]interface{}{key + consts.OPGt: value}
	} else {
		s.Unit.Where[key+consts.OPGt] = value
	}
	return s
}

func (s *Query) Lte(key string, value interface{}) *Query {
	if s.Unit.Where == nil {
		s.Unit.Where = map[string]interface{}{key + consts.OPLte: value}
	} else {
		s.Unit.Where[key+consts.OPLte] = value
	}
	return s
}

func (s *Query) Gte(key string, value interface{}) *Query {
	if s.Unit.Where == nil {
		s.Unit.Where = map[string]interface{}{key + consts.OPGte: value}
	} else {
		s.Unit.Where[key+consts.OPGte] = value
	}
	return s
}

func (s *Query) Between(key string, start, end interface{}) *Query {
	if s.Unit.Where == nil {
		s.Unit.Where = map[string]interface{}{key + consts.OPBetween: []interface{}{start, end}}
	} else {
		s.Unit.Where[key+consts.OPBetween] = []interface{}{start, end}
	}
	return s
}

func (s *Query) NotBetween(key string, start, end interface{}) *Query {
	if s.Unit.Where == nil {
		s.Unit.Where = map[string]interface{}{key + consts.OPNotBetween: []interface{}{start, end}}
	} else {
		s.Unit.Where[key+consts.OPNotBetween] = []interface{}{start, end}
	}
	return s
}

func (s *Query) Like(key string, value interface{}) *Query {
	if s.Unit.Where == nil {
		s.Unit.Where = map[string]interface{}{key + consts.OPLike: value}
	} else {
		s.Unit.Where[key+consts.OPLike] = value
	}
	return s
}

func (s *Query) NotLike(key string, value interface{}) *Query {
	if s.Unit.Where == nil {
		s.Unit.Where = map[string]interface{}{key + consts.OPNotLike: value}
	} else {
		s.Unit.Where[key+consts.OPNotLike] = value
	}
	return s
}

func (s *Query) MatchPhrase(key string, value interface{}) *Query {
	if s.Unit.Where == nil {
		s.Unit.Where = map[string]interface{}{key + consts.OPMatchPhrase: value}
	} else {
		s.Unit.Where[key+consts.OPMatchPhrase] = value
	}
	return s
}

func (s *Query) NotMatchPhrase(key string, value interface{}) *Query {
	if s.Unit.Where == nil {
		s.Unit.Where = map[string]interface{}{key + consts.OPNotMatchPhrase: value}
	} else {
		s.Unit.Where[key+consts.OPNotMatchPhrase] = value
	}
	return s
}

func (s *Query) Match(key string, value interface{}) *Query {
	if s.Unit.Where == nil {
		s.Unit.Where = map[string]interface{}{key + consts.OPMatch: value}
	} else {
		s.Unit.Where[key+consts.OPMatch] = value
	}
	return s
}

func (s *Query) NotMatch(key string, value interface{}) *Query {
	if s.Unit.Where == nil {
		s.Unit.Where = map[string]interface{}{key + consts.OPNotMatch: value}
	} else {
		s.Unit.Where[key+consts.OPNotMatch] = value
	}
	return s
}

// UpdateKV 更新字段，快速更新键值对 key = value
func (s *Query) UpdateKV(key string, value interface{}, kvs ...interface{}) *Query {
	s.Op("update")

	if len(s.Unit.Data) == 0 {
		s.Unit.Data = map[string]interface{}{}
	}

	if len(s.Unit.DataType) == 0 {
		s.Unit.DataType = make(map[string]types.Type)
	}

	typ := consts.GetDataType(value)
	if typ != 0 {
		s.Unit.DataType[key] = typ
	}

	s.Unit.Data[key] = value

	if len(kvs)%2 != 0 {
		s.Error = errs.New(errs.ErrReqParamInvalid, "UpdateKV pairs params must be even number")
		return s
	}

	if len(kvs) > 0 {
		var isStr bool
		for i, v := range kvs {
			if i%2 == 0 {
				key, isStr = v.(string)
				if !isStr {
					s.Error = errs.New(errs.ErrReqParamInvalid, "UpdateKV pairs params the first must be string")
					return s
				}
			} else {
				typ = consts.GetDataType(v)
				if typ != 0 {
					s.Unit.DataType[key] = typ
				}

				s.Unit.Data[key] = v
			}
		}
	}

	return s
}
