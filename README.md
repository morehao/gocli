[English](./README.md) | [简体中文](./README_cn.md)

# gocli Introduction

`gocli` is a command-line toolset written in Go, designed to boost development efficiency. It currently includes features for **code generation** and **quick project scaffolding**.

## Quick Start

### Installation

```bash
go install github.com/morehao/gocli@latest
```

## generate

`generate` is a tool that generates code quickly based on template files. The project structure and style are modeled after [go-gin-web](https://github.com/morehao/go-gin-web).

### Features

* Quickly generate a complete CRUD interface for a new module based on a MySQL table, ready for use.
* Generate `model` and `dao` layer code based on MySQL table name.
* Quickly scaffold a standard API interface from configuration.
* Customize layer names, parent directories, and layer name prefixes.
* Automatically formats the generated code using `gofmt`.

### Prerequisites

1. Run the command in the project root directory (e.g., `go-gin-web`), and specify the application name using the `--app` parameter (e.g., `demoapp`).
2. Ensure that the application contains a configuration file named `code_gen.yaml` at the path `apps/{appName}/config/code_gen.yaml`. Example:

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
  description: User login records
  table_name: user_login_log
model:
  package_name: user
  description: User
  table_name: user
api:
  package_name: user
  target_filename: user_login_log.go
  function_name: Delete
  http_method: POST
  description: Delete login record
  api_doc_tag: User login records
```

### Configuration Reference

| Field                   | Description                          | Example                                                                          |
| ----------------------- | ------------------------------------ | -------------------------------------------------------------------------------- |
| mysql\_dsn              | MySQL DSN connection string          | root:123456\@tcp(127.0.0.1:3306)/demo?charset=utf8mb4\&parseTime=True\&loc=Local |
| layer\_parent\_dir\_map | Map of parent directories for layers | model: model                                                                     |
| layer\_name\_map        | Mapping to rename code layers        | model: mysqlmodel                                                                |
| layer\_prefix\_map      | Prefix mapping for layer names       | service: srv                                                                     |

#### Module Configuration

| Field         | Description               | Example            |
| ------------- | ------------------------- | ------------------ |
| package\_name | Go package name           | user               |
| description   | Description of the module | User login records |
| table\_name   | MySQL table name          | user\_login\_log   |

#### Model Configuration

| Field         | Description              | Example |
| ------------- | ------------------------ | ------- |
| package\_name | Model package name       | user    |
| description   | Description of the model | User    |
| table\_name   | Table name in database   | user    |

#### API Configuration

| Field            | Description                | Example             |
| ---------------- | -------------------------- | ------------------- |
| package\_name    | API package name           | user                |
| target\_filename | Name of the generated file | user\_login\_log.go |
| function\_name   | Generated function name    | Delete              |
| http\_method     | HTTP method for the API    | POST                |
| description      | Description of the API     | Delete login record |
| api\_doc\_tag    | API doc tag                | User login records  |

### Command Usage

```bash
# Run the following commands in the project root directory (e.g., go-gin-web)

# Generate module code
gocli generate -m module -a demoapp

# Generate model code
gocli generate -m model -a demoapp

# Generate API interface code
gocli generate -m api -a demoapp
```

**Parameter Description:**
- `-m, --mode`: Generation mode, options: `module`, `model`, `api` (required)
- `-a, --app`: Application name, e.g., `demoapp` (required)

The [go-gin-web](https://github.com/morehao/go-gin-web) project contains example usage in its `Makefile`.

---

## cutter

`cutter` is a CLI tool for quickly creating a new Go project based on an existing template project.

### Features

* Must be executed from the root directory of the template project.
* Filters copied files using `.gitignore`.
* Replaces import paths automatically.
* Updates the module name in `go.mod`.
* Deletes the `.git` directory from the new project.

> ⚠️ Note: Be sure to run the command from the **root directory** of the template project.

### Command Usage

```bash
cd /appTemplatePath
gocli cutter -d /yourAppPath
```

* `-d, --destination`: Destination path for the new project, e.g., `/user/myApp` (required).

