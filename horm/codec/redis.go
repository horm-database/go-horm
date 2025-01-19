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
	j "encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"

	redigo "github.com/gomodule/redigo/redis"
	"github.com/horm-database/common/types"
)

func (dc *defaultCodec) decodeBase(src interface{}, dest interface{}) (err error) {
	switch s := src.(type) {
	case nil:
		// ignore
	case []byte:
		return dc.assignBytes(dest, s)
	case string:
		return dc.assignString(dest, s)
	case int64:
		return dc.assignInt64(dest, s)
	case float64:
		return dc.assignFloat64(dest, s)
	case bool:
		return dc.assignBool(dest, s)
	case []interface{}:
		return dc.assignInterfaces(dest, s)
	case j.Number:
		return dc.assignString(dest, s.String())
	case redigo.Error:
		err = s
	default:
		if dest == nil {
			return nil
		}
		err = dc.cvtFailure(reflect.ValueOf(dest), s)
	}
	return
}

func (dc *defaultCodec) decodeMapSlice(src interface{}, dest interface{}) (err error) {
	rv := inv(reflect.ValueOf(dest))

	switch s := src.(type) {
	case map[string]string:
		switch rv.Kind() {
		case reflect.Struct:
			return dc.decodeMap2Struct(s, rv)
		case reflect.Map:
			return dc.decodeMap2Map(s, rv)
		}
	case []interface{}:
		switch rv.Kind() {
		case reflect.Struct:
			return dc.decodeSlice2Struct(s, rv)
		case reflect.Map:
			return dc.decodeSlice2Map(s, rv)
		case reflect.Slice:
			return dc.decodeSlice2Slice(s, rv)
		}
	}

	return dc.cvtFailure(rv, src)
}

func (dc *defaultCodec) decodeMap2Struct(src map[string]string, rv reflect.Value) error {
	ss := types.GetStructDesc(dc.GetTag(), rv.Type())
	for name, s := range src {
		fs := ss.Cm[name]
		if fs == nil {
			continue
		}

		if err := dc.cvtAssignValue(rv.FieldByIndex(fs.Index), s); err != nil {
			return fmt.Errorf("redigo.ScanStruct: cannot assign field %s: %v", fs.Name, err)
		}
	}

	return nil
}

func (dc *defaultCodec) decodeMap2Map(src map[string]string, d reflect.Value) error {
	if d.Kind() != reflect.Map {
		return errors.New("dest is not map")
	}

	if d.IsNil() {
		d.Set(reflect.MakeMap(d.Type()))
	}

	keyType := d.Type().Key()
	elemType := d.Type().Elem()

	for k, v := range src {
		key := reflect.New(keyType).Elem()
		if err := dc.cvtAssignValue(key, k); err != nil {
			return err
		}
		tmp := reflect.New(elemType)
		value := tmp.Elem()
		if err := dc.cvtAssignValue(value, v); err != nil {
			return err
		}
		d.SetMapIndex(key, value)
	}
	return nil
}

func (dc *defaultCodec) decodeSlice2Struct(src []interface{}, rv reflect.Value) error {
	if len(src)%2 != 0 {
		return errors.New("number of values not a multiple of 2")
	}

	ss := types.GetStructDesc(dc.GetTag(), rv.Type())
	for i := 0; i < len(src); i += 2 {
		s := src[i+1]
		if s == nil {
			continue
		}
		name, err := getKey(src[i])
		if err != nil {
			return err
		}
		fs := ss.Cm[name]
		if fs == nil {
			continue
		}
		if err := dc.cvtAssignValue(rv.FieldByIndex(fs.Index), s); err != nil {
			return fmt.Errorf("redigo.ScanStruct: cannot assign field %s: %v", fs.Name, err)
		}
	}
	return nil
}

func (dc *defaultCodec) decodeSlice2Map(src []interface{}, d reflect.Value) error {
	if len(src)%2 != 0 {
		return errors.New("number of values not a multiple of 2")
	}

	if d.Kind() != reflect.Map {
		return errors.New("dest is not map")
	}

	if d.IsNil() {
		d.Set(reflect.MakeMap(d.Type()))
	}
	keyType := d.Type().Key()
	elemType := d.Type().Elem()
	for i := 0; i < len(src); i += 2 {
		key := reflect.New(keyType).Elem()
		if err := dc.cvtAssignValue(key, src[i]); err != nil {
			return err
		}
		tmp := reflect.New(elemType)
		value := tmp.Elem()
		if err := dc.cvtAssignValue(value, src[i+1]); err != nil {
			return err
		}
		d.SetMapIndex(key, value)
	}
	return nil
}

