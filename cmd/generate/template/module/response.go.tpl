package dto{{.PackageName}}

import (
	"{{.ModulePath}}/{{.AppPathInProject}}/object/obj{{.PackageName}}"
	"github.com/morehao/golib/gobject"
)

type {{.StructName}}CreateResp struct {
	// ID 数据自增 ID
	ID uint `json:"id"`
}

type {{.StructName}}DetailResp struct {
	// ID 数据自增 ID
	ID uint `json:"id" validate:"required"`
	obj{{.PackageName}}.{{.StructName}}BaseInfo
	gobject.OperatorBaseInfo

}

type {{.StructName}}PageListItem struct {
	// ID 数据自增 ID
	ID        uint `json:"id" validate:"required"`
	obj{{.PackageName}}.{{.StructName}}BaseInfo
	gobject.OperatorBaseInfo
}

type {{.StructName}}PageListResp struct {
	// List 数据列表
	List []{{.StructName}}PageListItem `json:"list"`
	// Total 数据总条数
	Total int64          `json:"total"`
}
