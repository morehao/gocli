package dto{{.PackageName}}

import (
	"{{.ModulePath}}/{{.AppPathInProject}}/object/objcommon"
	"{{.ModulePath}}/{{.AppPathInProject}}/object/obj{{.PackageName}}"
)

type {{.StructName}}CreateResp struct {
	ID uint `json:"id"` // 数据自增id
}

type {{.StructName}}DetailResp struct {
	ID        uint `json:"id" validate:"required"` // 数据自增id
	obj{{.PackageName}}.{{.StructName}}BaseInfo
	objcommon.OperatorBaseInfo

}

type {{.StructName}}PageListItem struct {
	ID        uint `json:"id" validate:"required"` // 数据自增id
	obj{{.PackageName}}.{{.StructName}}BaseInfo
	objcommon.OperatorBaseInfo
}

type {{.StructName}}PageListResp struct {
	List  []{{.StructName}}PageListItem `json:"list"`  // 数据列表
	Total int64          `json:"total"` // 数据总条数
}
