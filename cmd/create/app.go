package create

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/morehao/gocli/internal/scaffold"
)

// capFirst 首字母大写。
func capFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// renameAppToken 计算 app 名替换映射（key 长度降序由 ReplaceTokensInTree 保证，demoapp 不会被 demo 误伤）。
func renameAppToken(oldApp, newApp string) map[string]string {
	return map[string]string{
		oldApp + "app":  newApp + "app",
		capFirst(oldApp): capFirst(newApp),
		oldApp:          newApp,
	}
}

// createAppX 在 monorepo 的 backend/ 下新建一个 app。
// 从已有的 backend/apps/demo（已由 create project 生成为 <module>/demo）复制到 apps/<appName>，
// 再按 renameAppToken 把 demo/Demo/demoapp 全部替换为 appName 形态（含 import 路径里的
// <module>/demo/xxx，因为 ReplaceTokensInTree 会把 "demo" 子串一并替换为 appName）。
func createAppX(appName string, tidy bool) error {
	if err := scaffold.ValidateAppName(appName); err != nil {
		return err
	}
	backendDir, goWorkPath, err := findArkBackendDir()
	if err != nil {
		return err
	}
	appsDir := filepath.Join(backendDir, "apps")
	srcDemo := filepath.Join(appsDir, "demo")
	if _, err := os.Stat(srcDemo); err != nil {
		return fmt.Errorf("%s is not an ark monorepo: apps/demo does not exist (run create project first)", backendDir)
	}
	newAppDir := filepath.Join(appsDir, appName)
	if _, err := os.Stat(newAppDir); err == nil {
		return fmt.Errorf("app already exists: %s", newAppDir)
	}

	base, err := scaffold.InferBaseModule(backendDir)
	if err != nil {
		return err
	}
	newModule := base + "/" + appName

	// 复制 backend/apps/demo -> apps/<appName>（保留已生成的 <module>/demo 模块路径，稍后整体替换）
	if err := scaffold.CopyDirTree(srcDemo, newAppDir); err != nil {
		scaffold.RemoveDirIfExists(newAppDir)
		return fmt.Errorf("copy app demo dir fail: %w", err)
	}
	// token 替换 demo/Demo/demoapp（demo→X 同时把 import 路径中的 <module>/demo/... 改为 <module>/X/...）
	if err := scaffold.ReplaceTokensInTree(newAppDir, renameAppToken("demo", appName), ".go", ".yaml", ".yml", ".md"); err != nil {
		scaffold.RemoveDirIfExists(newAppDir)
		return err
	}
	// scripts/Dockerfile 无扩展名，无法被上一步按 ext 匹配；单独做文本替换（引用 apps/demo/cmd、/app/demo）
	if err := replaceTokensInFile(filepath.Join(newAppDir, "scripts", "Dockerfile"), renameAppToken("demo", appName)); err != nil {
		scaffold.RemoveDirIfExists(newAppDir)
		return err
	}
	// 模块重写
	appGoMod := filepath.Join(newAppDir, "go.mod")
	if err := scaffold.RewriteModuleStmt(appGoMod, newModule); err != nil {
		scaffold.RemoveDirIfExists(newAppDir)
		return err
	}
	// pkg 联动追加
	if err := scaffold.AppendAppConnector(filepath.Join(backendDir, "pkg"), "demo", appName); err != nil {
		scaffold.RemoveDirIfExists(newAppDir)
		return err
	}
	if goWorkPath != "" {
		if err := scaffold.AddGoWorkUse(goWorkPath, filepath.ToSlash(filepath.Join("apps", appName))); err != nil {
			scaffold.RemoveDirIfExists(newAppDir)
			return err
		}
	}
	if tidy {
		if err := scaffold.RunGoModTidy(newAppDir); err != nil {
			fmt.Printf("Warning: go mod tidy failed (run it manually in %s): %v\n", newAppDir, err)
		}
	}
	fmt.Printf("Successfully created app %s at %s (module: %s)\n", appName, newAppDir, newModule)
	fmt.Printf("  gocli generate module -a %s   # generate code for the new app\n", appName)
	return nil
}

// replaceTokensInFile 对单个文件按 repl 做全量文本替换（key 长度降序）。
func replaceTokensInFile(path string, repl map[string]string) error {
	if _, err := os.Stat(path); err != nil {
		return nil // 文件不存在则跳过
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	text := string(content)
	for _, k := range orderedKeys(repl) {
		text = strings.ReplaceAll(text, k, repl[k])
	}
	return os.WriteFile(path, []byte(text), 0o644)
}

// orderedKeys 返回 repl 的 key，按长度降序（避免短 token 吞长 token）。
func orderedKeys(repl map[string]string) []string {
	keys := make([]string, 0, len(repl))
	for k := range repl {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return len(keys[i]) > len(keys[j]) })
	return keys
}

// findArkBackendDir 定位含 backend/go.work 的 backend 目录；找不到则向上层查找仓库根下的 backend/。
func findArkBackendDir() (backendDir, goWorkPath string, err error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", "", err
	}
	for {
		cand := filepath.Join(dir, "backend", "go.work")
		if fi, err := os.Stat(cand); err == nil && !fi.IsDir() {
			return filepath.Join(dir, "backend"), cand, nil
		}
		gw := filepath.Join(dir, "go.work")
		if fi, err := os.Stat(gw); err == nil && !fi.IsDir() {
			if fi2, err := os.Stat(filepath.Join(dir, "apps")); err == nil && fi2.IsDir() {
				return dir, gw, nil // 已在 backend 目录
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", "", fmt.Errorf("not inside an ark monorepo: no backend/go.work found")
}
