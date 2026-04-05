package router

import "github.com/morehao/golib/biz/grouter/ginrouter"

// RegisterRouter 注册路由
// 生成的路由函数会自动注册到这里
func RegisterRouter(groups *ginrouter.RouterGroups) {
	userLoginLogRouter(groups)
}
