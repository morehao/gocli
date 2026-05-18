package dtouser

import (
	"github.com/morehao/example/apps/demoapp/object/objuser"
	"github.com/morehao/golib/biz/gobject"
)

type UserCreateReq struct {
	objuser.UserBaseInfo
}

type UserUpdateReq struct {
	UserID uint `json:"userID" validate:"required" label:"User自增id"` // 自增 ID
	objuser.UserBaseInfo
}

type UserDetailReq struct {
	UserID uint `json:"userID" form:"userID" validate:"required" label:"User自增id"` // 自增 ID
}

type UserPageListReq struct {
	gobject.PageQuery
}

type UserDeleteReq struct {
	UserID uint `json:"userID" form:"userID" validate:"required" label:"User自增id"` // 自增 ID
}
