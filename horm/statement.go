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
	"strconv"
	"strings"

	"github.com/horm-database/common/consts"
	"github.com/horm-database/common/errs"
	"github.com/horm-database/common/types"
	"github.com/horm-database/common/util"
	"github.com/horm-database/go-horm/horm/codec"
)

// Where 查询条件
type Where map[string]interface{}

// Map 更新操作 map 数据声明
type Map map[string]interface{}

// IN select * from table where column in (1, 2, 3)
type IN []interface{}

// AND 关联词
type AND map[string]interface{}

// OR 关联词
type OR map[string]interface{}

// Using sql Join using 结构声明
type Using []string

// On sql join On 结构声明
type On map[string]string

// Op 设置操作
func (s *Query) Op(op string) *Query {
	s.Unit.Op = strings.ToLower(op)
	return s
}

// Name 设置执行语句名称
func (s *Query) Name(name string) *Query {
	s.Unit.Name = name
	unitName, alias := util.Alias(name)
	if unitName == "" {
		s.Error = errs.Newf(errs.ErrReqUnitNameEmpty, "unit name is empty")
	}

	if alias != "" {
		s.Key = alias
	} else {
		s.Key = unitName
	}

	return s
}

// Shard 分片、分表
func (s *Query) Shard(shard ...string) *Query {
	s.Unit.Shard = shard
	return s
}

// Column 列
func (s *Query) Column(columns ...string) *Query {
	s.Unit.Column = columns
	return s
}

// Where 查询条件
func (s *Query) Where(where Where) *Query {
	s.Unit.Where = where
	return s
}

// Insert （批量）插入数据，参数可以是 struct / []struct / Map / []Map
func (s *Query) Insert(data interface{}) *Query {
	s.Op("insert")

	if data == nil {
		return s
	}

	mapData, err := s.GetCoder().Encode(codec.EncodeTypeInsertData, data)
	if err != nil {
		s.Error = err
		return s
	}

	return s.setMap(mapData)
}

// Replace （批量）替换数据，参数可以是 struct / []struct / Map / []Map
func (s *Query) Replace(data interface{}) *Query {
	s.Op("replace")

	if data == nil {
		return s
	}

	mapData, err := s.GetCoder().Encode(codec.EncodeTypeReplaceData, data)
	if err != nil {
		s.Error = err
		return s
	}

	return s.setMap(mapData)

}

// Update 更新数据，参数可以是 struct / Map
func (s *Query) Update(data interface{}, where ...Where) *Query {
	s.Op("update")

	if len(where) > 0 {
		s.Where(where[0])
	}

	m, err := s.GetCoder().Encode(codec.EncodeTypeUpdateData, data)
	if err != nil {
		s.Error = err
		return s
	}

	if m == nil {
		s.Error = errs.Newf(errs.ErrReqParamInvalid, "%s data is nil", s.Unit.Op)
	}

	switch v := m.(type) {
	case Map:
		s.Unit.Data = v
	case map[string]interface{}:
		s.Unit.Data = v
	default:
		s.Error = errs.New(errs.ErrReqParamInvalid, "update data`s type must be map/struct")
		return s
	}

	s.setDataType(s.Unit.Data)

	return s
}

// Delete 根据条件删除
func (s *Query) Delete(where ...Where) *Query {
	s.Op("delete")
	if len(where) > 0 {
		s.Where(where[0])
	}
	return s
}

// Find 查询满足条件的一条数据
func (s *Query) Find(where ...Where) *Query {
	s.Op("find")

	if len(where) != 0 {
		s.Where(where[0])
	}

	return s
}

// FindAll 查询满足条件的所有数据
func (s *Query) FindAll(where ...Where) *Query {
	s.Op("find_all")

	if len(where) != 0 {
		s.Where(where[0])
	}

	//默认取 100 条数据
	if s.Unit.Size == -1 {
		s.Unit.Size = 100
	}

	return s
}

// FindBy find where key1=value1 AND key2=value2 ...
// input must be key, value, key, value, key, value ...
func (s *Query) FindBy(key string, value interface{}, kvs ...interface{}) *Query {
	s.Find().Eq(key, value, kvs...)
	return s
}

// FindAllBy find_all where key1=value1 AND key2=value2 ...
// input must be key, value, key, value, key, value ...
func (s *Query) FindAllBy(key string, value interface{}, kvs ...interface{}) *Query {
	s.FindAll().Eq(key, value, kvs...)
	return s
}

// DeleteBy delete where key1=value1 AND key2=value2 ...
// input must be key, value, key, value, key, value ...
func (s *Query) DeleteBy(key string, value interface{}, kvs ...interface{}) *Query {
	s.Delete().Eq(key, value, kvs...)
	return s
}

// Page 分页
func (s *Query) Page(page, pageSize int) *Query {
	s.Unit.Page = page
	s.Unit.Size = pageSize
	return s
}

// Limit 排序
func (s *Query) Limit(limit int, offset ...uint64) *Query {
	s.Unit.Size = limit

	if len(offset) > 0 {
		s.Unit.From = offset[0]
	}

	return s
}