func (dc *defaultCodec) decodeSlice2Slice(src []interface{}, rv reflect.Value) error {
	if rv.Kind() != reflect.Slice {
		return errors.New("dest is not map")
	}
	ensureLen(rv, len(src))
	for i, v := range src {
		if v != nil {
			value := reflect.New(rv.Type().Elem()).Elem()
			if err := dc.cvtAssignValue(value, v); err != nil {
				return err
			}
			rv.Index(i).Set(value)
		}
	}
	return nil
}

func (dc *defaultCodec) assignBytes(d interface{}, s []byte) (err error) {
	switch d := d.(type) {
	case *string:
		*d = string(s)
	case *int:
		*d, err = strconv.Atoi(types.BytesToString(s))
	case *bool:
		*d, err = strconv.ParseBool(types.BytesToString(s))
	case *[]byte:
		*d = s
	case *interface{}:
		*d = s
	case nil:
		// skip value
	default:
		if d := reflect.ValueOf(d); d.Type().Kind() != reflect.Ptr {
			err = dc.cvtFailure(d, s)
		} else {
			err = dc.cvtAssignBytes(d.Elem(), s)
		}
	}
	return err
}

func (dc *defaultCodec) assignString(d interface{}, s string) (err error) {
	switch d := d.(type) {
	case *string:
		*d = s
	case *interface{}:
		*d = s
	case nil:
		// skip value
	default:
		if d := reflect.ValueOf(d); d.Type().Kind() != reflect.Ptr {
			err = dc.cvtFailure(d, s)
		} else {
			err = dc.cvtAssignString(d.Elem(), s)
		}
	}
	return err
}

func (dc *defaultCodec) assignInt64(d interface{}, s int64) (err error) {
	switch d := d.(type) {
	case *int64:
		*d = s
	case *int:
		x := int(s)
		if int64(x) != s {
			err = strconv.ErrRange
			x = 0
		}
		*d = x
	case *bool:
		*d = s != 0
	case *interface{}:
		*d = s
	case nil:
		// skip value
	default:
		if d := reflect.ValueOf(d); d.Type().Kind() != reflect.Ptr {
			err = dc.cvtFailure(d, s)
		} else {
			err = dc.cvtAssignInt(d.Elem(), s)
		}
	}
	return err
}

func (dc *defaultCodec) assignFloat64(d interface{}, s float64) (err error) {
	switch d := d.(type) {
	case *float64:
		*d = s
	case *float32:
		x := float32(s)
		if float64(x) != s {
			err = strconv.ErrRange
			x = 0
		}
		*d = x
	case *interface{}:
		*d = s
	case nil:
		// skip value
	default:
		if d := reflect.ValueOf(d); d.Type().Kind() != reflect.Ptr {
			err = dc.cvtFailure(d, s)
		} else {
			err = dc.cvtAssignFloat(d.Elem(), s)
		}
	}
	return err
}

func (dc *defaultCodec) assignBool(d interface{}, s bool) (err error) {
	switch d := d.(type) {
	case *bool:
		*d = s
	case *interface{}:
		*d = s
	case nil:
		// skip value
	default:
		if d := reflect.ValueOf(d); d.Type().Kind() != reflect.Ptr {
			err = dc.cvtFailure(d, s)
		} else {
			err = dc.cvtAssignBool(d.Elem(), s)
		}
	}
	return err
}

func (dc *defaultCodec) assignInterfaces(d interface{}, s []interface{}) (err error) {
	switch d := d.(type) {
	case *[]interface{}:
		*d = s
	case *interface{}:
		*d = s
	case nil:
		// skip value
	default:
		if d := reflect.ValueOf(d); d.Type().Kind() != reflect.Ptr {
			err = dc.cvtFailure(d, s)
		} else {
			err = dc.cvtAssignArray(d.Elem(), s)
		}
	}
	return err
}

