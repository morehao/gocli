package code

import (
	"fmt"

	"github.com/morehao/golib/biz/gconstant"
	"github.com/morehao/golib/biz/genericdao"
	"github.com/morehao/golib/gerror"
)

var errorMap = gerror.ErrorMap{}

func registerError(codeMsgMap gerror.CodeMsgMap) {
	for code, msg := range codeMsgMap {

		if _, ok := errorMap[code]; ok {
			panic(fmt.Sprintf("error code %d already exists", code))
		}
		errorMap[code] = gerror.Error{
			Code: code,
			Msg:  msg,
		}
	}
}

func GetError(code int) *gerror.Error {
	err := errorMap[code]
	return &err
}

func init() {
	// 业务错误码规范: 从 1002XX 开始
	// 模块划分: 1002XX(租户) 1003XX(公司) 1004XX(部门) 1005XX(用户) 1006XX(菜单) 1007XX(角色)
	registerError(genericdao.DBErrorMsgMap)
	registerError(gconstant.SystemErrorMsgMap)
	registerError(gconstant.AuthErrorMsgMap)
}
