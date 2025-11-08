package router

import (
    "{{.ModulePath}}/{{.AppPathInProject}}/internal/controller/ctr{{.PackageName}}"

	"github.com/gin-gonic/gin"
)
{{if .IsNewRouter}}
// {{.StructNameLowerCamel}}Router 初始化{{.Description}}路由信息
func {{.StructNameLowerCamel}}Router(routerGroup *gin.RouterGroup) {
	{{.StructNameLowerCamel}}Ctr := ctr{{.PackageName}}.New{{.StructName}}Ctr()
	{{.StructNameLowerCamel}}Group := routerGroup.Group("{{.StructNameLowerCamel}}")
	{
		{{.StructNameLowerCamel}}Group.{{.HttpMethod}}("{{.FunctionNameLowerCamel}}", {{.StructNameLowerCamel}}Ctr.{{.FunctionName}})   // {{.Description}}
	}
}
{{end}}
