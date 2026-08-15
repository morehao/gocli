package router

import (
	"{{.BaseModulePath}}/{{.AppModuleName}}/internal/controller/ctr{{.PackageName}}"
	"github.com/morehao/golib/biz/gserver/ginserver"
)

// {{.StructNameLowerCamel}}Router 初始化{{.Description}}路由信息
func {{.StructNameLowerCamel}}Router(groups *ginserver.RouterGroups) {
	{{.StructNameLowerCamel}}Ctr := ctr{{.PackageName}}.New{{.StructName}}Ctr()

	v1RouterGroup := groups.MustGetGroup(ginserver.ApiVersionV1)

	v1RouterGroup.POST("/{{toKebabCase (pluralize .StructNameLowerCamel)}}", {{.StructNameLowerCamel}}Ctr.Create)
	v1RouterGroup.GET("/{{toKebabCase (pluralize .StructNameLowerCamel)}}", {{.StructNameLowerCamel}}Ctr.PageList)
	v1RouterGroup.GET("/{{toKebabCase (pluralize .StructNameLowerCamel)}}/:{{.StructNameLowerCamel}}ID", {{.StructNameLowerCamel}}Ctr.Detail)
	v1RouterGroup.PUT("/{{toKebabCase (pluralize .StructNameLowerCamel)}}/:{{.StructNameLowerCamel}}ID", {{.StructNameLowerCamel}}Ctr.Update)
	v1RouterGroup.DELETE("/{{toKebabCase (pluralize .StructNameLowerCamel)}}/:{{.StructNameLowerCamel}}ID", {{.StructNameLowerCamel}}Ctr.Delete)
}
