package router

import (
	"github.com/morehao/golib/biz/grouter/ginrouter"
    "{{.ModulePath}}/{{.AppPathInProject}}/internal/controller/ctr{{.PackageName}}"
)
{{if .IsNewRouter}}
// {{.StructNameLowerCamel}}Router 初始化{{.Description}}路由信息
func {{.StructNameLowerCamel}}Router(groups *ginrouter.RouterGroups) {
	{{.StructNameLowerCamel}}Ctr := ctr{{.PackageName}}.New{{.StructName}}Ctr()

	groups.V1.{{.HttpMethod}}("/{{.StructNameLowerCamel}}/{{.FunctionNameLowerCamel}}", {{.StructNameLowerCamel}}Ctr.{{.FunctionName}})
}
{{end}}
