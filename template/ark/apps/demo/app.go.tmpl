package demo

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/morehao/go-ark-template/demo/config"
	_ "github.com/morehao/go-ark-template/demo/docs"
	"github.com/morehao/go-ark-template/demo/internal/middleware"
	"github.com/morehao/go-ark-template/demo/internal/router"
	"github.com/morehao/go-ark-template/pkg/dbclient"
	"github.com/morehao/golib/biz/gserver/gindocs"
	"github.com/morehao/golib/biz/gserver/ginserver"
	"github.com/morehao/golib/filestore"
	"github.com/morehao/golib/glog"
	"github.com/morehao/golib/storage"
	_ "github.com/morehao/golib/storage/driver/local"
)

const AppName = "demo"

func Routers(engine *gin.Engine) {
	routerGroups := ginserver.NewRouterGroups(engine, AppName, ginserver.VersionGroup{
		Version: ginserver.ApiVersionV1,
		Middlewares: []gin.HandlerFunc{
			middleware.Example(),
		},
	})

	if config.Conf.Server.Env == "dev" {
		gindocs.Register(engine.Group("/"+AppName), AppName)
	}

	if err := initFileStore(); err != nil {
		panic(fmt.Errorf("demo.Routers: init file store failed: %w", err))
	}

	router.RegisterRouter(routerGroups, AppName)
}

// initFileStore 根据配置初始化文件上传存储（storage 驱动 + filestore 单例）。
func initFileStore() error {
	cfg := config.Conf.FileStorage

	st, err := storage.New(cfg.Driver, cfg.Storage)
	if err != nil {
		return fmt.Errorf("storage.New: %w", err)
	}

	var storeOpts []filestore.StoreOption
	if cfg.Storage.SignSecret != "" {
		storeOpts = append(storeOpts, filestore.WithSignSecret(cfg.Storage.SignSecret))
	}

	filestore.Init(dbclient.DemoDB(context.TODO()), st, cfg.Bucket, storeOpts...)
	glog.Infof(context.TODO(), "[demo.initFileStore] filestore init done, driver=%s, bucket=%s, local=%v", cfg.Driver, cfg.Bucket, filestore.IsLocal())
	return nil
}
