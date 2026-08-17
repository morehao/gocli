package main

import (
	"context"
	"fmt"
	"time"

	"github.com/morehao/go-ark-template/demo"
	"github.com/morehao/go-ark-template/demo/config"
	"github.com/morehao/go-ark-template/pkg/dbclient"
	"github.com/morehao/golib/glog"
	_ "github.com/morehao/golib/glog/driver/zap"
	"github.com/morehao/golib/gtrace"
	"github.com/morehao/golib/gtrace/otlptracegrpc"
)

var traceProvider *gtrace.Provider

func serverInit() error {
	if err := preInit(); err != nil {
		return err
	}
	if err := initTrace(); err != nil {
		return err
	}
	if err := resourceInit(); err != nil {
		return err
	}
	return nil
}

func preInit() error {
	config.InitConf()
	defaultLogCfg := config.Conf.Log["default"]
	if err := glog.InitLogger(&defaultLogCfg); err != nil {
		return fmt.Errorf("init logger failed: %w", err)
	}
	return nil
}

func resourceInit() error {
	var gormLogConfig *glog.LogConfig
	if cfg, ok := config.Conf.Log["gorm"]; ok {
		gormLogConfig = &cfg
	}
	if err := dbclient.InitMultiDB(config.Conf.DBConfigs, gormLogConfig); err != nil {
		return fmt.Errorf("init db failed: %w", err)
	}
	var redisLogConfig *glog.LogConfig
	if cfg, ok := config.Conf.Log["redis"]; ok {
		redisLogConfig = &cfg
	}
	if err := dbclient.InitRedis(config.Conf.RedisConfig, redisLogConfig); err != nil {
		return fmt.Errorf("init redis failed: %w", err)
	}
	var esLogConfig *glog.LogConfig
	if cfg, ok := config.Conf.Log["es"]; ok {
		esLogConfig = &cfg
	}
	if err := dbclient.InitMultiEs(config.Conf.ESConfigs, esLogConfig); err != nil {
		return fmt.Errorf("init es failed: %w", err)
	}
	return nil
}

func initTrace() error {
	provider, err := otlptracegrpc.NewGRPCProvider(context.Background(), demo.AppName, config.Conf.Server.Env, config.Conf.Trace)
	if err != nil {
		glog.Errorf(context.Background(), "[%s.initTrace] init trace failed, fallback to disabled mode, err:%v", demo.AppName, err)
		return nil
	}
	traceProvider = provider
	return nil
}

func shutdownTraceProvider() {
	if traceProvider == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := traceProvider.Shutdown(ctx); err != nil {
		glog.Errorf(context.Background(), "[%s.shutdownTraceProvider] shutdown fail, err:%v", demo.AppName, err)
	}
}
