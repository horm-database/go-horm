package horm

import (
	"github.com/horm-database/common/errs"
	"github.com/horm-database/common/proto"
)

// IsAllSuccess 判断 es 批量插入是否全部成功
func IsAllSuccess(results []*proto.ModRet) bool {
	for _, ret := range results {
		if ret.Status != 0 {
			return false
		}
	}

	return true
}

// Error 返回信息
type Error proto.Error

// RspErrs 所有查询单元的返回码，key 为执行单元 name
type RspErrs map[string]*Error

// GetType 获取执行单元的错误类型
func (re *RspErrs) GetType(name string) int8 {
	if err, ok := (*re)[name]; ok {
		return int8(err.Type)
	}

	return errs.ErrorTypeSystem
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
		Type: int8(err.Type),
		Code: int(err.Code),
		Msg:  err.Msg,
	}
}
