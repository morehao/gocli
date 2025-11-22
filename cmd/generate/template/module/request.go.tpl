package dto{{.PackageName}}

import (
	"{{.ModulePath}}/{{.AppPathInProject}}/object/objcommon"
	"{{.ModulePath}}/{{.AppPathInProject}}/object/obj{{.PackageName}}"
)

type {{.StructName}}CreateReq struct {
	obj{{.PackageName}}.{{.StructName}}BaseInfo
}

type {{.StructName}}UpdateReq struct {
	// ID 数据自增 ID
	ID uint `json:"id" validate:"required" label:"数据自增id"`
	obj{{.PackageName}}.{{.StructName}}BaseInfo
}

type {{.StructName}}DetailReq struct {
	// ID 数据自增 ID
	ID uint `json:"id" form:"id" validate:"required" label:"数据自增id"`
}

type {{.StructName}}PageListReq struct {
	objcommon.PageQuery
}

type {{.StructName}}DeleteReq struct {
	// ID 数据自增 ID
	ID uint `json:"id" form:"id" validate:"required" label:"数据自增id"`
}
