package svcuser

import (
	"github.com/gin-gonic/gin"
	"github.com/example/demoapp/dao"
	"github.com/example/demoapp/internal/dto/dtouser"
	"github.com/example/demoapp/model"
	"github.com/example/demoapp/object/objuser"
	"github.com/example/pkg/code"
	"github.com/morehao/golib/biz/gcontext/gincontext"
	"github.com/morehao/golib/biz/genericdao"
	"github.com/morehao/golib/biz/gobject"
	"github.com/morehao/golib/glog"
	"github.com/morehao/golib/gutil"
)

type UserSvc interface {
	Create(ctx *gin.Context, req *dtouser.UserCreateReq) (*dtouser.UserCreateResp, error)
	Delete(ctx *gin.Context, req *dtouser.UserDeleteReq) error
	Update(ctx *gin.Context, req *dtouser.UserUpdateReq) error
	Detail(ctx *gin.Context, req *dtouser.UserDetailReq) (*dtouser.UserDetailResp, error)
	PageList(ctx *gin.Context, req *dtouser.UserPageListReq) (*dtouser.UserPageListResp, error)
}

type userSvc struct {
}

var _ UserSvc = (*userSvc)(nil)

func NewUserSvc() UserSvc {
	return &userSvc{}
}

// Create 创建用户登录记录
func (svc *userSvc) Create(ctx *gin.Context, req *dtouser.UserCreateReq) (*dtouser.UserCreateResp, error) {
	insertEntity := &model.UserEntity{
		CompanyID:    req.CompanyID,
		DepartmentID: req.DepartmentID,
		Name:         req.Name,
	}

	if err := dao.NewUserDao().Insert(ctx, insertEntity); err != nil {
		glog.Errorf(ctx, "[svcuser.UserCreate] dao Create fail, err:%v, req:%s", err, gutil.ToJsonString(req))
		return nil, code.GetError(code.UserCreateError)
	}
	return &dtouser.UserCreateResp{
		UserID: insertEntity.ID,
	}, nil
}

// Delete 删除用户登录记录
func (svc *userSvc) Delete(ctx *gin.Context, req *dtouser.UserDeleteReq) error {
	userEntity, err := dao.NewUserDao().GetByID(ctx, req.UserID)
	if err != nil {
		glog.Errorf(ctx, "[svcuser.UserDelete] dao GetByID fail, err:%v, req:%s", err, gutil.ToJsonString(req))
		return code.GetError(code.UserDeleteError)
	}
	if userEntity == nil || userEntity.ID == 0 {
		return code.GetError(code.UserNotExistError)
	}

	userID := gincontext.GetUserID(ctx)

	if err := dao.NewUserDao().Delete(ctx, req.UserID, userID); err != nil {
		glog.Errorf(ctx, "[svcuser.Delete] dao Delete fail, err:%v, req:%s", err, gutil.ToJsonString(req))
		return code.GetError(code.UserDeleteError)
	}
	return nil
}

// Update 更新用户登录记录
func (svc *userSvc) Update(ctx *gin.Context, req *dtouser.UserUpdateReq) error {
	userEntity, err := dao.NewUserDao().GetByID(ctx, req.UserID)
	if err != nil {
		glog.Errorf(ctx, "[svcuser.UserUpdate] dao GetByID fail, err:%v, req:%s", err, gutil.ToJsonString(req))
		return code.GetError(code.UserUpdateError)
	}
	if userEntity == nil || userEntity.ID == 0 {
		return code.GetError(code.UserNotExistError)
	}

	updateMap := map[string]any{}
	if err := dao.NewUserDao().UpdateMap(ctx, req.UserID, updateMap); err != nil {
		glog.Errorf(ctx, "[svcuser.UserUpdate] dao UpdateMap fail, err:%v, req:%s", err, gutil.ToJsonString(req))
		return code.GetError(code.UserUpdateError)
	}
	return nil
}

// Detail 根据id获取用户登录记录
func (svc *userSvc) Detail(ctx *gin.Context, req *dtouser.UserDetailReq) (*dtouser.UserDetailResp, error) {
	userEntity, err := dao.NewUserDao().GetByID(ctx, req.UserID)
	if err != nil {
		glog.Errorf(ctx, "[svcuser.UserDetail] dao GetByID fail, err:%v, req:%s", err, gutil.ToJsonString(req))
		return nil, code.GetError(code.UserGetDetailError)
	}
	if userEntity == nil || userEntity.ID == 0 {
		return nil, code.GetError(code.UserNotExistError)
	}
	resp := &dtouser.UserDetailResp{
		UserID: userEntity.ID,
		UserBaseInfo: objuser.UserBaseInfo{
			CompanyID:    userEntity.CompanyID,
			DepartmentID: userEntity.DepartmentID,
			Name:         userEntity.Name,
		},
		OperatorBaseInfo: gobject.OperatorBaseInfo{
			CreatedAt: userEntity.CreatedAt.Unix(),
			UpdatedAt: userEntity.UpdatedAt.Unix(),
		},
	}
	return resp, nil
}

// PageList 分页获取用户登录记录列表
func (svc *userSvc) PageList(ctx *gin.Context, req *dtouser.UserPageListReq) (*dtouser.UserPageListResp, error) {
	cond := &dao.UserCond{
		BaseCond: &genericdao.BaseCond{
			Page:     req.Page,
			PageSize: req.PageSize,
		},
	}
	userEntityList, total, err := dao.NewUserDao().GetPageListByCond(ctx, cond)
	if err != nil {
		glog.Errorf(ctx, "[svcuser.UserPageList] dao GetPageListByCond fail, err:%v, req:%s", err, gutil.ToJsonString(req))
		return nil, code.GetError(code.UserGetPageListError)
	}
	list := make([]dtouser.UserPageListItem, 0, len(userEntityList))
	for _, v := range userEntityList {
		list = append(list, dtouser.UserPageListItem{
			UserID: v.ID,
			UserBaseInfo: objuser.UserBaseInfo{
				CompanyID:    v.CompanyID,
				DepartmentID: v.DepartmentID,
				Name:         v.Name,
			},
			OperatorBaseInfo: gobject.OperatorBaseInfo{
				UpdatedAt: v.UpdatedAt.Unix(),
			},
		})
	}
	return &dtouser.UserPageListResp{
		List:  list,
		Total: total,
	}, nil
}