func (dc *defaultCodec) cvtAssignValue(d reflect.Value, s interface{}) (err error) {
	switch s := s.(type) {
	case nil:
		err = dc.cvtAssignNil(d)
	case []byte:
		err = dc.cvtAssignBytes(d, s)
	case string:
		err = dc.cvtAssignString(d, s)
	case int64:
		err = dc.cvtAssignInt(d, s)
	case float64:
		err = dc.cvtAssignFloat(d, s)
	case bool:
		err = dc.cvtAssignBool(d, s)
	case redigo.Error:
		err = dc.cvtAssignError(d, s)
	default:
		err = dc.cvtFailure(d, s)
	}
	return err
}

func (dc *defaultCodec) cvtAssignNil(d reflect.Value) (err error) {
	switch d.Type().Kind() {
	case reflect.Slice, reflect.Interface, reflect.Map, reflect.Ptr:
		d.Set(reflect.Zero(d.Type()))
	default:
		err = dc.cvtFailure(d, nil)
	}
	return err
}

func (dc *defaultCodec) cvtAssignBytes(d reflect.Value, s []byte) (err error) {
	switch d.Type().Kind() {
	case reflect.Float32, reflect.Float64:
		var x float64
		x, err = strconv.ParseFloat(types.BytesToString(s), d.Type().Bits())
		d.SetFloat(x)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var x int64
		x, err = strconv.ParseInt(types.BytesToString(s), 10, d.Type().Bits())
		d.SetInt(x)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var x uint64
		x, err = strconv.ParseUint(types.BytesToString(s), 10, d.Type().Bits())
		d.SetUint(x)
	case reflect.Bool:
		var x bool
		x, err = strconv.ParseBool(types.BytesToString(s))
		d.SetBool(x)
	case reflect.String:
		d.Addr()
		d.SetString(string(s))
	case reflect.Slice:
		// Handle []byte destination here to avoid unnecessary
		// []byte -> string -> []byte converion.
		if d.Type().Elem().Kind() == reflect.Uint8 {
			d.SetBytes(s)
		} else {
			err = dc.um(s, d.Addr().Interface())
		}
	case reflect.Map:
		err = dc.um(s, d.Addr().Interface())
	case reflect.Ptr:
		if d.IsNil() {
			d.Set(reflect.New(d.Type().Elem()))
		}
		err = dc.cvtAssignBytes(d.Elem(), s)
	case reflect.Interface:
		if d.Type() == iType {
			d.Set(reflect.ValueOf(s))
		} else {
			err = dc.cvtFailure(d, s)
		}
	case reflect.Struct:
		err = dc.um(s, d.Addr().Interface())
	default:
		err = dc.cvtFailure(d, s)
	}
	return
}

func (dc *defaultCodec) cvtAssignString(d reflect.Value, s string) (err error) {
	switch d.Type().Kind() {
	case reflect.Float32, reflect.Float64:
		var x float64
		x, err = strconv.ParseFloat(s, d.Type().Bits())
		d.SetFloat(x)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		var x int64
		x, err = strconv.ParseInt(s, 10, d.Type().Bits())
		d.SetInt(x)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		var x uint64
		x, err = strconv.ParseUint(s, 10, d.Type().Bits())
		d.SetUint(x)
	case reflect.Bool:
		var x bool
		x, err = strconv.ParseBool(s)
		d.SetBool(x)
	case reflect.String:
		d.Addr()
		d.SetString(s)
	case reflect.Slice:
		if d.Type().Elem().Kind() == reflect.Uint8 {
			d.SetBytes(types.StringToBytes(s))
		} else {
			err = dc.um(types.StringToBytes(s), d.Addr().Interface())
		}
	case reflect.Map:
		if d.IsNil() {
			d.Set(reflect.New(d.Type()).Elem())
		}
		err = dc.um(types.StringToBytes(s), d.Addr().Interface())
	case reflect.Ptr:
		if d.IsNil() {
			d.Set(reflect.New(d.Type().Elem()))
		}
		err = dc.cvtAssignString(d.Elem(), s)
	case reflect.Interface:
		if d.Type() == iType {
			d.Set(reflect.ValueOf(s))
		} else {
			err = dc.cvtFailure(d, s)
		}
	case reflect.Struct:
		err = dc.um(types.StringToBytes(s), d.Addr().Interface())
	default:
		err = dc.cvtFailure(d, s)
	}
	return
}

