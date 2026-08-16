# TestSetup

测试工具包，基于 `golib/biz/testkit` 封装，为 GoArk 项目提供统一的测试环境初始化能力。

## 快速开始

```go
package myservice_test

import (
    "os"
    "testing"
    "github.com/morehao/go-ark-template/pkg/testsetup"
)

func TestMain(m *testing.M) {
    testsetup.Initialize(testsetup.AppNameDemo)
    code := m.Run()
    testsetup.Close(testsetup.AppNameDemo)
    os.Exit(code)
}

func TestYourFunction(t *testing.T) {
    ctx := testsetup.NewContext(
        testsetup.WithUserID(123),
        testsetup.WithMethod("POST"),
        testsetup.WithJSON(),
    )
    // 测试代码...
}
```

## 核心 API

### 初始化与清理

| 函数 | 说明 |
|------|------|
| `Initialize(appName string)` | 初始化测试环境 |
| `Close(appName string)` | 清理测试资源 |
| `Init(appName string)` | Initialize 的别名 |
| `Done(appName string)` | Close 的别名 |

### 应用常量

```go
testsetup.AppNameDemo // "demo"
```

### 测试上下文

```go
ctx := testsetup.NewContext(opts...)
```

支持的可选参数：

| 函数 | 说明 |
|------|------|
| `WithUserID` | 用户 ID |
| `WithCompanyID` | 公司 ID |
| `WithRequestID` | 请求 ID |
| `WithKeyValue` | 自定义键值 |
| `WithHeader` | 单个 HTTP 头部 |
| `WithHeaders` | 多个 HTTP 头部 |
| `WithMethod` | 请求方法 |
| `WithURL` | 请求 URL |
| `WithQueryParam` | 单个查询参数 |
| `WithQueryParams` | 多个查询参数 |
| `WithContentType` | 内容类型 |
| `WithJSON` | JSON 请求 |
| `WithFormData` | Form 表单 |
| `WithMultipartFormData` | 多部分表单 |
| `WithAuth` | 认证信息 |
| `WithBearerToken` | Bearer Token |
| `WithClientIP` | 客户端 IP |
| `WithBody` | 请求体 |

## 架构设计

```
pkg/testsetup/
├── init.go                   # 核心 API 导出
├── constant.go               # 应用名称常量
├── base_initializer.go       # 基础初始化器
├── initializer_demo.go       # Demo 应用初始化器
└── README.md                 # 文档
```

## 初始化流程

1. 查找配置文件：`apps/{appName}/config/config.yaml`
2. 加载应用配置：调用 `config.LoadConfig(configPath)`
3. 初始化日志
4. 初始化数据库资源（GORM）
5. 初始化 Redis（如果配置）
6. 初始化 Elasticsearch（如果配置）

## 添加新应用

### 1. 创建初始化器

参考 `initializer_demo.go` 创建新应用的初始化器：

```go
// pkg/testsetup/initializer_myapp.go
package testsetup

import (
    "fmt"
    "os"

    "github.com/morehao/go-ark-template/apps/myapp/config"
    "github.com/morehao/golib/biz/testkit"
)

type myappInitializer struct {
    *baseAppInitializer
    *testkit.BaseInitializer
}

func newMyappInitializer() (Initializer, error) {
    base, err := newBaseAppInitializer(AppNameMyapp)
    if err != nil {
        return nil, err
    }
    baseInit, err := testkit.NewBaseInitializer(AppNameMyapp)
    if err != nil {
        return nil, err
    }
    return &myappInitializer{
        baseAppInitializer: base,
        BaseInitializer:    baseInit,
    }, nil
}

func (m *myappInitializer) Initialize() error {
    if _, err := os.Stat(m.ConfigPath); err != nil {
        return fmt.Errorf("config file not found: %s, error: %w", m.ConfigPath, err)
    }

    var panicErr interface{}
    func() {
        defer func() {
            if r := recover(); r != nil {
                panicErr = r
            }
        }()
        config.LoadConfig(m.ConfigPath)
    }()

    if panicErr != nil {
        return fmt.Errorf("load config failed: %v", panicErr)
    }

    m.Log = config.Conf.Log
    m.DBConfigs = config.Conf.DBConfigs
    m.RedisConfig = config.Conf.RedisConfig
    m.ESConfigs = config.Conf.ESConfigs

    if err := m.initLog(); err != nil {
        return fmt.Errorf("init log: %w", err)
    }

    if err := m.initResources(); err != nil {
        return fmt.Errorf("init resources: %w", err)
    }

    return nil
}
```

### 2. 注册应用

在 `init.go` 的 `initializerCreators` 中添加：

```go
var initializerCreators = map[string]InitializerFunc{
    AppNameDemo: newDemoappInitializer,
    "myapp":     newMyappInitializer,
}
```

### 3. 添加常量

在 `constant.go` 中添加：

```go
const (
    AppNameDemo  = "demo"
    AppNameMyapp = "myapp"
)
```

## 配置文件

配置文件路径：`apps/{appName}/config/config.yaml`

```yaml
log:
  default:
    service: myapp
    level: info
    writer: file
    dir: ../../../log
  gorm:
    level: warn

db_configs:
  - url: "mysql://root:password@127.0.0.1:3306/mydb?charset=utf8mb4&parseTime=True&loc=Local"

redis_config:
  service: myapp
  addr: 127.0.0.1:6379

es_configs:
  - service: myapp
    addr: http://127.0.0.1:9200
```