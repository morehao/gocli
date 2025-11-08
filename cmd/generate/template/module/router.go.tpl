package router

import (
	"{{.ModulePath}}/{{.AppPathInProject}}/internal/controller/ctr{{.PackageName}}"

	"github.com/gin-gonic/gin"
)

// {{.PackageName}}Router 初始化{{.Description}}路由信息
func {{.StructNameLowerCamel}}Router(routerGroup *gin.RouterGroup) {
	{{.StructNameLowerCamel}}Ctr := ctr{{.PackageName}}.New{{.StructName}}Ctr()
	{{.StructNameLowerCamel}}Group := routerGroup.Group("{{.StructNameLowerCamel}}")
	{
		{{.StructNameLowerCamel}}Group.POST("create", {{.StructNameLowerCamel}}Ctr.Create)   // 新建{{.Description}}
		{{.StructNameLowerCamel}}Group.POST("delete", {{.StructNameLowerCamel}}Ctr.Delete)   // 删除{{.Description}}
		{{.StructNameLowerCamel}}Group.POST("update", {{.StructNameLowerCamel}}Ctr.Update)   // 更新{{.Description}}
		{{.StructNameLowerCamel}}Group.GET("detail", {{.StructNameLowerCamel}}Ctr.Detail)    // 根据ID获取{{.Description}}
        {{.StructNameLowerCamel}}Group.GET("pageList", {{.StructNameLowerCamel}}Ctr.PageList)  // 获取{{.Description}}列表
	}
}
