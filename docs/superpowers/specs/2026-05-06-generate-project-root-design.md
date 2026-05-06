# Generate 命令适配多项目结构设计方案

## 日期：2026-05-06

## 背景

当前 `generate` 命令通过 `GetAppInfo` 函数解析项目结构，其中 `ProjectRootPath` 的确定逻辑是：从 `workDir` 向上查找 `apps` 目录，其上级即为项目根目录。

这种逻辑在简单 monorepo 结构下工作正常，但在存在中间层（如前后端分离项目）的场景下会失效。

## 问题描述

### 情况1：简单 monorepo
```
goark/
├── .git
├── pkg/code/
└── apps/demo/           # workDir
    └── go.mod           # module = github.com/morehao/goark/apps/demo
```
- 裁剪 `apps/demo` → `github.com/morehao/goark` ✓
- `ProjectRootPath` = `goark` ✓

### 情况2：前后端分离 monorepo
```
ark-iam/                 # git 仓库根
├── .git
├── pkg/code/
└── backend/
    └── apps/iam/        # workDir
        └── go.mod       # module = github.com/morehao/ark-iam/iam
```
- 裁剪 `apps/iam` → `github.com/morehao/ark-iam` (期望)
- 但旧逻辑 `ProjectRootPath` = `ark-iam/backend` ✗

### 核心矛盾

- **ModulePath**：从 go.mod 读取并裁剪 `apps/{appName}` 后缀，可正确得到 `github.com/morehao/ark-iam`
- **ProjectRootPath**：通过 `apps` 上级目录确定，得到 `ark-iam/backend` 而非期望的 `ark-iam`

## 解决方案

### 设计原则
- 不引入额外配置
- 自动检测，无需用户干预
- 向后兼容简单结构

### 实现方案：通过 `.git` 目录确定仓库根

新增 `findGitRoot` 函数，从 `workDir` 向上逐层查找 `.git` 目录，该目录即为仓库根目录。

```go
func findGitRoot(workDir string) (string, error) {
    current := workDir
    for {
        gitPath := filepath.Join(current, ".git")
        if _, err := os.Stat(gitPath); err == nil {
            return current, nil
        }
        parent := filepath.Dir(current)
        if parent == current {
            break
        }
        current = parent
    }
    return "", fmt.Errorf(".git directory not found")
}
```

修改 `GetAppInfo` 函数，在确定 `ProjectRootPath` 时：
1. 先用原有逻辑找到 `apps` 的上级目录作为初步结果
2. 调用 `findGitRoot` 查找 `.git` 所在目录
3. 如果 `.git` 目录在初步结果的上级，则使用 `.git` 所在目录

### 效果对比

| | 情况1 (goark) | 情况2 (ark-iam) |
|---|---|---|
| workDir | `goark/apps/demo` | `ark-iam/backend/apps/iam` |
| 旧 ProjectRootPath | `goark` ✓ | `ark-iam/backend` ✗ |
| 新 ProjectRootPath | `goark` ✓ | `ark-iam` ✓ |
| 错误码位置 | `goark/pkg/code` ✓ | `ark-iam/pkg/code` ✓ |

## 变更点

### 1. util.go - 新增 `findGitRoot` 函数
在 `util.go` 中新增通过 `.git` 目录查找仓库根的函数。

### 2. util.go - 修改 `GetAppInfo` 函数
调整 `ProjectRootPath` 的确定逻辑，优先使用 `.git` 检测结果。

### 3. 模板不变更
- 模板中 import 路径使用 `ModulePath`（从 go.mod 读取并裁剪）
- `ModulePath` 不受 `ProjectRootPath` 变更影响
- 情况2 的 module = `github.com/morehao/ark-iam/iam`，裁剪后 `ModulePath` = `github.com/morehao/ark-iam`

## 影响范围

### 需要修改的文件
- `cmd/generate/util.go`：新增 `findGitRoot`，修改 `GetAppInfo`

### 不需要修改的文件
- `cmd/generate/config.go`
- `cmd/generate/generate.go`
- `cmd/generate/gen_module.go`
- `cmd/generate/gen_model.go`
- `cmd/generate/gen_api.go`
- `cmd/generate/template/` 下所有模板文件

### 依赖 `ProjectRootPath` 的代码
仅用于确定错误码文件写入位置（`pkg/code/` 目录），不影响代码 import 路径。
