package dbclient

import (
	"context"
	"fmt"
	"sync"

	"github.com/morehao/golib/dbaccess/dbgorm"
	_ "github.com/morehao/golib/dbaccess/dbgorm/driver/mysql"
	"github.com/morehao/golib/dbaccess/gormplugin"
	"github.com/morehao/golib/glog"
	"gorm.io/gorm"
)

var (
	dbMap   = make(map[string]*gorm.DB)
	dbMutex sync.RWMutex

	// companyIDKey 是注入公司（租户）ID 的 context key，
	// 配合 gormplugin 的 company_id 过滤条件使用。
	companyIDKey = contextKey("company_id")
)

type contextKey string

const (
	dbNameDemo = "demo"
	dbNameIam  = "ark_iam"
)

// WithCompanyID 将公司（租户）ID 注入 context，
// 设置了之后，后续基于该 context 的查询/更新/删除会自动追加 company_id 过滤条件。
func WithCompanyID(ctx context.Context, companyID uint) context.Context {
	return context.WithValue(ctx, companyIDKey, companyID)
}

// companyIDFromCtx 从 context 中提取公司（租户）ID，
// 未设置或为 0 时返回 (nil, false)，此时 gormplugin 不注入过滤条件。
func companyIDFromCtx(ctx context.Context) (any, bool) {
	companyID, ok := ctx.Value(companyIDKey).(uint)
	if !ok || companyID == 0 {
		return nil, false
	}
	return companyID, true
}

func InitMultiDB(configs []dbgorm.Config, logConfig *glog.LogConfig) error {
	if len(configs) == 0 {
		return fmt.Errorf("mysql config is empty")
	}

	tenantPlugin, err := gormplugin.New(&gormplugin.ScopeConfig{
		FieldName:   "company_id",
		ExtractFunc: companyIDFromCtx,
	})
	if err != nil {
		return fmt.Errorf("init tenant plugin failed: %v", err)
	}

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
