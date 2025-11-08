# 代码生成示例

这是一个用于测试代码生成工具的示例项目结构。

## 目录结构

```
_example/
├── go.mod                          # 最小化的 go.mod（仅用于测试，包含 module 声明）
├── README.md                       # 本文件
├── apps/                           # 应用目录
│   └── demoapp/                   # 示例应用
│       ├── config/                # 配置目录
│       │   └── code_gen.yaml     # 代码生成配置文件
│       └── router/                # 路由目录
│           └── enter.go          # 路由注册入口文件
└── pkg/                           # 公共包目录
    └── code/                      # 错误码目录
        └── enter.go              # 错误码注册入口文件
```

**注意**: 单元测试位于 `cmd/generate/generate_test.go` 文件中。

## 使用说明

### 1. 配置文件

在 `apps/demoapp/config/code_gen.yaml` 中配置：

- **mysql_dsn**: MySQL 数据库连接字符串
- **layer_parent_dir_map**: 各层级代码的父目录映射
- **layer_name_map**: 层级名称映射（可选）
- **layer_prefix_map**: 层级文件名前缀映射（可选）
- **module**: 模块生成配置（包名、描述、表名）
- **model**: 模型生成配置（包名、描述、表名）
- **api**: API 生成配置（包名、目标文件名、函数名等）

### 2. 运行测试

测试文件位于 `cmd/generate/generate_test.go`，在 `cmd/generate` 目录下运行：

```bash
# 进入 generate 命令目录
cd cmd/generate

# 运行所有测试
go test -v

# 运行模板加载测试
go test -v -run TestLoadTemplates

# 运行配置加载测试
go test -v -run TestConfigLoading

# 运行代码生成测试
go test -v -run TestGenerateModelCode
go test -v -run TestGenerateModuleCode
go test -v -run TestGenerateApiCode
```

**注意**: 代码生成测试会自动切换到 `_example` 目录作为项目根目录。

### 3. 生成模式

支持三种代码生成模式：

#### model 模式
生成数据模型层代码，包括：
- model: 数据模型
- dao: 数据访问对象
- object: 业务对象

```bash
# 在项目根目录运行
gocli generate --mode model --app demoapp
```

#### module 模式
生成完整模块代码，包括：
- model: 数据模型
- dao: 数据访问对象
- object: 业务对象
- controller: 控制器
- service: 服务层
- dto: 数据传输对象
- router: 路由
- code: 错误码

```bash
# 在项目根目录运行
gocli generate --mode module --app demoapp
```

#### api 模式
在已有模块基础上，添加新的 API 接口：
- controller: 添加控制器方法
- service: 添加服务层方法
- dto: 添加请求/响应对象
- router: 添加路由

```bash
# 在项目根目录运行
gocli generate --mode api --app demoapp
```

## 注意事项

1. **数据库连接**: 运行测试前请确保 MySQL 数据库可访问，并且配置文件中的连接字符串正确
2. **表结构**: 确保配置的数据表在数据库中存在
3. **测试位置**: 单元测试位于 `cmd/generate/generate_test.go`，在 `cmd/generate` 目录运行测试
4. **工作目录**: 测试会自动切换到 `_example` 目录作为项目根目录
5. **go.mod**: 此文件仅包含最小化的 module 声明，用于满足代码生成工具读取模块路径的需求

## 生成的文件位置

- **model/dao/object**: 在 `apps/demoapp/` 下生成（可通过 layer_parent_dir_map 配置）
- **controller/service/dto**: 在 `apps/demoapp/internal/` 下生成（可通过 layer_parent_dir_map 配置）
- **router**: 在 `apps/demoapp/router/` 下生成
- **code**: 在项目根目录的 `pkg/code/` 下生成

## 自定义配置

你可以通过 `code_gen.yaml` 中的以下配置来自定义生成行为：

```yaml
# 自定义层级父目录
layer_parent_dir_map:
  controller: internal
  service: internal
  dto: internal

# 自定义层级名称（例如：model -> mysqlmodel）
layer_name_map:
  model: mysqlmodel
  dao: mysqldao

# 自定义文件名前缀（例如：model 文件前缀为 mysql_）
layer_prefix_map:
  model: mysql_
  dao: mysql_
```

