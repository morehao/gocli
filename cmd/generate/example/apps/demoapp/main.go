package main

import (
	"github.com/gin-gonic/gin"
	"github.com/morehao/example/apps/demoapp/internal/router"
	"github.com/morehao/golib/biz/gconstant"
	"github.com/morehao/golib/biz/gserver/ginserver"
)

func main() {
	engine := gin.Default()

	groups := ginserver.NewRouterGroups(
		engine,
		"demoapp",
		ginserver.Version{Name: gconstant.ApiVersionV1},
	)

	router.RegisterRouter(groups)

	engine.Run(":8080")
}
