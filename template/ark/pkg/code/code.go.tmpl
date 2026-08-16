package code

import (
	"fmt"

	"github.com/morehao/golib/gconstant"
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
	// 业务错误码规范: 从 1001XX 开始
	// 模块划分: 1001XX(组织) 1002XX(租户) 1004XX(部门) 1005XX(用户) 1006XX(菜单) 1007XX(角色) 1008XX(日志) 1009XX(API密钥)
	// 注: 100100-100109 被 application 使用
	registerError(gconstant.DBErrorMsgMap)
	registerError(gconstant.SystemErrorMsgMap)
	registerError(gconstant.AuthErrorMsgMap)
	registerError(organizationErrorMsgMap)
	registerError(tenantErrorMsgMap)
	registerError(departmentErrorMsgMap)
	registerError(menuErrorMsgMap)
	registerError(roleErrorMsgMap)
	registerError(authErrorMsgMap)
	registerError(applicationErrorMsgMap)
	registerError(operationLogErrorMsgMap)
	registerError(loginLogErrorMsgMap)
	registerError(apiKeyErrorMsgMap)
	registerError(oidcErrorMsgMap)
	registerError(userErrorMsgMap)
}
