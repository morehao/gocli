package main

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/morehao/go-ark-template/demo"
	"github.com/morehao/go-ark-template/demo/config"
	"github.com/morehao/golib/glog"
)

func main() {
	if err := serverInit(); err != nil {
		panic(fmt.Sprintf("server init failed, error: %v", err))
	}
	if config.Conf.Server.Env == "prod" {
		gin.SetMode(gin.ReleaseMode)
	}
	defer shutdownTraceProvider()
	defer func() {
		if err := glog.Close(); err != nil {
			fmt.Printf("failed to close logger: %v\n", err)
		}
	}()

	engine := gin.New()
	engine.Use(gin.Recovery())
	demo.Routers(engine)

	if err := engine.Run(fmt.Sprintf(":%s", config.Conf.Server.Port)); err != nil {
		glog.Errorf(context.Background(), "%s run fail, port:%s", demo.AppName, config.Conf.Server.Port)
		panic(err)
	} else {
		glog.Infof(context.Background(), "%s run success, port:%s", demo.AppName, config.Conf.Server.Port)
	}
}
