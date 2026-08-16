package router

import (
	"github.com/morehao/golib/biz/gserver/ginserver"
	"github.com/morehao/golib/biz/gserver/ginupload"
	"github.com/morehao/golib/filestore"
)

func RegisterRouter(groups *ginserver.RouterGroups, appName string) {
	formatRouter(groups)
	sseRouter(groups)
	clientRouter(groups)
	healthRouter(groups)
	fileRouter(groups)
	userRouter(groups)
}

// fileRouter 注册 golib ginupload 文件上传路由
func fileRouter(groups *ginserver.RouterGroups) {
	v1RouterGroup := groups.MustGetGroup(ginserver.ApiVersionV1)
	ginupload.Register(v1RouterGroup, filestore.Get())
}
