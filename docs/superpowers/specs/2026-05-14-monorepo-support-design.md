# gocli cutter app Monorepo 支持设计

## 问题

`gocli cutter app -s demoapp -n newapp` 在 monorepo 场景下报错：

```
Error cloning app: /Users/songhao/Documents/practice/go/goark is not a Go project (no go.mod found)
```

当前实现只检测当前目录是否有 `go.mod`，不支持 monorepo。

## 支持的场景

### 场景 1：Monorepo 有根 go.mod，子 app 无独立 go.mod
```
monorepo/
├── go.mod          # 根 module: example.com/monorepo
├── go.work         # 可选
└── apps/
    └── demoapp/     # 无独立 go.mod，共享根 go.mod
        └── ...
```

### 场景 2：Monorepo 用 go.work（无根 go.mod），子 app 有独立 go.mod
```
monorepo/
├── go.work         # workspace 根目录
└── apps/
    └── demoapp/
        └── go.mod  # 独立 module: example.com/monorepo/apps/demoapp
```

## 设计

### 核心函数：findProjectRoot

```go
// findProjectRoot 查找项目根目录
// currentDir: 当前工作目录
// 返回:
//   - rootDir: 项目根目录（go.work 所在目录 或 包含 go.mod 的目录）
//   - isWorkspace: 是否为 workspace 模式
//   - modulePath: 模块路径（从根 go.mod 或子 app go.mod 读取）
func findProjectRoot(currentDir string) (rootDir string, isWorkspace bool, modulePath string, err error)
```

### 处理逻辑

1. **检测 go.work**：从当前目录向上查找 `go.work`
2. **如果存在 go.work**：
   - workspace 模式
   - 根目录 = go.work 所在目录
   - modulePath = 从当前子目录的 go.mod 读取
3. **如果不存在 go.work**：
   - 向上遍历找 go.mod
   - 找到则根目录 = go.mod 所在目录
   - modulePath = 从该 go.mod 读取
4. **都找不到**：返回错误

### 代码改动

#### util.go 新增

```go
// findProjectRoot 查找项目根目录
func findProjectRoot(currentDir string) (rootDir string, isWorkspace bool, modulePath string, err error)

// isGoWork 检查是否存在 go.work 文件
func isGoWork(path string) bool

// readModulePath 读取 go.mod 的 module 路径
func readModulePath(goModPath string) (string, error)
```

#### app.go 修改

```go
// cloneApp 修改前
if !isGoProject(currentDir) {
    return fmt.Errorf("%s is not a Go project (no go.mod found)", currentDir)
}
modFilePath := filepath.Join(currentDir, "go.mod")

// cloneApp 修改后
rootDir, isWorkspace, modulePath, err := findProjectRoot(currentDir)
if err != nil {
    return err
}
appsDir := filepath.Join(rootDir, "apps")
```

## 实现计划

1. [ ] util.go 新增 `findProjectRoot`、`isGoWork`、`readModulePath`
2. [ ] app.go 修改 `cloneApp` 使用 `findProjectRoot`
3. [ ] 测试场景 1（根 go.mod，无子 go.mod）
4. [ ] 测试场景 2（go.work，有子 go.mod）

## 风险与注意事项

- go.work 可能存在于子目录（非根目录），需要正确处理
- 需要处理循环引用检测
- 保持向后兼容，单项目场景不受影响