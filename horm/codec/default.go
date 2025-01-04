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

package codec

import (
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/horm-database/common/consts"
	"github.com/horm-database/common/proto"
	"github.com/horm-database/common/snowflake"
	"github.com/horm-database/common/structs"
	"github.com/horm-database/common/types"
)

type defaultCodec struct {
	m             Marshal
	um            Unmarshal
	tag           string // the tag in the structure, use for encode and decode, if the tag does not exist, use the name of the structure as the field
	omitEmpty     bool   // if open omit empty , default is true
	map2structure *Map2Structure
	l             *time.Location
}

type EncodeType int8

const (
	EncodeTypeHmSET       EncodeType = 1
	EncodeTypeRedisArg    EncodeType = 2
	EncodeTypeInsertData  EncodeType = 3
	EncodeTypeReplaceData EncodeType = 4
	EncodeTypeUpdateData  EncodeType = 5
)

// SetMarshal set marshal/unmarshal
func (dc *defaultCodec) SetMarshal(m Marshal, um Unmarshal) {
	dc.m = m
	dc.um = um
}

// GetMarshal get marshal
func (dc *defaultCodec) GetMarshal() Marshal {
	return dc.m
}

// GetUnmarshal get unmarshal
func (dc *defaultCodec) GetUnmarshal() Unmarshal {
	return dc.um
}

// SetTag set tag
func (dc *defaultCodec) SetTag(tag string) {
	dc.tag = tag
}

// GetTag get tag
func (dc *defaultCodec) GetTag() string {
	if dc.tag == "" {
		return DefaultTag
	}

	return dc.tag
}

func (dc *defaultCodec) SetLocation(l *time.Location) {
	dc.l = l
}

func (dc *defaultCodec) GetLocation() *time.Location {
	return dc.l
}

// Encode redis 将 golang object 编码成 redis 请求
func (dc *defaultCodec) Encode(typ EncodeType, v interface{}) (interface{}, error) {
	if v == nil {
		return nil, nil
	}

	switch typ {
	case EncodeTypeHmSET:
		return dc.hmSet(v)
	case EncodeTypeInsertData:
		rv := reflect.Indirect(reflect.ValueOf(v))
		switch rv.Kind() {
		case reflect.Slice, reflect.Array:
			return dc.setStructs(v, true), nil
		case reflect.Struct:
			return dc.setStruct(v, true), nil
		default:
			return nil, fmt.Errorf("insert must be struct or array of struct")
		}
	case EncodeTypeReplaceData:
		rv := reflect.Indirect(reflect.ValueOf(v))
		switch rv.Kind() {
		case reflect.Slice, reflect.Array:
			return dc.setStructs(v, false), nil
		case reflect.Struct:
			return dc.setStruct(v, false), nil
		default:
			return nil, fmt.Errorf("insert must be struct or array of struct")
		}
	case EncodeTypeUpdateData:
		return dc.updateStruct(v), nil
	default:
		switch v.(type) {
		case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64,
			float32, float64, string, []byte, *bool, *int, *int8, *int16, *int32, *int64,
			*uint, *uint8, *uint16, *uint32, *uint64, *float32, *float64, *string, *[]byte:
			return v, nil
		default:
			return dc.encodeBase(reflect.ValueOf(v))
		}
	}
}

