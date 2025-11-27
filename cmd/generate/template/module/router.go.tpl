package router

import (
	"github.com/gin-gonic/gin"
	"{{.ModulePath}}/{{.AppPathInProject}}/internal/controller/ctr{{.PackageName}}"
)

// {{.StructNameLowerCamel}}Router 初始化{{.Description}}路由信息
func {{.StructNameLowerCamel}}Router(routerGroup *gin.RouterGroup) {
	{{.StructNameLowerCamel}}Ctr := ctr{{.PackageName}}.New{{.StructName}}Ctr()

	routerGroup.POST("/{{.StructNameLowerCamel}}/create", {{.StructNameLowerCamel}}Ctr.Create)    
	routerGroup.POST("/{{.StructNameLowerCamel}}/delete", {{.StructNameLowerCamel}}Ctr.Delete)    
	routerGroup.POST("/{{.StructNameLowerCamel}}/update", {{.StructNameLowerCamel}}Ctr.Update)    
	routerGroup.GET("/{{.StructNameLowerCamel}}/detail", {{.StructNameLowerCamel}}Ctr.Detail)     
	routerGroup.POST("/{{.StructNameLowerCamel}}/pageList", {{.StructNameLowerCamel}}Ctr.PageList) 
}
