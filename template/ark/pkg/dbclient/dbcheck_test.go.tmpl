package dbclient_test

import (
	"context"
	"testing"
	"time"

	"github.com/morehao/go-ark-template/pkg/dbclient"
	"github.com/morehao/go-ark-template/pkg/testsetup"
)

const (
	dbConnTimeout = 5 * time.Second
)

// TestDbcheckVerify 验证 dbclient 对 MySQL/Redis/ES/PG 的调用及调用日志是否生效。
//
// 依赖 testsetup.Initialize 初始化 dbclient 与 glog（读取 apps/demo/config/config.yaml，
// gorm/redis/es 日志级别为 debug 且输出到 console）。
//
// 运行方式：
//
//	go test ./pkg/dbclient/ -v -run TestDbcheckVerify
//
// 注意：本地需就绪 MySQL(3306)/Redis(6379)/ES(9200)。
func TestDbcheckVerify(t *testing.T) {
	testsetup.Initialize(testsetup.AppNameDemo)
	defer testsetup.Close(testsetup.AppNameDemo)

	ctx, cancel := context.WithTimeout(context.Background(), dbConnTimeout)
	defer cancel()

	t.Run("MySQL", func(t *testing.T) {
		db := dbclient.DemoDB(ctx)
		if db == nil {
			t.Fatal("DemoDB is nil")
		}
		var val int
		if err := db.Raw("SELECT 1").Scan(&val).Error; err != nil {
			t.Fatalf("select 1 failed: %v", err)
		}
		t.Logf("mysql select 1 => %d", val)
	})

	t.Run("Redis", func(t *testing.T) {
		if dbclient.RedisCli == nil {
			t.Fatal("RedisCli is nil")
		}
		if err := dbclient.RedisCli.Set(ctx, "demo:dbcheck", "ok", time.Minute).Err(); err != nil {
			t.Fatalf("redis set failed: %v", err)
		}
		val, err := dbclient.RedisCli.Get(ctx, "demo:dbcheck").Result()
		if err != nil {
			t.Fatalf("redis get failed: %v", err)
		}
		t.Logf("redis set/get => %s", val)
		dbclient.RedisCli.Del(ctx, "demo:dbcheck")
	})

	t.Run("Elasticsearch", func(t *testing.T) {
		if dbclient.DemoES == nil {
			t.Fatal("DemoES is nil")
		}
		res, err := dbclient.DemoES.Info()
		if err != nil {
			t.Fatalf("es info failed: %v", err)
		}
		defer res.Body.Close()
		t.Logf("es info status => %d", res.StatusCode)
	})
}
