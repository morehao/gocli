# 后端 Monorepo 项目

本目录由 `gocli create project` 生成，是一个基于 Go workspace 的后端 monorepo 骨架。

## 目录结构

```
my-backend/
├── go.work                  # Go workspace：apps/demoapp + pkg
├── go.work.sum
├── apps/                    # 应用目录（每个 app 独立模块）
│   └── demoapp/             # 示例应用（gin + gorm + golib）
│       ├── config/
│       │   └── code_gen.yaml     # 代码生成配置文件
│       ├── go.mod
│       ├── main.go
│       ├── model/           # 数据库模型
│       ├── dao/             # 数据访问层
│       ├── object/          # 业务对象
│       └── internal/        # controller / service / dto / router
└── pkg/                     # 公共库（独立模块）
    ├── go.mod
    ├── code/                # 错误码
    └── dbclient/            # 数据库客户端
```

## 常用命令

```bash
# 新增应用（自动注册进 go.work）
gocli create app -n userapp

# 为应用生成完整 CRUD 模块（基于数据库表结构）
gocli generate module -a demoapp

# 仅生成数据层（model + dao + object）
gocli generate model -a demoapp

# 为已有模块添加单个接口
gocli generate api -a demoapp
```

## 代码生成配置

在 `apps/<app>/config/code_gen.yaml` 中配置：

- **database_dsn**: 数据库连接字符串，格式：schema://dsn（支持 mysql 和 postgresql）
- **module / model / api**: 各模式的生成配置（包名、描述、表名等）

示例：

```yaml
database_dsn: mysql://root:123456@tcp(127.0.0.1:3306)/demo?charset=utf8mb4&parseTime=True&loc=Local
service_name: mysql
module:
  package_name: user
  description: 用户
  table_name: user
```

## 生成的文件位置

- **model/object**: 在 `apps/<app>/` 下生成（可通过 `layer_parent_dir_map` 配置）
- **dao**: 在 `apps/<app>/{appName}dao/` 下生成（如 `demoappdao`），使用 `gormdao.Dao` 封装
- **controller/service/dto**: 在 `apps/<app>/internal/` 下生成（可通过 `layer_parent_dir_map` 配置）
- **router**: 在 `apps/<app>/internal/router/` 下生成
- **code**: 在项目根目录的 `pkg/code/` 下生成

## 自定义生成行为

```yaml
# 自定义层级父目录
layer_parent_dir_map:
  controller: internal
  service: internal
  dto: internal

# 自定义层级名称（例如：model -> pgmodel）
layer_name_map:
  model: pgmodel

# 自定义文件名前缀（例如：model 文件前缀为 pg_）
layer_prefix_map:
  model: pg_
```

**注意**：dao 层会自动生成到 `{appName}dao` 目录下（如 `demoappdao`），包名也为 `{appName}dao`，无需额外配置。
