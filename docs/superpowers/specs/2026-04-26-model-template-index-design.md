# Model 模板索引支持设计

## 概述

为 `model.go.tpl` 模板增加索引支持，并修复空分号和字段注释标签问题。

## 问题

1. **空分号污染**：当 `NullableDesc`、`DefaultValue`、`GormComment` 为空时会产生 `;;`
2. **字段注释标签错误**：使用 `comment` 而非 GORM 的 `columnComment`
3. **缺少索引支持**：无法生成 `index`/`uniqueIndex` 标签

## 修改内容

### 1. golib/codegen/db.go

新增索引结构体：

```go
type mysqlIndexInfo struct {
    IndexName  string
    ColumnName string
    NonUnique  int
    SeqInIndex int
}

type postgresqlIndexInfo struct {
    IndexName  string
    ColumnName string
    IsUnique   bool
    IsPrimary  bool
    SeqInIndex int
}
```

`ModelField` 新增字段：

```go
type ModelField struct {
    // ... 现有字段 ...
    IndexName     string
    IsUniqueIndex bool
}
```

### 2. golib/codegen/mysql.go

新增 `getIndexInfo` 方法，获取 `INFORMATION_SCHEMA.STATISTICS` 中的索引信息。

### 3. golib/codegen/postgresql.go

新增 `getIndexInfo` 方法，通过 `pg_index` 等系统表获取索引信息。

### 4. gocli/cmd/generate/gen_model.go

- `ModelField` 结构体新增 `IndexName`、`IsUniqueIndex` 字段
- `genModel` 函数中填充索引相关字段

### 5. cmd/generate/template/model/model.go.tpl

**修复后第18行：**

```go
`gorm:"column:{{.ColumnName}};type:{{.ColumnType}};{{.NullableDesc}};{{.DefaultValue}};{{if .IndexName}}index:{{.IndexName}}{{if .IsUniqueIndex}},uniqueIndex{{end}}{{end}};{{if .Comment}}columnComment:{{.Comment}}{{end}}"`
```

**修复点：**
- `{{.NullableDesc}}` 和 `{{.DefaultValue}}` 空值时不输出
- `comment` → `columnComment`
- 新增 `index:xxx` 和 `uniqueIndex` 标签支持

## 生成示例

**单字段索引：**
```go
Name string `gorm:"column:name;type:varchar(64);not null;default '';index:idx_name;columnComment:名称"`
```

**复合索引：**
```go
ColA string `gorm:"column:col_a;type:varchar(64);not null;default '';index:idx_abc;columnComment:A列"`
ColB string `gorm:"column:col_b;type:varchar(64);not null;default '';index:idx_abc;columnComment:B列"`
ColC string `gorm:"column:col_c;type:varchar(64);not null;default '';index:idx_abc;columnComment:C列"`
```

**唯一索引：**
```go
Code string `gorm:"column:code;type:varchar(32);not null;default '';index:uni_code,uniqueIndex;columnComment:编码"`
```

## 实现顺序

1. `golib/codegen/db.go` - 新增结构体
2. `golib/codegen/mysql.go` - 添加 getIndexInfo 方法
3. `golib/codegen/postgresql.go` - 添加 getIndexInfo 方法
4. `gocli/cmd/generate/gen_model.go` - 更新 ModelField 和构造逻辑
5. `cmd/generate/template/model/model.go.tpl` - 修复模板