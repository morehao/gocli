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
// moduleOverride 非空时强制作为新 app 的模块路径；否则推断为 base + "/" + appName。
func createAppX(appName, moduleOverride string, tidy bool) error {
	if err := scaffold.ValidateAppName(appName); err != nil {
		return err
	}
	backendDir, err := findArkBackendDir()
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

	// baseModule 是仓库的基础模块路径（如 github.com/acme/demo），用作 demo->appName token
	// 替换时保护的前缀：base 自身可能含 "demo" 子串（如项目名就叫 demo），不能被 appName 误伤。
	baseModule, err := scaffold.InferBaseModule(backendDir)
	if err != nil {
		return err
	}

	newModule := ""
	if moduleOverride != "" {
		if err := scaffold.ValidateModulePath(moduleOverride); err != nil {
			return err
		}
		newModule = moduleOverride
	} else {
		newModule = baseModule + "/" + appName
	}

	// 复制 backend/apps/demo -> apps/<appName>（保留已生成的 <module>/demo 模块路径，稍后整体替换）
	if err := scaffold.CopyDirTree(srcDemo, newAppDir); err != nil {
		scaffold.RemoveDirIfExists(newAppDir)
		return fmt.Errorf("copy app demo dir fail: %w", err)
	}
	// token 替换 demo/Demo/demoapp（demo→X 把 app 名与 <base>/demo/... 路径段改为 <base>/X/...）。
	// 先保护 baseModule，避免其自身含 demo 子串被误伤。
	if err := replaceTokensProtectedBase(newAppDir, baseModule, renameAppToken("demo", appName), ".go", ".yaml", ".yml", ".md"); err != nil {
		scaffold.RemoveDirIfExists(newAppDir)
		return err
	}
	// scripts/Dockerfile 无扩展名，无法被上一步按 ext 匹配；单独做带 base 保护的文本替换
	if err := replaceTokensInFileProtected(filepath.Join(newAppDir, "scripts", "Dockerfile"), baseModule, renameAppToken("demo", appName)); err != nil {
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
	if goWorkPath := filepath.Join(filepath.Dir(backendDir), "go.work"); goWorkPath != "" {
		if err := scaffold.AddGoWorkUse(goWorkPath, filepath.ToSlash(filepath.Join("backend", "apps", appName))); err != nil {
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

// baseModuleSentinel 用于在 token 替换期间暂存 base 模块串，防止其含 demo 子串被误换。
const baseModuleSentinel = "\x00BASEMODULE\x00"

// replaceTokensProtectedBase 等价于 scaffold.ReplaceTokensInTree，但先把 baseModule 占位为
// 哨兵串，再按 repl 替换，最后还原。避免 baseModule 自身恰含 repl 的 token（如 "demo"）被误伤。
func replaceTokensProtectedBase(rootDir, baseModule string, repl map[string]string, exts ...string) error {
	files, err := collectExtFiles(rootDir, exts...)
	if err != nil {
		return err
	}
	if err := replaceInFiles(files, baseModule, baseModuleSentinel); err != nil {
		return err
	}
	if err := scaffold.ReplaceTokensInTree(rootDir, repl, exts...); err != nil {
		return err
	}
	return replaceInFiles(files, baseModuleSentinel, baseModule)
}

// replaceTokensInFileProtected 对单个文件做带 base 保护的 token 替换（文件不存在则跳过）。
func replaceTokensInFileProtected(path, baseModule string, repl map[string]string) error {
	if _, err := os.Stat(path); err != nil {
		return nil
	}
	if err := replaceInFiles([]string{path}, baseModule, baseModuleSentinel); err != nil {
		return err
	}
	if err := replaceTokensInFile(path, repl); err != nil {
		return err
	}
	return replaceInFiles([]string{path}, baseModuleSentinel, baseModule)
}

// replaceInFiles 对 files 中每个文件做 old -> new 全量替换（内容变化才写回）。
func replaceInFiles(files []string, old, new string) error {
	for _, f := range files {
		content, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		text := strings.ReplaceAll(string(content), old, new)
		if text != string(content) {
			if err := os.WriteFile(f, []byte(text), 0o644); err != nil {
				return err
			}
		}
	}
	return nil
}

// collectExtFiles 返回 rootDir 下扩展名在 exts 中的文件（过滤规则与 scaffold.ReplaceTokensInTree 一致）。
func collectExtFiles(rootDir string, exts ...string) ([]string, error) {
	extSet := make(map[string]bool, len(exts))
	for _, e := range exts {
		extSet[strings.ToLower(e)] = true
	}
	var files []string
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if scaffold.SkipIgnoredDir(path) {
				return filepath.SkipDir
			}
			return nil
		}
		base := info.Name()
		if base == ".git" || base == "go.sum" || base == "go.work.sum" {
			return nil
		}
		if !extSet[strings.ToLower(filepath.Ext(base))] {
			return nil
		}
		files = append(files, path)
		return nil
	})
	return files, err
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

// findArkBackendDir 定位仓库下的 backend 目录：
//   - 从当前目录向上，找到"含 backend/apps/ 或 backend/pkg/ 的目录"视为仓库根，返回其 backend/ 子目录；
//   - 若当前已在 backend/（其下直接有 apps/ 且 apps/demo 存在），返回当前目录。
func findArkBackendDir() (backendDir string, err error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		// 当前目录即 backend/ ？
		if isBackendDir(dir) {
			return dir, nil
		}
		// 当前目录含 backend/ 子目录 ？
		if isBackendDir(filepath.Join(dir, "backend")) {
			return filepath.Join(dir, "backend"), nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fmt.Errorf("not inside an ark monorepo: no backend/apps found (run create project first)")
}

// isBackendDir 判断 dir 是否为 ark backend 目录：其下有 apps/ 目录且 apps/demo 存在（或 pkg/ 存在）。
func isBackendDir(dir string) bool {
	if fi, err := os.Stat(filepath.Join(dir, "apps", "demo")); err == nil && fi.IsDir() {
		return true
	}
	if fi, err := os.Stat(filepath.Join(dir, "pkg")); err == nil && fi.IsDir() {
		return true
	}
	return false
}
