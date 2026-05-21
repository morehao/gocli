package code

import "github.com/morehao/golib/gerror"

const (
	UserCreateError      = 100100
	UserDeleteError      = 100101
	UserUpdateError      = 100102
	UserGetDetailError   = 100103
	UserGetPageListError = 100104
	UserNotExistError    = 100105
)

var userErrorMsgMap = gerror.CodeMsgMap{
	UserCreateError:      "创建用户登录记录失败",
	UserDeleteError:      "删除用户登录记录失败",
	UserUpdateError:      "修改用户登录记录失败",
	UserGetDetailError:   "查看用户登录记录失败",
	UserGetPageListError: "查看用户登录记录列表失败",
	UserNotExistError:    "用户登录记录不存在",
}
