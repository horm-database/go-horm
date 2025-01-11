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
	"github.com/horm-database/common/errs"
	"github.com/horm-database/common/proto"
)

// Type elastic search 版本 v7 以前有 type， v7之后 type 统一为 _doc
func (s *Query) Type(typ string) *Query {
	s.Unit.Type = typ
	return s
}

// ID 主键 _id 查询
func (s *Query) ID(value interface{}) *Query {
	return s.Eq("_id", value)
}

// Scroll 查询，size 为每次 scroll 大小，where 为 scroll 条件。
func (s *Query) Scroll(scroll string, size int, where ...Where) *Query {
	s.Op("scroll")

	if len(where) != 0 {
		s.Where(where[0])
	}

	if size <= 0 {
		s.Error = errs.New(errs.ErrReqParamInvalid, "scroll size can`t be zero")
		return s
	}

	s.Limit(size)

	if s.Unit.Scroll == nil {
		s.Unit.Scroll = new(proto.Scroll)
	}

	s.Unit.Scroll.Info = scroll
	return s
}

// ScrollByID 根据 scrollID 滚动查询。
func (s *Query) ScrollByID(id string) *Query {
	s.Op("scroll")

	if s.Unit.Scroll == nil {
		s.Unit.Scroll = new(proto.Scroll)
	}

	s.Unit.Scroll.ID = id
	return s
}

// Refresh 更新数据立即刷新
func (s *Query) Refresh() *Query {
	s.SetParam("refresh", true)
	return s
}

// Routing 路由
func (s *Query) Routing(routing string) *Query {
	s.SetParam("routing", routing)
	return s
}

// HighLight 返回高亮
func (s *Query) HighLight(field string, preTag, postTag string, replace ...bool) *Query {
	highLight := map[string]interface{}{}
	highLight["field"] = field
	highLight["pre_tag"] = preTag
	highLight["post_tag"] = postTag

	if len(replace) > 0 && replace[0] {
		highLight["replace"] = true
	}

	highLights := []map[string]interface{}{}
	if len(s.Unit.Params) > 0 {
		v, ok := s.Unit.Params["highlights"].([]map[string]interface{})
		if ok {
			highLights = v
		}
	}

	highLights = append(highLights, highLight)
	s.SetParam("highlights", highLights)
	return s
}

// Collapse collapse search results
func (s *Query) Collapse(field string) *Query {
	collapse := map[string]interface{}{}
	collapse["field"] = field
	s.SetParam("collapse", collapse)
	return s
}
