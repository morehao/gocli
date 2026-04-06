package code

import "fmt"

const (
	UserLoginLogCreateError = 10001 + iota
	UserLoginLogDeleteError
	UserLoginLogUpdateError
	UserLoginLogGetDetailError
	UserLoginLogGetPageListError
	UserLoginLogNotExistError
)

var userLoginLogErrorMsgMap = map[int]string{
	UserLoginLogCreateError:      "创建失败",
	UserLoginLogDeleteError:      "删除失败",
	UserLoginLogUpdateError:      "更新失败",
	UserLoginLogGetDetailError:   "查询详情失败",
	UserLoginLogGetPageListError: "查询列表失败",
	UserLoginLogNotExistError:    "数据不存在",
}

func GetError(code int) error {
	if msg, ok := userLoginLogErrorMsgMap[code]; ok {
		return fmt.Errorf("code=%d,msg=%s", code, msg)
	}
	return fmt.Errorf("code=%d", code)
}