func (dc *defaultCodec) Decode(typ consts.RetType, src interface{}, dest []interface{}) (err error) {
	if len(dest) == 0 {
		return nil
	}

	for _, v := range dest {
		rv := reflect.ValueOf(v)
		if rv.Kind() != reflect.Ptr {
			return errors.New("decode dest is not a pointer")
		}

		if rv.IsNil() {
			return errors.New("decode dest is nil pointer")
		}
	}

	if typ >= consts.RedisRetTypeNil && typ <= consts.RedisRetTypeMemberScore { // decode redis result
		switch typ {
		case consts.RedisRetTypeNil:
			return nil
		case consts.RedisRetTypeBool, consts.RedisRetTypeInt64, consts.RedisRetTypeFloat64, consts.RedisRetTypeString:
			err = dc.decodeBase(src, dest[0])
			if err != nil {
				return fmt.Errorf("unmarshal redis result failed: %v", err)
			}

			return nil
		case consts.RedisRetTypeStrings:
			var strArr []interface{}

			switch s := src.(type) {
			case []byte:
				err = dc.um(s, &strArr)
			default:
				err = dc.map2structure.Decode(src, &strArr)
			}

			if err != nil {
				return fmt.Errorf("unmarshal redis strings result failed: %v", err)
			}

			err = dc.decodeMapSlice(strArr, &dest[0])
			if err != nil {
				return fmt.Errorf("decode redis strings to receiver failed: %v", err)
			}
			return nil
		case consts.RedisRetTypeMapString:
			var mapStr map[string]string

			switch s := src.(type) {
			case map[string]string:
				mapStr = s
			case map[string][]byte:
				mapStr = make(map[string]string, len(s))
				for k, v := range s {
					mapStr[k] = types.BytesToString(v)
				}
			case []byte:
				err = dc.um(s, &mapStr)
			default:
				err = dc.map2structure.Decode(src, &mapStr)
			}

			if err != nil {
				return fmt.Errorf("unmarshal redis map-string failed: %v", err)
			}

			switch d := dest[0].(type) {
			case *map[string]string:
				*d = mapStr
				return nil
			case *map[string]interface{}:
				*d = make(map[string]interface{}, len(mapStr))
				for k, v := range mapStr {
					(*d)[k] = v
				}
				return nil
			}

			err = dc.decodeMapSlice(mapStr, &dest[0])
			if err != nil {
				return fmt.Errorf("decode redis map-string to receiver failed: %v", err)
			}

			return nil
		case consts.RedisRetTypeMemberScore:
			var memberScore = new(proto.MemberScore)

			switch s := src.(type) {
			case proto.MemberScore:
				memberScore = &s
			case *proto.MemberScore:
				memberScore = s
			case []byte:
				err = dc.um(s, memberScore)
			default:
				err = dc.map2structure.Decode(src, memberScore)
			}

			if err != nil {
				return fmt.Errorf("unmarshal redis member with score failed: %v", err)
			}

			if msDest, ok := dest[0].(*proto.MemberScore); ok {
				*msDest = *memberScore
				return nil
			}

			if len(dest) >= 2 {
				if v, ok := dest[1].(*[]float64); ok {
					*v = memberScore.Score
				} else {
					return fmt.Errorf("redis member with score, the second param of result receiver must be []float64")
				}
			}

			interfaces := make([]interface{}, len(memberScore.Member))
			for k, v := range memberScore.Member {
				interfaces[k] = v
			}

			err = dc.decodeMapSlice(interfaces, &dest[0])
			if err != nil {
				return fmt.Errorf("decode redis member with score to receiver failed: %v", err)
			}
			return nil
		}
	}

	switch dest0 := dest[0].(type) {
	case *proto.Detail:
		var pageRet = new(proto.PageResult)

		switch s := src.(type) {
		case proto.PageResult:
			pageRet = &s
		case *proto.PageResult:
			pageRet = s
		case []byte:
			err = dc.um(s, pageRet)
		default:
			err = dc.map2structure.Decode(src, pageRet)
		}

		if err != nil {
			return fmt.Errorf("unmarshal page result failed: %v", err)
		}

		if pageRet != nil && pageRet.Detail != nil {
			dest0.Total = pageRet.Detail.Total
			dest0.TotalPage = pageRet.Detail.TotalPage
			dest0.Page = pageRet.Detail.Page
			dest0.Size = pageRet.Detail.Size
			dest0.Scroll = pageRet.Detail.Scroll
			dest0.Extras = pageRet.Detail.Extras

			if len(dest) > 1 {
				err = dc.map2structure.Decode(pageRet.Data, dest[1])
				if err != nil {
					return fmt.Errorf("decode data of page result`s to receiver failed: %v", err)
				}
			}
		}

		return nil
	case *proto.PageResult:
		switch s := src.(type) {
		case proto.PageResult:
			*dest0 = s
		case *proto.PageResult:
			*dest0 = *s
		case []byte:
			err = dc.um(s, dest0)
		default:
			err = dc.map2structure.Decode(src, dest0)
		}

		if err != nil {
			return fmt.Errorf("decode page result to receiver failed: %v", err)
		}
		return nil
	case *proto.ModRet:
		switch s := src.(type) {
		case proto.ModRet:
			*dest0 = s
		case *proto.ModRet:
			*dest0 = *s
		case []byte:
			err = dc.um(s, dest0)
		default:
			err = dc.map2structure.Decode(src, dest0)
		}

		if err != nil {
			return fmt.Errorf("decode mod result to receiver failed: %v", err)
		}

		return nil
	}

	switch s := src.(type) {
	case []byte:
		switch dest0 := dest[0].(type) {
		case *[]byte:
			copy(*dest0, s)
		case *string:
			*dest0 = string(s)
		default:
			err = dc.um(s, dest[0])
		}

	case proto.PageResult:
		err = dc.map2structure.Decode(s.Data, dest[0])
	case *proto.PageResult:
		err = dc.map2structure.Decode(s.Data, dest[0])
	default:
		err = dc.map2structure.Decode(src, dest[0])
	}

	if err != nil {
		return fmt.Errorf("decode result to receiver failed: %v", err)
	}

	return nil
}

func (dc *defaultCodec) setStruct(val interface{}, isInsert bool) map[string]interface{} {
	v := reflect.Indirect(reflect.ValueOf(val))

	ss := structs.GetStructSpec(dc.GetTag(), v.Type())

	data := make(map[string]interface{})

	for _, fs := range ss.M {
		if fs.Tag != "orm" {
			continue
		}

		iv := v.Field(fs.I)
		isEmpty := types.IsEmpty(iv)

		if dc.omitEmpty && (fs.OmitEmpty ||
			(isInsert && fs.OmitInsertEmpty) ||
			(!isInsert && fs.OmitReplaceEmpty)) && isEmpty { // 忽略零值
			continue
		}

		//自动插入当前时间，仅在值为零值时才自动赋值
		if (fs.OnCreateTime || fs.OnUpdateTime) && isEmpty {
			data[fs.Column] = dc.getFormatTimeOrData(nowTime(fs.Type), fs)
		} else if fs.OnUniqueID && isInsert && isEmpty && iv.Kind() == reflect.Uint64 {
			data[fs.Column] = snowflake.GenerateID()
		} else {
			data[fs.Column] = dc.getValue(fs, iv)
		}

		if fs.EsID {
			data["_id"] = data[fs.Column]
		}
	}

	return data
}

