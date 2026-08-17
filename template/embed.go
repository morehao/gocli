// Package template 内嵌 gocli 自带的全部模板资产：
//
//   - monorepo: 后端 monorepo 项目模板（供 create 命令使用；go.mod/go.sum/go.work/Go 源文件
//     以 .tmpl 后缀存放，创建时由 internal/scaffold.RestoreTemplateFiles 恢复为标准文件名，
//     以绕过 go:embed 拒绝嵌入嵌套 go.mod 模块的限制，并避免 .go 文件被父模块编译）
//   - generate: 代码生成模板（供 generate 命令使用；.tpl 为 text/template 模板文件，
//     后缀由 golib codegen 硬编码，不可更改）
package template

import "embed"

//go:embed monorepo
var MonorepoFS embed.FS

//go:embed generate
var GenerateFS embed.FS

//go:embed ark
var ArkFS embed.FS
