package router

import (
	"github.com/example/demoapp/internal/controller/ctruser"
	"github.com/morehao/golib/biz/gserver/ginserver"
)



// userRouter 初始化用户登录记录路由信息
func userRouter(groups *ginserver.RouterGroups) {
	userCtr := ctruser.NewUserCtr()

	v1RouterGroup := groups.MustGetGroup(ginserver.ApiVersionV1)

	v1RouterGroup.POST("/user/create", userCtr.Create)
	v1RouterGroup.POST("/user/delete", userCtr.Delete)
	v1RouterGroup.POST("/user/update", userCtr.Update)
	v1RouterGroup.GET("/user/detail", userCtr.Detail)
	v1RouterGroup.POST("/user/pageList", userCtr.PageList)
}