func (dc *defaultCodec) setStructs(val interface{}, isInsert bool) []map[string]interface{} {
	arrv := reflect.Indirect(reflect.ValueOf(val))

	arrLen := arrv.Len() //数组长度
	if arrLen <= 0 {
		return nil
	}

	ss := structs.GetStructSpec(dc.GetTag(), reflect.Indirect(arrv.Index(0)).Type())

	ignores := dc.getIgnores(ss, arrv, arrLen, isInsert)

	datas := []map[string]interface{}{}

	//插入语句
	for k := 0; k < arrLen; k++ {
		kv := reflect.Indirect(arrv.Index(k))

		data := map[string]interface{}{}

		for name, fs := range ss.M {
			if fs.Tag != "orm" {
				continue
			}

			if ignore := ignores[name]; !ignore {
				iv := kv.Field(fs.I)
				isEmpty := types.IsEmpty(iv)

				//自动插入当前时间，仅在值为零值时才自动赋值
				if (fs.OnCreateTime || fs.OnUpdateTime) && isEmpty {
					data[fs.Column] = dc.getFormatTimeOrData(nowTime(fs.Type), fs)
				} else if fs.OnUniqueID && isInsert && isEmpty && iv.Kind() == reflect.Uint64 {
					data[fs.Column] = snowflake.GenerateID()
				} else {
					data[fs.Column] = dc.getValue(fs, iv)
				}

				if fs.EsID {
					data["_id"] = data[fs.Column]
				}
			}
		}

		datas = append(datas, data)
	}

	return datas
}

func (dc *defaultCodec) updateStruct(val interface{}) map[string]interface{} {
	v := reflect.Indirect(reflect.ValueOf(val))

	ss := structs.GetStructSpec(dc.GetTag(), v.Type())

	data := make(map[string]interface{})

	for name, fs := range ss.M {
		iv := v.FieldByName(name)

		if dc.omitEmpty && (fs.OmitUpdateEmpty || fs.OmitEmpty) && types.IsEmpty(iv) { //UPDATE 忽略零值
			continue
		}

		if fs.OnUpdateTime && types.IsEmpty(iv) { //修改时自动赋值当前时间，仅在值为零值时才自动赋值
			data[fs.Column] = dc.getFormatTimeOrData(nowTime(fs.Type), fs)
		} else {
			data[fs.Column] = dc.getValue(fs, iv)
		}

	}
	return data
}

func (dc *defaultCodec) getValue(fs *structs.FieldSpec, iv reflect.Value) interface{} {
	if !iv.CanInterface() {
		return nil
	}

	data := iv.Interface()

	if fs.Type == structs.TypeJSON {
		switch data.(type) {
		case bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64,
			float32, float64, string, []byte, *bool, *int, *int8, *int16, *int32, *int64,
			*uint, *uint8, *uint16, *uint32, *uint64, *float32, *float64, *string, *[]byte:
			return data
		default:
			ret, _ := dc.encodeBase(iv)
			return ret
		}
	} else {
		return dc.getFormatTimeOrData(data, fs)
	}
}

func (dc *defaultCodec) getFormatTimeOrData(data interface{}, fs *structs.FieldSpec) interface{} {
	if fs.TimeFmt != "" {
		t, ok := types.GetRealTime(data)
		if ok {
			return t.Format(fs.TimeFmt)
		}
	}

	return data
}

// getIgnores 获取忽略字段
func (dc *defaultCodec) getIgnores(ss *structs.StructSpec,
	arrv reflect.Value, arrLen int, isInsert bool) map[string]bool {
	//获取忽略字段
	ignores := map[string]bool{}
	for name, fs := range ss.M {
		if dc.omitEmpty && (fs.OmitEmpty || (isInsert && fs.OmitInsertEmpty) || (!isInsert && fs.OmitReplaceEmpty)) {
			ignores[name] = true
		} else {
			ignores[name] = false
		}
	}

	for k := 0; k < arrLen; k++ {
		kv := reflect.Indirect(arrv.Index(k))

		for name := range ss.M {
			if ignore := ignores[name]; ignore {
				iv := kv.FieldByName(name)
				if !types.IsEmpty(iv) { // 存在非空值，则该字段不忽略
					ignores[name] = false
				}
			}
		}
	}

	return ignores
}

func nowTime(typ structs.Type) interface{} {
	switch typ {
	case structs.TypeInt, structs.TypeInt32, structs.TypeInt64,
		structs.TypeUint, structs.TypeUint32, structs.TypeUint64:
		return time.Now().Unix()
	default:
		return time.Now()
	}
}
