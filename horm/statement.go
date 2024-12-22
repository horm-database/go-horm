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
	"strconv"
	"strings"

	"github.com/horm-database/common/consts"
	"github.com/horm-database/common/errs"
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
	return s.setData(codec.EncodeTypeInsertData, data)
}

// Replace （批量）替换数据，参数可以是 struct / []struct / Map / []Map
func (s *Query) Replace(data interface{}) *Query {
	return s.setData(codec.EncodeTypeReplaceData, data)
}

// Update 更新数据，参数可以是 struct / Map
func (s *Query) Update(data interface{}, where ...Where) *Query {
	s.Op("update")

	if len(where) > 0 {
		s.Where(where[0])
	}

	switch v := data.(type) {
	case Map:
		return s.updateMap(v)
	case map[string]interface{}:
		return s.setMap(v)
	}

	v := reflect.Indirect(reflect.ValueOf(data))

	if v.Kind() != reflect.Struct {
		s.Error = errs.New(errs.ErrReqParamInvalid, "Update first param`s type must be struct/Map")
		return s
	}

	d, _ := s.GetCoder().Encode(codec.EncodeTypeUpdateData, data)
	s.Unit.Data = d.(map[string]interface{})
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

	s.Unit.DataType = map[string]consts.DataType{}

	for k, v := range args {
		typ := consts.GetDataType(v)
		if typ != consts.DataTypeOther {
			s.Unit.DataType[strconv.Itoa(k)] = typ
		}
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

// SetParam 与数据库相关的请求参数，例如 redis 的 WITHSCORES， elastic 的 collapse、runtime_mappings、track_total_hits 等等。
func (s *Query) SetParam(key string, value interface{}) *Query {
	if s.Unit.Params == nil {
		s.Unit.Params = map[string]interface{}{key: value}
	} else {
		s.Unit.Params[key] = value
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

func (s *Query) setMap(data Map) *Query {
	if len(data) == 0 {
		return s
	}

	s.Unit.Data = data
	s.setDataType(data)
	return s
}

func (s *Query) setMaps(datas []Map) *Query {
	if len(datas) == 0 {
		return s
	}

	for _, data := range datas {
		s.Unit.Datas = append(s.Unit.Datas, data)
	}

	s.setDataType(datas[0])
	return s
}

func (s *Query) updateMap(data Map) *Query {
	if len(data) == 0 {
		return s
	}

	s.Unit.Data = data
	s.setDataType(data)
	return s
}

func (s *Query) setData(typ codec.EncodeType, data interface{}) *Query {
	var op string
	if typ == codec.EncodeTypeInsertData {
		op = "insert"
	} else {
		op = "replace"
	}

	s.Op(op)

	if data == nil {
		return s
	}

	switch v := data.(type) {
	case Map:
		return s.setMap(v)
	case []Map:
		return s.setMaps(v)
	case map[string]interface{}:
		return s.setMap(v)
	case []map[string]interface{}:
		mapArr := make([]Map, len(v))
		for i, val := range v {
			mapArr[i] = val
		}
		return s.setMaps(mapArr)
	}

	// struct 转 map
	v := reflect.Indirect(reflect.ValueOf(data))
	switch v.Kind() {
	case reflect.Struct, reflect.Array, reflect.Slice:
		d, _ := s.GetCoder().Encode(typ, data)
		if d == nil {
			return s
		}

		switch v := d.(type) {
		case map[string]interface{}:
			s.Unit.Data = v
			return s.setDataType(s.Unit.Data)
		case []map[string]interface{}:
			s.Unit.Datas = v
			return s.setDataType(s.Unit.Datas[0])
		}
	}

	s.Error = errs.Newf(errs.ErrReqParamInvalid, "%s param`s type must be struct/[]struct/Map/[]Map", op)
	return s
}

func (s *Query) setDataType(data map[string]interface{}) *Query {
	s.Unit.DataType = map[string]consts.DataType{}

	for key, value := range data {
		typ := consts.GetDataType(value)
		if typ != consts.DataTypeOther {
			s.Unit.DataType[key] = typ
		}
	}
	return s
}

// struct 转 map
func (s *Query) structToMap(data interface{}) (ret interface{}) {
	if data == nil {
		return nil
	}

	switch v := data.(type) {
	case Map, []Map, map[string]interface{}, []map[string]interface{}:
		return v
	}

	v := reflect.Indirect(reflect.ValueOf(data))
	switch v.Kind() {
	case reflect.Struct, reflect.Array, reflect.Slice:
		d, err := s.GetCoder().Encode(codec.EncodeTypeInsertData, data)
		if err == nil {
			return d
		}
	}

	return data
}
