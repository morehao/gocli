package dto{{.PackageName}}

import (
	"{{.ModulePath}}/{{.AppPathInProject}}/object/obj{{.PackageName}}"
	"github.com/morehao/golib/biz/gobject"
)

type {{.StructName}}CreateResp struct {
	ID uint `json:"id"` // 数据自增 ID
}

type {{.StructName}}DetailResp struct {
	ID uint `json:"id" validate:"required"` // 数据自增 ID
	obj{{.PackageName}}.{{.StructName}}BaseInfo
	gobject.OperatorBaseInfo

}

type {{.StructName}}PageListItem struct {
	ID        uint `json:"id" validate:"required"` // 数据自增 ID
	obj{{.PackageName}}.{{.StructName}}BaseInfo
	gobject.OperatorBaseInfo
}

type {{.StructName}}PageListResp struct {
	List []{{.StructName}}PageListItem `json:"list"` // 数据列表
	Total int64          `json:"total"` // 数据总条数
}
