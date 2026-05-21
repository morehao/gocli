package router

import (
    "{{.BaseModulePath}}/{{.AppModuleName}}/internal/controller/ctr{{.PackageName}}"
	"github.com/morehao/golib/biz/gconstant"
	"github.com/morehao/golib/biz/gserver/ginserver"
)
{{if .IsNewRouter}}
// {{.StructNameLowerCamel}}Router 初始化{{.Description}}路由信息
func {{.StructNameLowerCamel}}Router(groups *ginserver.RouterGroups) {
	{{.StructNameLowerCamel}}Ctr := ctr{{.PackageName}}.New{{.StructName}}Ctr()

	v1RouterGroup := groups.MustGetGroup(gconstant.ApiVersionV1)

	v1RouterGroup.{{.HttpMethod}}("/{{.StructNameLowerCamel}}/{{.FunctionNameLowerCamel}}", {{.StructNameLowerCamel}}Ctr.{{.FunctionName}})
}
{{end}}
