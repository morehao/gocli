package dto{{.PackageName}}

import (
	"{{.ModulePath}}/{{.AppPathInProject}}/object/obj{{.PackageName}}"
	"github.com/morehao/golib/biz/gobject"
)

type {{.StructName}}CreateReq struct {
	obj{{.PackageName}}.{{.StructName}}BaseInfo
}

type {{.StructName}}UpdateReq struct {
	ID uint `json:"id" validate:"required" label:"数据自增id"` // 数据自增 ID
	obj{{.PackageName}}.{{.StructName}}BaseInfo
}

type {{.StructName}}DetailReq struct {
	ID uint `json:"id" form:"id" validate:"required" label:"数据自增id"` // 数据自增 ID
}

type {{.StructName}}PageListReq struct {
	gobject.PageQuery
}

type {{.StructName}}DeleteReq struct {
	ID uint `json:"id" form:"id" validate:"required" label:"数据自增id"` // 数据自增 ID
}
