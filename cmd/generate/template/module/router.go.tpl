package router

import (
	"github.com/morehao/golib/biz/grouter/ginrouter"
	"{{.ModulePath}}/{{.AppPathInProject}}/internal/controller/ctr{{.PackageName}}"
)

// {{.StructNameLowerCamel}}Router 初始化{{.Description}}路由信息
func {{.StructNameLowerCamel}}Router(groups *ginrouter.RouterGroups) {
	{{.StructNameLowerCamel}}Ctr := ctr{{.PackageName}}.New{{.StructName}}Ctr()

	groups.V1.POST("/{{.StructNameLowerCamel}}/create", {{.StructNameLowerCamel}}Ctr.Create)
	groups.V1.POST("/{{.StructNameLowerCamel}}/delete", {{.StructNameLowerCamel}}Ctr.Delete)
	groups.V1.POST("/{{.StructNameLowerCamel}}/update", {{.StructNameLowerCamel}}Ctr.Update)
	groups.V1.GET("/{{.StructNameLowerCamel}}/detail", {{.StructNameLowerCamel}}Ctr.Detail)
	groups.V1.POST("/{{.StructNameLowerCamel}}/pageList", {{.StructNameLowerCamel}}Ctr.PageList)
}
