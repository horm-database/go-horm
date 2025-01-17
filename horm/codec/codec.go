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
	"time"

	"github.com/horm-database/common/consts"
	"github.com/json-iterator/go"
)

type Marshal func(v interface{}) (string, error)
type Unmarshal func(data []byte, v interface{}) error

type Codec interface {
	SetMarshal(m Marshal, um Unmarshal) // marshal/unmarshal for single item
	GetMarshal() Marshal                // get marshal
	GetUnmarshal() Unmarshal            // get unmarshal

	SetTag(tag string) // set tag
	GetTag() string    // get tag

	Encode(typ EncodeType, v interface{}) (interface{}, error)            // encode
	Decode(typ consts.RetType, src interface{}, dest []interface{}) error // decode

	SetLocation(l *time.Location)
	GetLocation() *time.Location
}

var (
	DefaultTag   = "orm"
	DefaultCodec = NewCodec(jsonMarshal, jsonUnmarshal, DefaultTag, time.Local, true, true)
)

var marshalConfig = jsoniter.Config{
	EscapeHTML: true,
	TagKey:     "json",
}.Froze()

var (
	jsonMarshal = func(v interface{}) (string, error) {
		mv, err := marshalConfig.Marshal(v)
		return string(mv), err
	}
	jsonUnmarshal = func(data []byte, v interface{}) error {
		return marshalConfig.Unmarshal(data, v)
	}
)

func NewCodec(m Marshal, um Unmarshal, tag string, l *time.Location, weaklyType, squash bool) Codec {
	c := defaultCodec{}
	c.m = m
	c.um = um
	c.tag = tag
	c.l = l

	c.map2structure = &Map2Structure{
		tagName:    tag,
		squash:     squash,
		weaklyType: weaklyType,
		l:          l,
	}

	return &c
}