func (dc *defaultCodec) cvtAssignInt(d reflect.Value, s int64) (err error) {
	switch d.Type().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		d.SetInt(s)
		if d.Int() != s {
			err = strconv.ErrRange
			d.SetInt(0)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if s < 0 {
			err = strconv.ErrRange
		} else {
			x := uint64(s)
			d.SetUint(x)
			if d.Uint() != x {
				err = strconv.ErrRange
				d.SetUint(0)
			}
		}
	case reflect.Bool:
		d.SetBool(s != 0)
	case reflect.String:
		d.SetString(fmt.Sprint(s))
	case reflect.Slice:
		if d.Type().Elem().Kind() == reflect.Uint8 {
			d.Set(reflect.ValueOf(types.StringToBytes(fmt.Sprint(s))))
		}
	case reflect.Interface:
		d.Set(reflect.ValueOf(s))
	default:
		err = dc.cvtFailure(d, s)
	}
	return
}

func (dc *defaultCodec) cvtAssignFloat(d reflect.Value, s float64) (err error) {
	switch d.Type().Kind() {
	case reflect.Float64, reflect.Float32:
		d.SetFloat(s)
		if d.Float() != s {
			err = strconv.ErrRange
			d.SetFloat(0)
		}
	case reflect.String:
		d.SetString(fmt.Sprint(s))
	case reflect.Slice:
		if d.Type().Elem().Kind() == reflect.Uint8 {
			d.Set(reflect.ValueOf(types.StringToBytes(fmt.Sprint(s))))
		}
	case reflect.Interface:
		d.Set(reflect.ValueOf(s))
	default:
		err = dc.cvtFailure(d, s)
	}
	return
}

func (dc *defaultCodec) cvtAssignBool(d reflect.Value, s bool) (err error) {
	switch d.Type().Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if s {
			d.SetInt(1)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if s {
			d.SetUint(1)
		}
	case reflect.Bool:
		d.SetBool(s)
	case reflect.Interface:
		d.Set(reflect.ValueOf(s))
	default:
		err = dc.cvtFailure(d, s)
	}
	return
}

func (dc *defaultCodec) cvtAssignError(d reflect.Value, s redigo.Error) (err error) {
	if d.Kind() == reflect.String {
		d.SetString(s.Error())
	} else if d.Kind() == reflect.Slice && d.Type().Elem().Kind() == reflect.Uint8 {
		d.SetBytes(types.StringToBytes(s.Error()))
	} else {
		err = dc.cvtFailure(d, s)
	}
	return
}

func (dc *defaultCodec) cvtAssignArray(d reflect.Value, s []interface{}) error {
	if d.Type().Kind() != reflect.Slice {
		return dc.cvtFailure(d, s)
	}
	ensureLen(d, len(s))
	for i := 0; i < len(s); i++ {
		if err := dc.cvtAssignValue(d.Index(i), s[i]); err != nil {
			return err
		}
	}
	return nil
}

func (dc *defaultCodec) cvtFailure(d reflect.Value, s interface{}) error {
	var stype string
	switch s.(type) {
	case string:
		stype = "orm simple string"
	case redigo.Error:
		stype = "orm error"
	case int64:
		stype = "orm integer"
	case []byte:
		stype = "orm bytes"
	case []interface{}:
		stype = "orm array"
	case nil:
		stype = "orm nil"
	default:
		stype = reflect.TypeOf(s).String()
	}
	return fmt.Errorf("cannot convert from %s to %s", stype, d.Type())
}

func ensureLen(d reflect.Value, n int) {
	if n > d.Cap() {
		d.Set(reflect.MakeSlice(d.Type(), n, n))
	} else {
		d.SetLen(n)
	}
}

func inv(v reflect.Value) reflect.Value {
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		fallthrough
	case reflect.Interface:
		return inv(v.Elem())
	default:
		return v
	}
}

func getKey(val interface{}) (string, error) {
	switch v := val.(type) {
	case string:
		return v, nil
	case []byte:
		return types.BytesToString(v), nil
	default:
		return "", fmt.Errorf("key %d must string or bytes value", v)
	}
}

var iType = reflect.ValueOf(struct {
	e interface{}
}{}).Field(0).Type()
