package ctruser

import (
	"github.com/gin-gonic/gin"
	"github.com/morehao/go-ark-template/demo/internal/dto/dtouser"
	"github.com/morehao/go-ark-template/demo/internal/service/svcuser"
	"github.com/morehao/golib/biz/gcontext/gincontext"
)

type UserCtr interface {
	Create(ctx *gin.Context)
	Delete(ctx *gin.Context)
	Update(ctx *gin.Context)
	Detail(ctx *gin.Context)
	PageList(ctx *gin.Context)
}

type userCtr struct {
	userSvc svcuser.UserSvc
}

var _ UserCtr = (*userCtr)(nil)

func NewUserCtr() UserCtr {
	return &userCtr{
		userSvc: svcuser.NewUserSvc(),
	}
}

// Create 创建用户管理
// @Tags 用户管理
// @Summary 创建用户管理
// @accept application/json
// @Produce application/json
// @Param req body dtouser.UserCreateReq true "创建用户管理"
// @Success 200 {object} gincontext.DtoRender{data=dtouser.UserCreateResp} "{"code": 0, "requestID": "xxx", "data": "ok", "msg": "success"}"
// @Router /v1/demo/users [post]
func (ctr *userCtr) Create(ctx *gin.Context) {
	var req dtouser.UserCreateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		gincontext.Fail(ctx, err)
		return
	}
	res, err := ctr.userSvc.Create(ctx, &req)
	if err != nil {
		gincontext.Fail(ctx, err)
		return
	}
	gincontext.Success(ctx, res)
}

// Delete 删除用户管理
// @Tags 用户管理
// @Summary 删除用户管理
// @accept application/json
// @Produce application/json
// @Param userID path int true "userID"
// @Success 200 {object} gincontext.DtoRender{data=string} "{"code": 0, "requestID": "xxx", "data": "ok", "msg": "success"}"
// @Router /v1/demo/users/{userID} [delete]
func (ctr *userCtr) Delete(ctx *gin.Context) {
	var req dtouser.UserDeleteReq
	if err := gincontext.BindPathParams(ctx, &req); err != nil {
		gincontext.Fail(ctx, err)
		return
	}

	if err := ctr.userSvc.Delete(ctx, &req); err != nil {
		gincontext.Fail(ctx, err)
		return
	}
	gincontext.Success(ctx, "删除成功")
}

// Update 修改用户管理
// @Tags 用户管理
// @Summary 修改用户管理
// @accept application/json
// @Produce application/json
// @Param userID path int true "userID"
// @Param req body dtouser.UserUpdateReq true "修改用户管理"
// @Success 200 {object} gincontext.DtoRender{data=string} "{"code": 0, "requestID": "xxx", "data": "ok", "msg": "修改成功"}"
// @Router /v1/demo/users/{userID} [put]
func (ctr *userCtr) Update(ctx *gin.Context) {
	var req dtouser.UserUpdateReq
	if err := gincontext.BindPathParams(ctx, &req); err != nil {
		gincontext.Fail(ctx, err)
		return
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		gincontext.Fail(ctx, err)
		return
	}
	if err := ctr.userSvc.Update(ctx, &req); err != nil {
		gincontext.Fail(ctx, err)
		return
	}
	gincontext.Success(ctx, "修改成功")
}

// Detail 用户管理详情
// @Tags 用户管理
// @Summary 用户管理详情
// @accept application/json
// @Produce application/json
// @Param userID path int true "userID"
// @Success 200 {object} gincontext.DtoRender{data=dtouser.UserDetailResp} "{"code": 0, "requestID": "xxx", "data": "ok", "msg": "success"}"
// @Router /v1/demo/users/{userID} [get]
func (ctr *userCtr) Detail(ctx *gin.Context) {
	var req dtouser.UserDetailReq
	if err := gincontext.BindPathParams(ctx, &req); err != nil {
		gincontext.Fail(ctx, err)
		return
	}
	res, err := ctr.userSvc.Detail(ctx, &req)
	if err != nil {
		gincontext.Fail(ctx, err)
		return
	}
	gincontext.Success(ctx, res)
}

// PageList 用户管理列表
// @Tags 用户管理
// @Summary 用户管理列表分页
// @accept application/json
// @Produce application/json
// @Param req query dtouser.UserPageListReq true "用户管理列表"
// @Success 200 {object} gincontext.DtoRender{data=dtouser.UserPageListResp} "{"code": 0, "requestID": "xxx", "data": "ok", "msg": "success"}"
// @Router /v1/demo/users [get]
func (ctr *userCtr) PageList(ctx *gin.Context) {
	var req dtouser.UserPageListReq
	if err := ctx.ShouldBindQuery(&req); err != nil {
		gincontext.Fail(ctx, err)
		return
	}
	res, err := ctr.userSvc.PageList(ctx, &req)
	if err != nil {
		gincontext.Fail(ctx, err)
		return
	}
	gincontext.Success(ctx, res)
}
