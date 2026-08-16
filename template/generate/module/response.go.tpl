package dto{{.PackageName}}

import (
	"{{.BaseModulePath}}/{{.AppModuleName}}/object/obj{{.PackageName}}"
	"github.com/morehao/golib/biz/gobject"
)

type {{.StructName}}CreateResp struct {
	{{.StructName}}ID {{.PKFieldType}} `json:"{{.StructNameLowerCamel}}ID"` // 主键 ID
}

type {{.StructName}}DetailResp struct {
	{{.StructName}}ID {{.PKFieldType}} `json:"{{.StructNameLowerCamel}}ID" binding:"required"` // 主键 ID
	obj{{.PackageName}}.{{.StructName}}BaseInfo
	gobject.OperatorBaseInfo

}

type {{.StructName}}PageListItem struct {
	{{.StructName}}ID {{.PKFieldType}} `json:"{{.StructNameLowerCamel}}ID" binding:"required"` // 主键 ID
	obj{{.PackageName}}.{{.StructName}}BaseInfo
	gobject.OperatorBaseInfo
}

type {{.StructName}}PageListResp struct {
	List []{{.StructName}}PageListItem `json:"list"` // 数据列表
	Total int64          `json:"total"` // 数据总条数
}
