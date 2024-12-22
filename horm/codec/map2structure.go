// Copyright (c) 2024 The horm-database Authors (such as CaoHao <18500482693@163.com>). All rights reserved.
//
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

	"github.com/horm-database/common/codec/mapstructure"
	"github.com/horm-database/common/json"
)

var DefaultMap2Structure = &Map2Structure{DefaultTag, true, true}

type Map2Structure struct {
	tagName    string
	weaklyType bool
	squash     bool
}

var typeTime = reflect.TypeOf(time.Time{})
var typeString = reflect.TypeOf("")

func (m *Map2Structure) Decode(src, dest interface{}) error {
	config := &mapstructure.DecoderConfig{
		WeaklyTypedInput: m.weaklyType,
		Squash:           m.squash,
		Result:           dest,
		TagName:          m.tagName,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
				if t == typeTime && f == typeString {
					tt := time.Time{}
					err := json.Api.Unmarshal([]byte(`"`+data.(string)+`"`), &tt)
					return tt, err
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
