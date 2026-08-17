package ctrexample

import (
	"github.com/gin-gonic/gin"
	"github.com/morehao/go-ark-template/demo/client/httpbingo"
	"github.com/morehao/golib/biz/gcontext/gincontext"
	"github.com/morehao/golib/glog"
)

type ClientCtr interface {
	CallGetHttpbingo(ctx *gin.Context)
}

type clientCtr struct {
}

var _ ClientCtr = (*clientCtr)(nil)

func NewClientCtr() ClientCtr {
	return &clientCtr{}
}

func (ctr *clientCtr) CallGetHttpbingo(ctx *gin.Context) {
	getReq := httpbingo.GetRequest{
		ID: 1,
	}
	res, err := httpbingo.Get(ctx, &getReq)
	if err != nil {
		glog.Errorf(ctx, "[CallGetHttpbingo] get failed, err: %v", err)
		gincontext.Fail(ctx, err)
		return
	}
	gincontext.Success(ctx, res)
}
