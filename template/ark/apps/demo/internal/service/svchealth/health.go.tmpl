package svchealth

import (
	"context"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/morehao/go-ark-template/demo/internal/dto/dtohealth"
	"github.com/morehao/go-ark-template/pkg/dbclient"
	"github.com/morehao/golib/glog"
	"gorm.io/gorm"
)

type HealthSvc interface {
	Check(ctx *gin.Context) *dtohealth.DbCheckResp
}

type healthSvc struct {
}

var _ HealthSvc = (*healthSvc)(nil)

func NewHealthSvc() HealthSvc {
	return &healthSvc{}
}

func (svc *healthSvc) Check(ctx *gin.Context) *dtohealth.DbCheckResp {
	resp := &dtohealth.DbCheckResp{
		MySQL: "ok",
		Redis: "ok",
		ES:    "ok",
	}
	checkCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := pingDB(checkCtx, dbclient.DemoDB(checkCtx)); err != nil {
		glog.Errorf(ctx, "[svchealth.Check] mysql ping fail, err:%v", err)
		resp.MySQL = err.Error()
	}
	if err := pingRedis(checkCtx); err != nil {
		glog.Errorf(ctx, "[svchealth.Check] redis ping fail, err:%v", err)
		resp.Redis = err.Error()
	}
	if err := pingES(checkCtx); err != nil {
		glog.Errorf(ctx, "[svchealth.Check] es ping fail, err:%v", err)
		resp.ES = err.Error()
	}
	glog.Infof(ctx, "[svchealth.Check] result, mysql:%s, redis:%s, es:%s",
		resp.MySQL, resp.Redis, resp.ES)
	return resp
}

func pingDB(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return context.DeadlineExceeded
	}
	var val int
	return db.WithContext(ctx).Raw("SELECT 1").Scan(&val).Error
}

func pingRedis(ctx context.Context) error {
	if dbclient.RedisCli == nil {
		return context.DeadlineExceeded
	}
	return dbclient.RedisCli.Ping(ctx).Err()
}

func pingES(ctx context.Context) error {
	if dbclient.DemoES == nil {
		return context.DeadlineExceeded
	}
	res, err := dbclient.DemoES.Info(dbclient.DemoES.Info.WithContext(ctx))
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.IsError() {
		return context.DeadlineExceeded
	}
	return nil
}
