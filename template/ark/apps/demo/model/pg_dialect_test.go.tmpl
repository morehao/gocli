package model

import (
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm/schema"
)

// postgresModels 是受本用例守护的模型清单，新增 model 时补进来。
var postgresModels = []any{
	&UserEntity{},
}

// nonPostgresTypeTokens 是禁止出现在 gorm `type:` tag 里的其它方言字面量。
//
// 命中的原因：gorm 只把 bool/int/uint/float/string/time/bytes 这类**逻辑类型名**交给方言映射，
// 其它取值会被原样拼进 DDL。例如 `type:bigint unsigned` 在 PostgreSQL 上实测直接
// `syntax error at or near "unsigned" (SQLSTATE 42601)`（PG 的等价写法是 type:bigint）。
// 显式写 PG 类型是允许的（type:bigint、type:varchar(255)、type:jsonb 均可），与
// backend/scripts/sql/table_init.sql 保持同款即可。
var nonPostgresTypeTokens = []string{
	"unsigned", "zerofill", "auto_increment",
	"tinyint", "mediumint", "tinytext", "mediumtext", "longtext",
	"tinyblob", "mediumblob", "longblob", "varbinary",
	"datetime", "enum(", "set(", "int(", "engine=",
}

// TestModelTagsArePostgresCompatible 守住「本仓库只适配 PostgreSQL」这条约束：
// model 的 gorm tag 不得携带其它方言的类型字面量，否则建表 SQL 在 PG 上直接语法错误。
//
// 本用例不需要数据库：只读 tag，不做任何连接。
func TestModelTagsArePostgresCompatible(t *testing.T) {
	for _, dest := range postgresModels {
		s, err := schema.Parse(dest, &sync.Map{}, schema.NamingStrategy{})
		assert.NoError(t, err)
		if err != nil {
			continue
		}
		for _, dbName := range s.DBNames {
			field := s.FieldsByDBName[dbName]
			rawType := strings.ToLower(strings.TrimSpace(field.TagSettings["TYPE"]))
			for _, token := range nonPostgresTypeTokens {
				assert.NotContains(t, rawType, token,
					"表 %s 的列 %s 用了非 PostgreSQL 的 type:%s，请改成 PG 写法（如 type:bigint，详见 docs/agents/backend-conventions.md#模型字段类型规范postgresql）",
					s.Table, dbName, field.TagSettings["TYPE"])
			}
		}
	}
}
