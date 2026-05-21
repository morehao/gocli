package router

import (
	"{{.BaseModulePath}}/{{.AppModuleName}}/internal/controller/ctr{{.PackageName}}"
	"github.com/morehao/golib/biz/gconstant"
	"github.com/morehao/golib/biz/gserver/ginserver"
)

// {{.StructNameLowerCamel}}Router 初始化{{.Description}}路由信息
func {{.StructNameLowerCamel}}Router(groups *ginserver.RouterGroups) {
	{{.StructNameLowerCamel}}Ctr := ctr{{.PackageName}}.New{{.StructName}}Ctr()

	v1RouterGroup := groups.MustGetGroup(gconstant.ApiVersionV1)

	v1RouterGroup.POST("/{{.StructNameLowerCamel}}/create", {{.StructNameLowerCamel}}Ctr.Create)
	v1RouterGroup.POST("/{{.StructNameLowerCamel}}/delete", {{.StructNameLowerCamel}}Ctr.Delete)
	v1RouterGroup.POST("/{{.StructNameLowerCamel}}/update", {{.StructNameLowerCamel}}Ctr.Update)
	v1RouterGroup.GET("/{{.StructNameLowerCamel}}/detail", {{.StructNameLowerCamel}}Ctr.Detail)
	v1RouterGroup.POST("/{{.StructNameLowerCamel}}/pageList", {{.StructNameLowerCamel}}Ctr.PageList)
}
