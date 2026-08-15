# 模板资产

本目录集中存放 gocli 的全部内置模板，通过 `template/embed.go` 中的 `go:embed` 内嵌进二进制。

## 目录结构

```
template/
├── embed.go        # go:embed 入口（MonorepoFS / GenerateFS）
├── README.md       # 本文件
├── generate/       # generate 命令的代码生成模板
│   ├── module/     # 完整 CRUD 模块生成（model/dao/service/controller/dto/router/code）
│   ├── model/      # 仅数据层生成（model/dao/object）
│   └── api/        # 单个接口生成（controller/service/dto/router）
└── monorepo/       # create 命令的后端 monorepo 项目模板
    ├── go.work.tmpl
    ├── apps/demoapp/   # 示例应用（gin + gorm）
    └── pkg/            # 公共库
```

## 后缀约定（两类模板的后缀不可混用）

| 目录 | 后缀 | 含义 | 原因 |
| ---- | ---- | ---- | ---- |
| `generate/` | `.tpl` | text/template 模板引擎文件，文件名即模板名（如 `controller.go.tpl` 生成 `controller.go`） | golib codegen 硬编码 `.tpl` 后缀用于识别模板并推导目标文件名，不可更改 |
| `monorepo/` | `.tmpl` | 原始文件改名存储，创建时恢复为标准文件名（`go.mod.tmpl` → `go.mod`、`main.go.tmpl` → `main.go`） | 绕过 go:embed 拒绝嵌入含 `go.mod` 的嵌套模块目录，并避免模板内 `.go` 文件被父模块编译 |

**注意**：不要将 `generate/` 下的模板改为 `.tmpl`，也不要将 `monorepo/` 下的文件改为 `.tpl`，两者被不同的消费方（golib codegen / internal/scaffold.RestoreTemplateFiles）以固定后缀解析。
