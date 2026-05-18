package code

import "github.com/morehao/golib/gerror"

var globalErrorMap = make(gerror.CodeMsgMap)

// GetError 根据错误码获取错误
func GetError(code int) error {
	if msg, ok := globalErrorMap[code]; ok {
		return gerror.Error{Code: code, Msg: msg}
	}
	return gerror.Error{Code: code, Msg: "未知错误"}
}

// registerError 注册错误码
func registerError(codeMap gerror.CodeMsgMap) {
	for code, msg := range codeMap {
		globalErrorMap[code] = msg
	}
}

func init() {
	registerError(userErrorMsgMap)
}
