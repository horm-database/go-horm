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
	"reflect"
	"time"

	"github.com/araddon/dateparse"
	"github.com/horm-database/common/codec/mapstructure"
	"github.com/horm-database/common/types"
)

type Map2Structure struct {
	tagName    string
	weaklyType bool
	squash     bool
	l          *time.Location
}

var typeString = reflect.TypeOf("")

func (m *Map2Structure) Decode(src, dest interface{}) error {
	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: m.weaklyType,
		Squash:           m.squash,
		Result:           dest,
		TagName:          m.tagName,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			func(str reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
				if str == typeString && types.IsTime(t) {
					var tt time.Time
					var err error

					if m.l == nil {
						tt, err = dateparse.ParseAny(data.(string))
					} else {
						tt, err = dateparse.ParseIn(data.(string), m.l)
					}

					if err != nil {
						return nil, err
					}

					return reflect.ValueOf(tt).Convert(t).Interface(), err
				}

				return data, nil
			},
		),
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(src)
}
