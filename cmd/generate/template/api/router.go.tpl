package router

import (
	"github.com/gin-gonic/gin"
    "{{.ModulePath}}/{{.AppPathInProject}}/internal/controller/ctr{{.PackageName}}"
)
{{if .IsNewRouter}}
// {{.StructNameLowerCamel}}Router 初始化{{.Description}}路由信息
func {{.StructNameLowerCamel}}Router(routerGroup *gin.RouterGroup) {
	{{.StructNameLowerCamel}}Ctr := ctr{{.PackageName}}.New{{.StructName}}Ctr()

	routerGroup.{{.HttpMethod}}("/{{.StructNameLowerCamel}}/{{.FunctionNameLowerCamel}}", {{.StructNameLowerCamel}}Ctr.{{.FunctionName}})
}
{{end}}
