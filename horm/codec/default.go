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
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/horm-database/common/consts"
	"github.com/horm-database/common/proto"
	"github.com/horm-database/common/types"
)

type defaultCodec struct {
	m             Marshal
	um            Unmarshal
	tag           string // the tag in the structure, use for encode and decode, if the tag does not exist, use the name of the structure as the field
	map2structure *Map2Structure
	l             *time.Location
}

type EncodeType int8

const (
	EncodeTypeInsertData  EncodeType = 1
	EncodeTypeReplaceData EncodeType = 2
	EncodeTypeUpdateData  EncodeType = 3
	EncodeTypeRedisVal    EncodeType = 4
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

// Encode 编码
func (dc *defaultCodec) Encode(typ EncodeType, data interface{}) (interface{}, error) {
	if data == nil {
		return nil, nil
	}

	val := types.Indirect(data)
	switch v := val.(type) {
	case string, []byte, bool, int, int8, int16, int32, int64, uint,
		uint8, uint16, uint32, uint64, float32, float64, json.Number, time.Time,
		types.Map, []types.Map, map[string]interface{}, []map[string]interface{}:
		return v, nil
	}

	// 结构体 转 map
	var op int8 = types.OpUnknown

	switch typ {
	case EncodeTypeInsertData:
		op = types.OpInsert
	case EncodeTypeReplaceData:
		op = types.OpReplace
	case EncodeTypeUpdateData:
		op = types.OpUpdate
	}

	rv := reflect.ValueOf(val)
	if types.IsStruct(rv.Type()) {
		return types.StructToMap(rv, dc.GetTag(), op), nil
	} else if types.IsStructArray(rv) {
		return types.StructsToMaps(rv, dc.GetTag(), op), nil
	}

	return data, nil
}

// Decode 解码
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
