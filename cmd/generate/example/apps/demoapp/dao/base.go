package dao

import (
	"github.com/gin-gonic/gin"
	"github.com/morehao/example/pkg/dbclient"
	"gorm.io/gorm"
)

type Base struct {
	Tx *gorm.DB
}

// DB 获取DB
func (base *Base) DB(ctx *gin.Context) (db *gorm.DB) {
	if base.Tx != nil {
		return base.Tx.WithContext(ctx)
	}

	db = dbclient.DemoDBGetter(ctx)
	return
}
