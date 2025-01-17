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

// Error 返回信息
type Error proto.Error

// RspErrs 所有查询单元的返回码，key 为执行单元 name
type RspErrs map[string]*Error

// GetType 获取执行单元的错误类型
func (re *RspErrs) GetType(name string) errs.EType {
	if err, ok := (*re)[name]; ok {
		return errs.EType(err.Type)
	}

	return errs.ETypeSystem
}

// GetCode 获取执行单元的返回码
func (re *RspErrs) GetCode(name string) int {
	if err, ok := (*re)[name]; ok {
		return int(err.Code)
	}

	return errs.Success
}

// GetMsg 获取执行单元的错误信息
func (re *RspErrs) GetMsg(name string) string {
	if err, ok := (*re)[name]; ok {
		return err.Msg
	}

	return "unknown error"
}

// IsSuccess 执行单元是否成功
func (re *RspErrs) IsSuccess(name string) bool {
	return re.GetCode(name) == errs.Success
}

// Error 根据错误信息返回 error
func (re *RspErrs) Error(name string) error {
	if re.GetCode(name) == errs.Success {
		return nil
	}

	err, ok := (*re)[name]
	if !ok {
		return nil
	}

	return &errs.Error{
		Type: errs.EType(err.Type),
		Code: int(err.Code),
		Msg:  err.Msg,
		Sql:  err.Sql,
	}
}

// IsAllSuccess 判断 es 批量插入是否全部成功
func IsAllSuccess(results []*proto.ModRet) bool {
	for _, ret := range results {
		if ret.Status != 0 {
			return false
		}
	}

	return true
}
