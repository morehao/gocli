[简体中文](./README.zh.md) | [English](./README.md)


# gocli 介绍
`gocli` 是一个使用 Go 语言开发的命令行工具集合，旨在提升开发效率，目前包含**代码生成**和**快速新建项目**功能。

# 快速开始

## 安装

```bash
go install github.com/morehao/gocli@latest
```

## generate

`generate`是一个强大的代码生成工具，基于模板文件和数据库结构快速生成代码。项目结构和风格参照[goark](https://github.com/morehao/goark)。

### 功能特性

* 🚀 **快速开发**：基于 MySQL/PostgreSQL 表结构快速生成完整的 CRUD 模块
* 📦 **多层代码生成**：支持 model、dao、service、controller、dto、router 等多层代码
* 🎯 **三种生成模式**：module（完整模块）、model（仅数据层）、api（单个接口）
* 🔧 **高度可定制**：可配置层级名称、父级目录、文件名前缀
* ✨ **自动格式化**：生成的代码自动使用 `gofmt` 格式化
* 📖 **数据库驱动**：读取 MySQL/PostgreSQL 表结构生成准确的模型定义

### 生成模式

#### 1. **module** - 完整模块生成

生成包含所有层级的完整 CRUD 模块：
- **model**：数据库模型
- **dao**：数据访问对象
- **object**：业务对象
- **controller**：HTTP 请求处理器
- **service**：业务逻辑层
- **dto**：请求/响应对象
- **router**：路由注册
- **code**：错误码定义

**使用场景**：从零开始创建新功能模块

```bash
gocli generate -m module -a demoapp
```

#### 2. **model** - 数据层生成

仅生成数据层代码：
- **model**：带 GORM 标签的数据库模型
- **dao**：数据访问方法
- **object**：用于数据转换的业务对象

**使用场景**：只需添加数据库表，无需完整 CRUD 操作

```bash
gocli generate -m model -a demoapp
```

#### 3. **api** - 单个接口生成

为现有模块添加新的 API 接口：
- **controller**：新的控制器方法
- **service**：新的服务方法
- **dto**：请求/响应结构体
- **router**：路由注册

**使用场景**：为已有功能添加新的接口

```bash
gocli generate -m api -a demoapp
```

### 命令执行前提

1. **在项目根目录执行**：需在项目根目录下执行命令（例如 `go-gin-web` 目录）
2. **指定应用名称**：通过 `--app` 参数指定要生成代码的应用名称（如 `demoapp`）
3. **配置文件必需**：确保 `apps/{appName}/config/code_gen.yaml` 文件存在

示例配置文件：
```yaml
database_dsn: mysql://root:123456@tcp(127.0.0.1:3306)/demo?charset=utf8mb4&parseTime=True&loc=Local
service_name: mysql
module:
  package_name: user
  description: 用户登录记录
  table_name: user_login_log
  table_prefix: ""   # 可选：表名前缀，生成结构体名时会去除此前缀（如 "iam_"）
model:
  package_name: user
  description: 用户
  table_name: user
  table_prefix: ""   # 可选：表名前缀
api:
  package_name: user
  target_filename: user_login_log.go
  function_name: Delete
  http_method: POST
  description: 删除登录记录
  api_doc_tag: 用户登录记录
```

**数据库连接格式说明：**

| 数据库类型 | DSN 格式 |
|-----------|---------|
| MySQL | `mysql://user:password@tcp(host:port)/dbname?params` |
| PostgreSQL | `postgres1l://user:password@host:port/dbname?params` |

**示例：**
```yaml
# MySQL
database_dsn: mysql://root:123456@tcp(127.0.0.1:3306)/demo?charset=utf8mb4&parseTime=True&loc=Local

# PostgreSQL
database_dsn: postgresql://postgres:password@localhost:5432/demo?sslmode=disable
```
### 配置说明

#### 全局配置

| 配置项 | 说明 | 示例值 | 是否必填 |
| ----- | ---- | ------ | ------- |
| `database_dsn` | 数据库连接字符串，格式：schema://dsn | `mysql://root:123456@tcp(127.0.0.1:3306)/demo?charset=utf8mb4&parseTime=True&loc=Local` | ✅ 必填 |
| `service_name` | model/dao 层目录名称前缀及数据库连接名 | `mysql` | ✅ 必填 |



| 配置项 | 说明 | 示例值 | 是否必填 |
| ----- | ---- | ------ | ------- |
| `package_name` | 模块包名 | `user` | ✅ 必填 |
| `description` | 模块描述（用于注释） | `用户登录记录` | ✅ 必填 |
| `table_name` | 数据库表名 | `user_login_log` | ✅ 必填 |
| `table_prefix` | 表名前缀，生成结构体名时会去除此前缀 | `iam_` | ❌ 可选 |

#### 模型配置（用于 `model` 模式）

| 配置项 | 说明 | 示例值 | 是否必填 |
| ----- | ---- | ------ | ------- |
| `package_name` | 模型包名 | `user` | ✅ 必填 |
| `description` | 模型描述 | `用户` | ✅ 必填 |
| `table_name` | 数据库表名 | `user` | ✅ 必填 |
| `table_prefix` | 表名前缀，生成结构体名时会去除此前缀 | `iam_` | ❌ 可选 |

#### API 配置（用于 `api` 模式）

| 配置项 | 说明 | 示例值 | 是否必填 |
| ----- | ---- | ------ | ------- |
| `package_name` | API 包名 | `user` | ✅ 必填 |
| `target_filename` | 生成的目标文件名 | `user_login_log.go` | ✅ 必填 |
| `function_name` | 函数/方法名 | `Delete` | ✅ 必填 |
| `http_method` | HTTP 请求方法 | `POST`、`GET`、`PUT`、`DELETE` | ✅ 必填 |
| `description` | API 描述 | `删除登录记录` | ✅ 必填 |
| `api_doc_tag` | Swagger/API 文档标签 | `用户登录记录` | ✅ 必填 |

### 命令使用说明

```bash
# 在项目根目录（如 go-gin-web）下执行以下命令

# 生成完整模块（model + dao + service + controller + dto + router + code）
gocli generate -m module -a demoapp

# 仅生成数据层（model + dao + object）
gocli generate -m model -a demoapp

# 生成单个 API 接口（controller + service + dto + router）
gocli generate -m api -a demoapp
```

**参数说明：**
- `-m, --mode`：生成模式 - `module`、`model` 或 `api`（必填）
- `-a, --app`：应用名称，例如：`demoapp`（必填）

**使用技巧：**
- 💡 从零开始新功能时使用 `module` 模式
- 💡 只需数据库模型时使用 `model` 模式
- 💡 为现有模块添加新接口时使用 `api` 模式
- 💡 查看 [goark](https://github.com/morehao/goark) 项目的 `Makefile` 了解实际使用示例

### 生成的文件结构

当你执行 `gocli generate -m module -a demoapp` 时，工具会生成：

```
apps/demoapp/
├── model/              # 数据库模型
│   └── user.go
├── demoappdao/         # 数据访问层（命名为 {appName}dao）
│   └── user.go
├── object/             # 业务对象
│   └── objuser/
│       └── user.go
├── internal/
│   ├── controller/     # HTTP 处理器
│   │   └── ctruser/
│   │       └── user.go
│   ├── service/        # 业务逻辑层
│   │   └── svcuser/
│   │       └── user.go
│   └── dto/            # 请求/响应 DTO
│       └── dtouser/
│           ├── request.go
│           └── response.go
└── router/             # 路由注册
    └── user.go

pkg/code/               # 共享错误码
└── user.go
```

**注意**：dao 层以单层目录生成，命名为 `{appName}dao`（如 `demoappdao`），使用 `genericdao.GenericDao` 封装通用 CRUD 操作。

## cutter
`cutter`是一个命令行工具，用于快速基于现有 Go 项目创建新的 Go 项目，或在同一项目内克隆应用。

### 功能特性

#### 1. 克隆整个项目

- 在现有项目根路径下执行命令可创建新的Go项目
- 创建新项目时基于.gitignore文件过滤创建的文件
- 自动替换 import 路径
- 自动更新 go.mod 文件中的模块名称
- 自动删除 .git 目录

> ⚠️ 注意：一定要在模板项目的根路径下执行命令

#### 2. 克隆项目内的应用（新功能！）

- 在同一项目内将现有应用克隆到新应用
- 必须在项目根目录下执行命令
- 自动替换包名和 import 路径
- 替换配置文件（`.yaml`、`.yml`）中的应用名称
- 遵循 `.gitignore` 规则

### 命令使用说明

#### 克隆整个项目

```shell
cd /appTemplatePath
gocli cutter -d /yourAppPath
```

**参数说明：**
- `-d, --destination`：新项目的目标路径，例如：`/user/myApp`。此参数为必填项。

#### 克隆项目内的应用

```bash
# 在项目根目录下执行（例如 go-gin-web）
cd /path/to/go-gin-web

# 将 demoapp 克隆到 newapp
gocli cutter app -n newapp

# 或指定源应用
gocli cutter app -s demoapp -n myapp
```

**参数说明：**
- `-s, --source`：要克隆的源应用名称（默认值：`demoapp`）
- `-n, --name`：新应用名称（必填）

**使用示例：**
```bash
# 将 apps/demoapp 克隆到 apps/userapp
gocli cutter app -n userapp

# 将 apps/demoapp 克隆到 apps/adminapp
gocli cutter app -s demoapp -n adminapp
```

此命令会自动完成：
1. 复制整个应用目录结构
2. 替换所有 import 路径：`module/apps/demoapp/...` → `module/apps/newapp/...`
3. 替换配置文件中的应用名称
4. 保持 Go 代码的正确格式


