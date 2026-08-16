package dto{{.PackageName}}

import (
	"{{.BaseModulePath}}/{{.AppModuleName}}/object/obj{{.PackageName}}"
	"github.com/morehao/golib/biz/gobject"
)

type {{.StructName}}CreateReq struct {
	obj{{.PackageName}}.{{.StructName}}BaseInfo
}

type {{.StructName}}UpdateReq struct {
	{{.StructName}}ID {{.PKFieldType}} `json:"-" uri:"{{.StructNameLowerCamel}}ID" binding:"required"` // 主键 ID
	obj{{.PackageName}}.{{.StructName}}BaseInfo
}

type {{.StructName}}DetailReq struct {
	{{.StructName}}ID {{.PKFieldType}} `json:"-" uri:"{{.StructNameLowerCamel}}ID" binding:"required"` // 主键 ID
}

type {{.StructName}}PageListReq struct {
	gobject.PageQuery
}

type {{.StructName}}DeleteReq struct {
	{{.StructName}}ID {{.PKFieldType}} `json:"-" uri:"{{.StructNameLowerCamel}}ID" binding:"required"` // 主键 ID
}
