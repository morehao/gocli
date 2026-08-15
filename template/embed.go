// Package template 内嵌 gocli 自带的项目模板资产。
package template

import "embed"

//go:embed monorepo
var MonorepoFS embed.FS
