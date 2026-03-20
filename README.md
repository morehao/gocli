[简体中文](./README_CN.md) | [English](./README.md)

# gocli Introduction

`gocli` is a command-line toolset written in Go, designed to boost development efficiency. It currently includes features for **code generation** and **quick project scaffolding**.

## Quick Start

### Installation

```bash
go install github.com/morehao/gocli@latest
```

## generate

`generate` is a powerful code generation tool based on template files and database schema. The project structure and style are modeled after [goark](https://github.com/morehao/goark).

### Features

* 🚀 **Fast Development**: Quickly generate a complete CRUD module based on MySQL table structure
* 📦 **Multi-Layer Generation**: Supports model, dao, service, controller, dto, router, and more
* 🎯 **Three Generation Modes**: module (full module), model (data layer only), api (single API endpoint)
* 🔧 **Highly Customizable**: Configure layer names, parent directories, and file name prefixes
* ✨ **Auto Formatting**: Automatically formats generated code using `gofmt`
* 📖 **Database-Driven**: Reads MySQL table structure to generate accurate model definitions

### Generation Modes

#### 1. **module** - Full Module Generation

Generates a complete CRUD module including all layers:
- **model**: Database model
- **dao**: Data Access Object
- **object**: Business object
- **controller**: HTTP request handler
- **service**: Business logic layer
- **dto**: Request/Response objects
- **router**: Route registration
- **code**: Error code definitions

**Use Case**: Creating a new feature module from scratch

```bash
gocli generate -m module -a demoapp
```

#### 2. **model** - Data Layer Generation

Generates only the data layer code:
- **model**: Database model with GORM tags
- **dao**: Data access methods
- **object**: Business object for data transformation

**Use Case**: Adding a new database table without full CRUD operations

```bash
gocli generate -m model -a demoapp
```

#### 3. **api** - Single API Endpoint

Adds a new API endpoint to an existing module:
- **controller**: New controller method
- **service**: New service method
- **dto**: Request/Response structs
- **router**: Route registration

**Use Case**: Adding a new endpoint to an existing feature

```bash
gocli generate -m api -a demoapp
```

### Prerequisites

1. **Execute in project root**: Run the command in the project root directory (e.g., `go-gin-web`)
2. **Specify app name**: Use the `--app` parameter to specify the application name (e.g., `demoapp`)
3. **Configuration file required**: Ensure `apps/{appName}/config/code_gen.yaml` exists

Example configuration file:

```yaml
mysql_dsn: root:123456@tcp(127.0.0.1:3306)/demo?charset=utf8mb4&parseTime=True&loc=Local
service_name: mysql
module:
  package_name: user
  description: User login records
  table_name: user_login_log
  table_prefix: ""   # Optional: Table name prefix, will be removed when generating struct name (e.g., "iam_")
model:
  package_name: user
  description: User
  table_name: user
  table_prefix: ""   # Optional: Table name prefix
api:
  package_name: user
  target_filename: user_login_log.go
  function_name: Delete
  http_method: POST
  description: Delete login record
  api_doc_tag: User login records
```

### Configuration Reference

#### Global Configuration

| Field | Description | Example | Required |
| ----- | ----------- | ------- | -------- |
| `mysql_dsn` | MySQL database connection string | `root:123456@tcp(127.0.0.1:3306)/demo?charset=utf8mb4&parseTime=True&loc=Local` | ✅ Yes |
| `service_name` | Layer name prefix for model/dao directories and DB connection name | `mysql` | ✅ Yes |



| Field | Description | Example | Required |
| ----- | ----------- | ------- | -------- |
| `package_name` | Package name for the module | `user` | ✅ Yes |
| `description` | Module description (for comments) | `User login records` | ✅ Yes |
| `table_name` | MySQL table name | `user_login_log` | ✅ Yes |
| `table_prefix` | Table name prefix, removed when generating struct name | `iam_` | ❌ Optional |

#### Model Configuration (for `model` mode)

| Field | Description | Example | Required |
| ----- | ----------- | ------- | -------- |
| `package_name` | Package name for the model | `user` | ✅ Yes |
| `description` | Model description | `User` | ✅ Yes |
| `table_name` | MySQL table name | `user` | ✅ Yes |
| `table_prefix` | Table name prefix, removed when generating struct name | `iam_` | ❌ Optional |

#### API Configuration (for `api` mode)

| Field | Description | Example | Required |
| ----- | ----------- | ------- | -------- |
| `package_name` | Package name for the API | `user` | ✅ Yes |
| `target_filename` | Target file name for generated code | `user_login_log.go` | ✅ Yes |
| `function_name` | Function/method name | `Delete` | ✅ Yes |
| `http_method` | HTTP method | `POST`, `GET`, `PUT`, `DELETE` | ✅ Yes |
| `description` | API description | `Delete login record` | ✅ Yes |
| `api_doc_tag` | Swagger/API doc tag | `User login records` | ✅ Yes |

### Command Usage

```bash
# Run commands in the project root directory (e.g., go-gin-web)

# Generate a complete module (model + dao + service + controller + dto + router + code)
gocli generate -m module -a demoapp

# Generate only data layer (model + dao + object)
gocli generate -m model -a demoapp

# Generate a single API endpoint (controller + service + dto + router)
gocli generate -m api -a demoapp
```

**Parameters:**
- `-m, --mode`: Generation mode - `module`, `model`, or `api` (required)
- `-a, --app`: Application name, e.g., `demoapp` (required)

**Quick Tips:**
- 💡 Use `module` mode when starting a new feature from scratch
- 💡 Use `model` mode when you only need database models
- 💡 Use `api` mode to add new endpoints to existing modules
- 💡 Check the [goark](https://github.com/morehao/goark) `Makefile` for practical examples

### Generated File Structure

When you run `gocli generate -m module -a demoapp`, the tool generates:

```
apps/demoapp/
├── model/              # Database models
│   └── user.go
├── demoappdao/         # Data access layer (named as {appName}dao)
│   └── user.go
├── object/             # Business objects
│   └── objuser/
│       └── user.go
├── internal/
│   ├── controller/     # HTTP handlers
│   │   └── ctruser/
│   │       └── user.go
│   ├── service/        # Business logic
│   │   └── svcuser/
│   │       └── user.go
│   └── dto/            # Request/Response DTOs
│       └── dtouser/
│           ├── request.go
│           └── response.go
└── router/             # Route registration
    └── user.go

pkg/code/               # Shared error codes
└── user.go
```

**Note**: The dao layer is generated as a single-level directory named `{appName}dao` (e.g., `demoappdao`), using `genericdao.GenericDao` for common CRUD operations.

---

## cutter

`cutter` is a CLI tool for quickly creating a new Go project based on an existing template project, or cloning an app within the same project.

### Features

#### 1. Clone Entire Project

* Must be executed from the root directory of the template project.
* Filters copied files using `.gitignore`.
* Replaces import paths automatically.
* Updates the module name in `go.mod`.
* Deletes the `.git` directory from the new project.

> ⚠️ Note: Be sure to run the command from the **root directory** of the template project.

#### 2. Clone App Within Project (New!)

* Clone an existing app to a new app within the same project.
* Must be executed from the project root directory.
* Automatically replaces package names and import paths.
* Replaces app names in configuration files (`.yaml`, `.yml`).
* Follows `.gitignore` rules.

### Command Usage

#### Clone Entire Project

```bash
cd /appTemplatePath
gocli cutter -d /yourAppPath
```

**Parameters:**
* `-d, --destination`: Destination path for the new project, e.g., `/user/myApp` (required).

#### Clone App Within Project

```bash
# Run in project root directory (e.g., go-gin-web)
cd /path/to/go-gin-web

# Clone demoapp to newapp
gocli cutter app -n newapp

# Or specify source app
gocli cutter app -s demoapp -n myapp
```

**Parameters:**
* `-s, --source`: Source app name to clone from (default: `demoapp`).
* `-n, --name`: New app name (required).

**Example:**
```bash
# Clone apps/demoapp to apps/userapp
gocli cutter app -n userapp

# Clone apps/demoapp to apps/adminapp
gocli cutter app -s demoapp -n adminapp
```

This command will:
1. Copy the entire app directory structure
2. Replace all import paths: `module/apps/demoapp/...` → `module/apps/newapp/...`
3. Replace app names in configuration files
4. Maintain proper Go code formatting

