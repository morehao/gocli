package code

import "github.com/morehao/golib/gerror"

const (
	UserLoginLogCreateError      = 100100
	UserLoginLogDeleteError      = 100101
	UserLoginLogUpdateError      = 100102
	UserLoginLogGetDetailError   = 100103
	UserLoginLogGetPageListError = 100104
	UserLoginLogNotExistError    = 100105
)

var userLoginLogErrorMsgMap = gerror.CodeMsgMap{
	UserLoginLogCreateError:      "创建用户登录记录失败",
	UserLoginLogDeleteError:      "删除用户登录记录失败",
	UserLoginLogUpdateError:      "修改用户登录记录失败",
	UserLoginLogGetDetailError:   "查看用户登录记录失败",
	UserLoginLogGetPageListError: "查看用户登录记录列表失败",
	UserLoginLogNotExistError:    "用户登录记录不存在",
}