// Order 排序, 首字母 + 表示升序，- 表示降序
func (s *Query) Order(orders ...string) *Query {
	if s.Unit.Order == nil {
		s.Unit.Order = []string{}
	}

	s.Unit.Order = append(s.Unit.Order, orders...)
	return s
}

// Group 分组
func (s *Query) Group(group ...string) *Query {
	s.Unit.Group = group
	return s
}

// Having 分组条件
func (s *Query) Having(having Where) *Query {
	s.Unit.Having = having
	return s
}

// SetKey 给 key 赋值
func (s *Query) SetKey(key string) *Query {
	s.Unit.Key = key
	return s
}

// SetField 给 field 赋值
func (s *Query) SetField(field string) *Query {
	s.Unit.Field = field
	return s
}

func (s *Query) SetVal(val interface{}) *Query {
	if len(s.Unit.DataType) == 0 {
		s.Unit.DataType = make(map[string]types.Type)
	}

	if val == nil {
		s.Unit.Val = val
	}

	v, err := s.GetCoder().Encode(codec.EncodeTypeRedisVal, val)
	if err != nil {
		s.Error = err
		return s
	}

	switch vv := v.(type) {
	case types.Map:
		s.Unit.Data = vv
		s.setDataType(vv)
	case map[string]interface{}:
		s.Unit.Data = vv
		s.setDataType(vv)
	case []types.Map:
		s.Unit.Datas = make([]map[string]interface{}, len(vv))
		for k, iv := range vv {
			s.Unit.Datas[k] = iv
		}

		if len(vv) > 0 {
			s.setDataType(vv[0])
		}
	case []map[string]interface{}:
		s.Unit.Datas = vv
		if len(vv) > 0 {
			s.setDataType(vv[0])
		}
	default:
		s.Unit.Val = v
	}

	return s
}

// SetParam 与数据库相关的请求参数，例如 redis 的 WITHSCORES， elastic 的 collapse、runtime_mappings、track_total_hits 等等。
func (s *Query) SetParam(key string, value interface{}) *Query {
	if s.Unit.Params == nil {
		s.Unit.Params = map[string]interface{}{key: value}
	} else {
		s.Unit.Params[key] = value
	}

	return s
}

// Bytes 字节码
func (s *Query) Bytes(bs []byte) *Query {
	s.Unit.Bytes = bs
	return s
}

// Source 直接输入查询语句查询
func (s *Query) Source(q string, args ...interface{}) *Query {
	s.Op("find_all")
	s.Unit.Query = q
	s.Unit.Args = args

	s.Unit.DataType = map[string]types.Type{}

	for k, v := range args {
		typ := consts.GetDataType(v)
		if typ != 0 {
			s.Unit.DataType[strconv.Itoa(k)] = typ
		}
	}

	return s
}

// Extend 扩展信息会被传入到每个插件。用于自定义功能
func (s *Query) Extend(key string, value interface{}) *Query {
	if s.Unit.Extend == nil {
		s.Unit.Extend = map[string]interface{}{key: value}
	} else {
		s.Unit.Extend[key] = value
	}

	return s
}

// Create 创建表
func (s *Query) Create(name, shard string, ifNotExists ...bool) *Query {
	if len(ifNotExists) > 0 && ifNotExists[0] {
		s.SetParam("if_not_exists", true)
	}

	s.Op("create")
	s.Name(name)
	s.Shard(shard)
	return s
}

func (s *Query) setMap(data interface{}) *Query {
	if data == nil {
		s.Error = errs.Newf(errs.ErrReqParamInvalid, "%s data is nil", s.Unit.Op)
	}

	switch v := data.(type) {
	case Map:
		s.Unit.Data = v
	case map[string]interface{}:
		s.Unit.Data = v
	case []Map:
		if len(v) == 0 {
			s.Error = errs.Newf(errs.ErrReqParamInvalid, "%s data is empty", s.Unit.Op)
		}

		s.Unit.Datas = make([]map[string]interface{}, len(v))
		for k, tmp := range v {
			s.Unit.Datas[k] = tmp
		}
	case []map[string]interface{}:
		if len(v) == 0 {
			s.Error = errs.Newf(errs.ErrReqParamInvalid, "%s data is empty", s.Unit.Op)
		}

		s.Unit.Datas = make([]map[string]interface{}, len(v))
		for k, tmp := range v {
			s.Unit.Datas[k] = tmp
		}
	default:
		s.Error = errs.Newf(errs.ErrReqParamInvalid, "%s data`s type must be struct/[]struct/map/[]map", s.Unit.Op)
		return s
	}

	if len(s.Unit.Datas) > 0 {
		s.setDataType(s.Unit.Datas[0])
	} else {
		s.setDataType(s.Unit.Data)
	}

	return s
}

func (s *Query) setDataType(data map[string]interface{}) *Query {
	if data == nil {
		return s
	}

	s.Unit.DataType = map[string]types.Type{}
	for key, value := range data {
		typ := consts.GetDataType(value)
		if typ != 0 {
			s.Unit.DataType[key] = typ
		}
	}
	return s
}
