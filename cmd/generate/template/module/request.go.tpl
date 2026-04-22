package dto{{.PackageName}}

import (
	"{{.ModulePath}}/{{.AppPathInProject}}/object/obj{{.PackageName}}"
	"github.com/morehao/golib/biz/gobject"
)

type {{.StructName}}CreateReq struct {
	obj{{.PackageName}}.{{.StructName}}BaseInfo
}

type {{.StructName}}UpdateReq struct {
	{{.StructName}}ID uint `json:"{{.StructNameLowerCamel}}ID" validate:"required" label:"{{.StructName}}自增id"` // 自增 ID
	obj{{.PackageName}}.{{.StructName}}BaseInfo
}

type {{.StructName}}DetailReq struct {
	{{.StructName}}ID uint `json:"{{.StructNameLowerCamel}}ID" form:"{{.StructNameLowerCamel}}ID" validate:"required" label:"{{.StructName}}自增id"` // 自增 ID
}

type {{.StructName}}PageListReq struct {
	gobject.PageQuery
}

type {{.StructName}}DeleteReq struct {
	{{.StructName}}ID uint `json:"{{.StructNameLowerCamel}}ID" form:"{{.StructNameLowerCamel}}ID" validate:"required" label:"{{.StructName}}自增id"` // 自增 ID
}
