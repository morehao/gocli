[English](./README.md) | [简体中文](./README_cn.md)


# gocli 介绍
`gocli` 是一个使用 Go 语言开发的命令行工具集合，旨在提升开发效率，目前包含**代码生成**和**快速新建项目**功能。

# 快速开始

## 安装

```bash
go install github.com/morehao/gocli@latest
```

## generate

`generate`是一个基于模版文件快速生成代码的工具，项目结构和风格参照[go-gin-web](https://github.com/morehao/go-gin-web)。

### 功能特性
- 可以基于`MySQL`数据库表名快速生成一个新模块的增删改查接口，并且达到可用状态。
- 基于`MySQL`数据库表名快速生成`model`和`dao`层代码
- 可以根据配置快速生成一个标准接口的骨架。
- 可以自定义各层名称、各层父级目录、各层前缀关键字。
- 生成的代码会自动进行 `gofmt` 的格式化处理。

### 命令执行前提
1. 需要在项目根目录下执行命令，例如在 `go-gin-web` 目录下执行，通过 `--app` 参数指定要生成代码的应用名称（如 `demoapp`）。
2. 需要项目对应应用下有代码生成的配置文件`code_gen.yaml`，配置文件路径为 `apps/{appName}/config/code_gen.yaml`，示例配置如下：
```yaml
mysql_dsn: root:123456@tcp(127.0.0.1:3306)/demo?charset=utf8mb4&parseTime=True&loc=Local
#layer_parent_dir_map:
#  model: model
#  dao: dao
#layer_name_map:
#  model: mysqlmodel
#  dao: mysqldao
#layer_prefix_map:
#  service: srv
module:
  package_name: user
  description: 用户登录记录
  table_name: user_login_log
model:
  package_name: user
  description: 用户
  table_name: user
api:
  package_name: user
  target_filename: user_login_log.go
  function_name: Delete
  http_method: POST
  description: 删除登录记录
  api_doc_tag: 用户登录记录
```
### 配置说明
| 配置项                  | 说明                   | 示例值                                                                           |
|----------------------|----------------------|-------------------------------------------------------------------------------|
| mysql_dsn            | MySQL 数据库连接字符串       | root:123456@tcp(127.0.0.1:3306)/demo?charset=utf8mb4&parseTime=True&loc=Local |
| layer_parent_dir_map | 层级目录映射，指定各层代码所在的父目录  | model: model                                                                  |
| layer_name_map       | 层级名称映射，用于重命名层目录      | model: mysqlmodel                                                             |
| layer_prefix_map     | 层级前缀映射关系，用于修改默认的层级名称 | service: srv                                                                  |

| 模块（module）配置  | 说明                                   | 示例值                                   |
|--------------------|--------------------------------------|------------------------------------------|
| package_name       | 模块对应的包名                         | user                                     |
| description        | 模块描述                             | 用户登录记录                             |
| table_name         | 数据库表名                           | user_login_log                           |

| 模型（model）配置   | 说明                                   | 示例值                                   |
|--------------------|--------------------------------------|------------------------------------------|
| package_name       | 模型对应的包名                       | user                                     |
| description        | 模型描述                           | 用户                                   |
| table_name         | 数据库表名                         | user                                   |

| API 配置           | 说明                                   | 示例值                                   |
|--------------------|--------------------------------------|------------------------------------------|
| package_name       | API 所属包名                        | user                                     |
| target_filename    | 生成的目标文件名                    | user_login_log.go                        |
| function_name      | 生成的函数名                        | Delete                                  |
| http_method        | HTTP 请求方法                      | POST                                   |
| description        | API 描述                           | 删除登录记录                           |
| api_doc_tag        | API 文档标签                       | 用户登录记录                           |

### 命令使用说明
```bash
# 在项目根目录（如 go-gin-web）下执行以下命令

## 生成模块代码
gocli generate -m module -a demoapp

## 生成模型代码
gocli generate -m model -a demoapp

## 生成Api接口代码
gocli generate -m api -a demoapp
```

**参数说明：**
- `-m, --mode`：生成模式，可选值：`module`（模块）、`model`（模型）、`api`（接口）
- `-a, --app`：应用名称，例如：`demoapp`（必填）

相关命令在[go-gin-web](https://github.com/morehao/go-gin-web)项目中的`Makefile`已配置相关脚本。

## cutter
`cutter`是一个命令行工具，用于快速基于现有 Go 项目创建新的 Go 项目。

### 功能特性
- 在现有项目根路径下执行命令可创建新的Go项目
- 创建新项目时基于.gitignore文件过滤创建的文件
- 自动替换 import 路径
- 自动更新 go.mod 文件中的模块名称
- 自动删除 .git 目录

> ⚠️ 注意：一定要在模板项目的根路径下执行命令***
### 命令使用说明
```shell
cd /appTemplatePath
gocli cutter -d /yourAppPath
```

- `-d, --destination`：新项目的目标路径，例如：`/user/myApp`。此参数为必填项。


