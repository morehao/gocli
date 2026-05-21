package dbclient

import (
	"context"
	"fmt"
	"sync"

	"github.com/morehao/golib/biz/gormplugin"
	"github.com/morehao/golib/dbaccess/dbgorm"
	"github.com/morehao/golib/glog"
	"gorm.io/gorm"
)

var (
	dbMap   = make(map[string]*gorm.DB)
	dbMutex sync.RWMutex
)

const (
	dbNameDemo = "demo"
	dbNameIam  = "ark_iam"
)

func InitMultiDB(configs []dbgorm.GormConfig, logConfig *glog.LogConfig) error {
	if len(configs) == 0 {
		return fmt.Errorf("mysql config is empty")
	}

	tenantPlugin := gormplugin.New()

	var opts []dbgorm.Option
	if logConfig != nil {
		opts = append(opts, dbgorm.WithLogConfig(logConfig))
	}
	for _, cfg := range configs {
		client, err := dbgorm.New(&cfg, opts...)
		if err != nil {
			return fmt.Errorf("init mysql failed: %v", err)
		}
		if err := client.Use(tenantPlugin); err != nil {
			return fmt.Errorf("register tenant plugin failed: %v", err)
		}
		dbMutex.Lock()
		dbMap[cfg.Service] = client
		dbMutex.Unlock()
	}
	return nil
}

func GetDB(ctx context.Context, dbName string) *gorm.DB {
	dbMutex.RLock()
	defer dbMutex.RUnlock()
	return dbMap[dbName].WithContext(ctx)
}

func IamDB(ctx context.Context) *gorm.DB {
	return GetDB(ctx, dbNameIam)
}

func DemoDB(ctx context.Context) *gorm.DB {
	return GetDB(ctx, dbNameDemo)
}
