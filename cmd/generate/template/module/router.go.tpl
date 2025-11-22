package router

import (
	"{{.ModulePath}}/{{.AppPathInProject}}/internal/controller/ctr{{.PackageName}}"

	"github.com/gin-gonic/gin"
)

// {{.StructNameLowerCamel}}Router 初始化{{.Description}}路由信息
func {{.StructNameLowerCamel}}Router(routerGroup *gin.RouterGroup) {
	{{.StructNameLowerCamel}}Ctr := ctr{{.PackageName}}.New{{.StructName}}Ctr()

	routerGroup.POST("/{{.StructNameLowerCamel}}/create", {{.StructNameLowerCamel}}Ctr.Create)    // 新建{{.Description}}
	routerGroup.POST("/{{.StructNameLowerCamel}}/delete", {{.StructNameLowerCamel}}Ctr.Delete)    // 删除{{.Description}}
	routerGroup.POST("/{{.StructNameLowerCamel}}/update", {{.StructNameLowerCamel}}Ctr.Update)    // 更新{{.Description}}
	routerGroup.GET("/{{.StructNameLowerCamel}}/detail", {{.StructNameLowerCamel}}Ctr.Detail)     // 根据ID获取{{.Description}}
	routerGroup.GET("/{{.StructNameLowerCamel}}/pageList", {{.StructNameLowerCamel}}Ctr.PageList) // 获取{{.Description}}列表
}
