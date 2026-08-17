package cutter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/morehao/gocli/internal/scaffold"
	"golang.org/x/mod/modfile"
)

type projectContext struct {
	rootDir       string
	goModPath     string
	goWorkPath    string
	goWorkUseDirs []string
}

// appsDir 返回项目内的应用目录。兼容两种 monorepo 布局：
//   - ark 结构（create project 生成）：backend/apps
//   - 经典结构（cutter 原支持）：apps
//
// 若 rootDir 下存在 backend/apps 目录则优先使用，否则回退到 rootDir/apps。
func (ctx projectContext) appsDir() string {
	arkApps := filepath.Join(ctx.rootDir, "backend", "apps")
	if fi, err := os.Stat(arkApps); err == nil && fi.IsDir() {
		return arkApps
	}
	return filepath.Join(ctx.rootDir, "apps")
}

type resolvedModule struct {
	rootDir         string
	goModPath       string
	modulePath      string
	oldImportPrefix string
}

type projectCloneMode string

const (
	projectCloneModeRootModule    projectCloneMode = "root-module"
	projectCloneModeWorkspaceOnly projectCloneMode = "workspace-only"
)

// isGoProject 检查指定路径是否为Go项目
func isGoProject(path string) bool {
	return scaffold.IsGoProject(path)
}

// shouldIgnore 检查路径是否应该被忽略（实现见 internal/scaffold.ShouldIgnore）
func shouldIgnore(relativePath string) bool {
	return scaffold.ShouldIgnore(relativePath)
}

// copyFile 复制文件（实现见 internal/scaffold.CopyFile）
func copyFile(src, dst string) error {
	return scaffold.CopyFile(src, dst)
}

// isGoWork 检查指定路径是否为 Go workspace（是否包含 go.work 文件）
// isGoWork 检查指定路径是否为 Go workspace（是否包含 go.work 文件）
func isGoWork(path string) bool {
	return scaffold.IsGoWork(path)
}

// readModulePath 读取 go.mod 的 module 路径（实现见 internal/scaffold.ReadModulePath）
func readModulePath(goModPath string) (string, error) {
	return scaffold.ReadModulePath(goModPath)
}

func resolveProjectCloneMode(ctx projectContext) projectCloneMode {
	if ctx.rootDir != "" && ctx.goModPath == filepath.Join(ctx.rootDir, "go.mod") {
		return projectCloneModeRootModule
	}
	return projectCloneModeWorkspaceOnly
}

func maybeModifyGoMod(dstDir, moduleName string) error {
	goModPath := filepath.Join(dstDir, "go.mod")
	if _, err := os.Stat(goModPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return modifyGoMod(dstDir, moduleName)
}

func maybeModifyAppGoMod(appDir, oldAppName, newAppName string) error {
	goModPath := filepath.Join(appDir, "go.mod")
	if _, err := os.Stat(goModPath); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	content, err := os.ReadFile(goModPath)
	if err != nil {
		return err
	}
	modFile, err := modfile.Parse(goModPath, content, nil)
	if err != nil {
		return err
	}
	if modFile.Module == nil {
		return nil
	}

	oldModulePath := modFile.Module.Mod.Path
	changed := false
	if newModulePath := renameAppModulePath(oldModulePath, oldModulePath, oldAppName, newAppName); newModulePath != oldModulePath {
		if err := modFile.AddModuleStmt(newModulePath); err != nil {
			return err
		}
		changed = true
	}

	type requireRewrite struct {
		oldPath string
		newPath string
		version string
	}
	requireRewrites := make([]requireRewrite, 0)
	for _, req := range modFile.Require {
		if !belongsToAppModulePrefix(req.Mod.Path, oldModulePath) {
			continue
		}
		if newRequirePath := renameAppModulePath(req.Mod.Path, oldModulePath, oldAppName, newAppName); newRequirePath != req.Mod.Path {
			requireRewrites = append(requireRewrites, requireRewrite{
				oldPath: req.Mod.Path,
				newPath: newRequirePath,
				version: req.Mod.Version,
			})
		}
	}
	for _, rewrite := range requireRewrites {
		if err := modFile.DropRequire(rewrite.oldPath); err != nil {
			return err
		}
		if err := modFile.AddRequire(rewrite.newPath, rewrite.version); err != nil {
			return err
		}
		changed = true
	}

	type replaceRewrite struct {
		oldPath    string
		oldVersion string
		newOldPath string
		newPath    string
		newVersion string
	}
	replaceRewrites := make([]replaceRewrite, 0)
	for _, rep := range modFile.Replace {
		rewritten := false
		newOldPath := rep.Old.Path
		if belongsToAppModulePrefix(rep.Old.Path, oldModulePath) {
			if rewrittenOldPath := renameAppModulePath(rep.Old.Path, oldModulePath, oldAppName, newAppName); rewrittenOldPath != rep.Old.Path {
				newOldPath = rewrittenOldPath
				rewritten = true
			}
		}

		newPath := rep.New.Path
		if rep.New.Version != "" {
			if belongsToAppModulePrefix(rep.New.Path, oldModulePath) {
				if rewrittenNewPath := renameAppModulePath(rep.New.Path, oldModulePath, oldAppName, newAppName); rewrittenNewPath != rep.New.Path {
					newPath = rewrittenNewPath
					rewritten = true
				}
			}
		} else if isModulePathReplacement(rep.New.Path) {
			if belongsToAppModulePrefix(rep.New.Path, oldModulePath) {
				if rewrittenNewPath := renameAppModulePath(rep.New.Path, oldModulePath, oldAppName, newAppName); rewrittenNewPath != rep.New.Path {
					newPath = rewrittenNewPath
					rewritten = true
				}
			}
		}

		if rewritten {
			replaceRewrites = append(replaceRewrites, replaceRewrite{
				oldPath:    rep.Old.Path,
				oldVersion: rep.Old.Version,
				newOldPath: newOldPath,
				newPath:    newPath,
				newVersion: rep.New.Version,
			})
		}
	}
	for _, rewrite := range replaceRewrites {
		if err := modFile.DropReplace(rewrite.oldPath, rewrite.oldVersion); err != nil {
			return err
		}
		if err := modFile.AddReplace(rewrite.newOldPath, rewrite.oldVersion, rewrite.newPath, rewrite.newVersion); err != nil {
			return err
		}
		changed = true
	}

	if !changed {
		return nil
	}

	formatted, err := modFile.Format()
	if err != nil {
		return err
	}
	return os.WriteFile(goModPath, formatted, 0o644)
}

func belongsToAppModulePrefix(path, oldModulePath string) bool {
	return hasModulePathPrefix(path, oldModulePath)
}

func renameAppModulePath(modulePath, oldModulePath, oldAppName, newAppName string) string {
	if modulePath == "" || oldModulePath == "" || oldAppName == "" || oldAppName == newAppName {
		return modulePath
	}
	if !belongsToAppModulePrefix(modulePath, oldModulePath) {
		return modulePath
	}

	parts := strings.Split(modulePath, "/")
	for i := len(parts) - 1; i >= 0; i-- {
		part := parts[i]
		if part == oldAppName {
			parts[i] = newAppName
			return strings.Join(parts, "/")
		}
	}

	return modulePath
}

func buildModulePathMappings(dstDir, oldProjectName, newProjectName string) (map[string]string, error) {
	mappings := make(map[string]string)
	rootMappings, err := buildRootModulePathMappings(dstDir, newProjectName)
	if err != nil {
		return nil, err
	}
	for oldPath, newPath := range rootMappings {
		mappings[oldPath] = newPath
	}

	var rootOldModulePath string
	var rootNewModulePath string
	for oldPath, newPath := range rootMappings {
		rootOldModulePath = oldPath
		rootNewModulePath = newPath
		break
	}

	type moduleInfo struct {
		relDir     string
		modulePath string
	}
	modules := make([]moduleInfo, 0)
	err = filepath.Walk(dstDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Base(path) != "go.mod" {
			return nil
		}

		modulePath, err := readModulePath(path)
		if err != nil {
			return err
		}

		relDir, err := filepath.Rel(dstDir, filepath.Dir(path))
		if err != nil {
			return err
		}
		modules = append(modules, moduleInfo{relDir: relDir, modulePath: modulePath})
		return nil
	})
	if err != nil {
		return nil, err
	}

	if rootOldModulePath == "" {
		for _, module := range modules {
			derivedRootPath, ok := deriveWorkspaceRootModulePath(module.modulePath, module.relDir)
			if !ok {
				continue
			}
			rootOldModulePath = derivedRootPath
			rootNewModulePath = buildRenamedModulePath(rootOldModulePath, newProjectName)
			mappings[rootOldModulePath] = rootNewModulePath
			break
		}
	}

	for _, module := range modules {
		if module.relDir == "." {
			continue
		}

		if rootOldModulePath != "" && hasModulePathPrefix(module.modulePath, rootOldModulePath) {
			relModulePath := strings.TrimPrefix(module.modulePath, rootOldModulePath)
			if relModulePath == "" {
				mappings[module.modulePath] = rootNewModulePath
				continue
			}
			if strings.HasPrefix(relModulePath, "/") {
				mappings[module.modulePath] = rootNewModulePath + relModulePath
				continue
			}
		}

		newModulePath, ok := deriveWorkspaceModulePath(module.modulePath, module.relDir, newProjectName)
		if ok {
			mappings[module.modulePath] = newModulePath
		}
	}

	return mappings, nil
}

func deriveWorkspaceRootModulePath(modulePath, relDir string) (string, bool) {
	relModulePath := filepath.ToSlash(filepath.Clean(relDir))
	if relModulePath == "." {
		return "", false
	}
	suffix := "/" + relModulePath
	if !strings.HasSuffix(modulePath, suffix) {
		return "", false
	}
	return strings.TrimSuffix(modulePath, suffix), true
}

func deriveWorkspaceModulePath(oldModulePath, relDir, newProjectName string) (string, bool) {
	relModulePath := filepath.ToSlash(filepath.Clean(relDir))
	if relModulePath == "." {
		return "", false
	}
	idx := strings.LastIndex(oldModulePath, "/"+relModulePath)
	if idx == -1 || idx+1+len(relModulePath) != len(oldModulePath) {
		return "", false
	}
	basePath := oldModulePath[:idx]
	if basePath == "" {
		return "", false
	}
	return basePath + "/" + filepath.ToSlash(filepath.Join(newProjectName, relDir)), true
}

func buildRootModulePathMappings(dstDir, newProjectName string) (map[string]string, error) {
	goModPath := filepath.Join(dstDir, "go.mod")
	if _, err := os.Stat(goModPath); err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}

	oldModulePath, err := readModulePath(goModPath)
	if err != nil {
		return nil, err
	}
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return nil, err
	}
	modFile, err := modfile.Parse(goModPath, content, nil)
	if err != nil {
		return nil, err
	}
	if err := modFile.AddModuleStmt(buildRenamedModulePath(oldModulePath, newProjectName)); err != nil {
		return nil, err
	}
	formatted, err := modFile.Format()
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(goModPath, formatted, 0o644); err != nil {
		return nil, err
	}

	newModulePath, err := readModulePath(goModPath)
	if err != nil {
		return nil, err
	}
	return map[string]string{oldModulePath: newModulePath}, nil
}

func buildRenamedModulePath(oldModulePath, newProjectName string) string {
	if strings.Contains(oldModulePath, "/") {
		lastSlash := strings.LastIndex(oldModulePath, "/")
		return oldModulePath[:lastSlash+1] + newProjectName
	}
	return newProjectName
}

func maybeModifyWorkspaceGoMods(dstDir, oldProjectName, newProjectName string) error {
	mappings, err := buildModulePathMappings(dstDir, oldProjectName, newProjectName)
	if err != nil {
		return err
	}

	return filepath.Walk(dstDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Base(path) != "go.mod" {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		modFile, err := modfile.Parse(path, content, nil)
		if err != nil {
			return err
		}

		changed, err := rewriteWorkspaceGoMod(modFile, mappings)
		if err != nil {
			return err
		}
		if !changed {
			return nil
		}

		formatted, err := modFile.Format()
		if err != nil {
			return err
		}
		return os.WriteFile(path, formatted, 0o644)
	})
}

func rewriteWorkspaceGoMod(modFile *modfile.File, mappings map[string]string) (bool, error) {
	changed := false

	if modFile.Module != nil {
		if newModulePath, ok := rewriteModulePathForProjectClone(modFile.Module.Mod.Path, mappings); ok {
			if err := modFile.AddModuleStmt(newModulePath); err != nil {
				return false, err
			}
			changed = true
		}
	}

	type requireRewrite struct {
		oldPath string
		newPath string
		version string
	}
	requireRewrites := make([]requireRewrite, 0)
	for _, req := range modFile.Require {
		if newModulePath, ok := rewriteModulePathForProjectClone(req.Mod.Path, mappings); ok {
			requireRewrites = append(requireRewrites, requireRewrite{
				oldPath: req.Mod.Path,
				newPath: newModulePath,
				version: req.Mod.Version,
			})
		}
	}
	for _, rewrite := range requireRewrites {
		if err := modFile.DropRequire(rewrite.oldPath); err != nil {
			return false, err
		}
		if err := modFile.AddRequire(rewrite.newPath, rewrite.version); err != nil {
			return false, err
		}
		changed = true
	}

	type replaceRewrite struct {
		oldPath    string
		oldVersion string
		newOldPath string
		newPath    string
		newVersion string
	}
	replaceRewrites := make([]replaceRewrite, 0)
	for _, rep := range modFile.Replace {
		rewritten := false
		newOldPath := rep.Old.Path
		if rewrittenOldPath, ok := rewriteModulePathForProjectClone(rep.Old.Path, mappings); ok {
			newOldPath = rewrittenOldPath
			rewritten = true
		}

		newPath := rep.New.Path
		if rep.New.Version == "" {
			if isModulePathReplacement(rep.New.Path) {
				if newNewPath, ok := rewriteModulePathForProjectClone(rep.New.Path, mappings); ok {
					newPath = newNewPath
					rewritten = true
				}
			}
			if rewritten {
				replaceRewrites = append(replaceRewrites, replaceRewrite{
					oldPath:    rep.Old.Path,
					oldVersion: rep.Old.Version,
					newOldPath: newOldPath,
					newPath:    newPath,
					newVersion: rep.New.Version,
				})
			}
			continue
		}
		if newNewPath, ok := rewriteModulePathForProjectClone(rep.New.Path, mappings); ok {
			newPath = newNewPath
			rewritten = true
		}
		if rewritten {
			replaceRewrites = append(replaceRewrites, replaceRewrite{
				oldPath:    rep.Old.Path,
				oldVersion: rep.Old.Version,
				newOldPath: newOldPath,
				newPath:    newPath,
				newVersion: rep.New.Version,
			})
		}
	}
	for _, rewrite := range replaceRewrites {
		if err := modFile.DropReplace(rewrite.oldPath, rewrite.oldVersion); err != nil {
			return false, err
		}
		if err := modFile.AddReplace(rewrite.newOldPath, rewrite.oldVersion, rewrite.newPath, rewrite.newVersion); err != nil {
			return false, err
		}
		changed = true
	}

	return changed, nil
}

func rewriteModulePathForProjectClone(modulePath string, mappings map[string]string) (string, bool) {
	bestOldPath := ""
	bestNewPath := ""
	for oldPath, newPath := range mappings {
		if !hasModulePathPrefix(modulePath, oldPath) {
			continue
		}
		if len(oldPath) > len(bestOldPath) {
			bestOldPath = oldPath
			bestNewPath = newPath
		}
	}
	if bestOldPath == "" {
		return modulePath, false
	}
	if modulePath == bestOldPath {
		return bestNewPath, true
	}
	return bestNewPath + strings.TrimPrefix(modulePath, bestOldPath), true
}

// hasModulePathPrefix 判断 path 是否等于 prefix 或以 prefix/ 开头（实现见 internal/scaffold.HasModulePathPrefix）
func hasModulePathPrefix(path, prefix string) bool {
	return scaffold.HasModulePathPrefix(path, prefix)
}

func isModulePathReplacement(path string) bool {
	if path == "" {
		return false
	}
	if strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../") || strings.HasPrefix(path, "/") {
		return false
	}
	if len(path) >= 2 && ((path[0] >= 'A' && path[0] <= 'Z') || (path[0] >= 'a' && path[0] <= 'z')) && path[1] == ':' {
		return false
	}
	return true
}

func isPathWithinDir(baseDir, targetPath string) (bool, error) {
	resolvedBaseDir, err := resolvePathForComparison(baseDir)
	if err != nil {
		return false, err
	}
	resolvedTargetPath, err := resolvePathForComparison(targetPath)
	if err != nil {
		return false, err
	}
	return dirContains(resolvedBaseDir, resolvedTargetPath), nil
}

func resolvePathForComparison(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	resolvedParent, err := resolveExistingDir(filepath.Dir(absPath))
	if err != nil {
		return "", err
	}
	return filepath.Join(resolvedParent, filepath.Base(absPath)), nil
}

func resolveExistingDir(path string) (string, error) {
	cleanPath := filepath.Clean(path)
	if _, err := os.Stat(cleanPath); err == nil {
		return filepath.EvalSymlinks(cleanPath)
	} else if !os.IsNotExist(err) {
		return "", err
	}

	parent := filepath.Dir(cleanPath)
	if parent == cleanPath {
		return cleanPath, nil
	}

	resolvedParent, err := resolveExistingDir(parent)
	if err != nil {
		return "", err
	}
	return filepath.Join(resolvedParent, filepath.Base(cleanPath)), nil
}

func detectProjectContext(currentDir string) (projectContext, error) {
	dir := currentDir
	var nearestGoModPath string
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if nearestGoModPath == "" {
			if _, err := os.Stat(goModPath); err == nil {
				nearestGoModPath = goModPath
			}
		}

		goWorkPath := filepath.Join(dir, "go.work")
		if _, err := os.Stat(goWorkPath); err == nil {
			goWorkUseDirs, err := parseGoWorkUseDirs(goWorkPath)
			if err != nil {
				return projectContext{}, err
			}
			if nearestGoModPath != "" && !pathInDirs(filepath.Dir(nearestGoModPath), goWorkUseDirs) {
				nearestGoModPath = ""
			}
			ctx := projectContext{
				rootDir:       dir,
				goModPath:     nearestGoModPath,
				goWorkPath:    goWorkPath,
				goWorkUseDirs: goWorkUseDirs,
			}
			return ctx, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	if nearestGoModPath != "" {
		return projectContext{
			rootDir:   filepath.Dir(nearestGoModPath),
			goModPath: nearestGoModPath,
		}, nil
	}

	return projectContext{}, fmt.Errorf("current directory is not a Go project root: no go.mod or go.work found")
}

func parseGoWorkUseDirs(goWorkPath string) ([]string, error) {
	content, err := os.ReadFile(goWorkPath)
	if err != nil {
		return nil, err
	}

	workFile, err := modfile.ParseWork(goWorkPath, content, nil)
	if err != nil {
		return nil, err
	}

	useDirs := make([]string, 0, len(workFile.Use))
	goWorkDir := filepath.Dir(goWorkPath)
	for _, use := range workFile.Use {
		useDir, err := filepath.Abs(filepath.Join(goWorkDir, use.Path))
		if err != nil {
			return nil, err
		}
		useDirs = append(useDirs, filepath.Clean(useDir))
	}

	return useDirs, nil
}

func pathInDirs(path string, dirs []string) bool {
	cleanPath := filepath.Clean(path)
	for _, dir := range dirs {
		if cleanPath == dir {
			return true
		}
	}
	return false
}

func resolveAppModulePath(ctx projectContext, sourceAppName string) (resolvedModule, error) {
	appDir := filepath.Join(ctx.appsDir(), sourceAppName)
	appGoModPath := filepath.Join(appDir, "go.mod")
	if _, err := os.Stat(appGoModPath); err == nil {
		modulePath, err := readModulePath(appGoModPath)
		if err != nil {
			return resolvedModule{}, err
		}
		return resolvedModule{
			rootDir:         appDir,
			goModPath:       appGoModPath,
			modulePath:      modulePath,
			oldImportPrefix: modulePath,
		}, nil
	}

	if rootModule, ok, err := resolveRootModule(ctx, appDir); err != nil {
		return resolvedModule{}, err
	} else if ok {
		return rootModule, nil
	}

	if len(ctx.goWorkUseDirs) == 0 {
		return resolvedModule{}, fmt.Errorf("go.work found, but cannot resolve module for apps/%s", sourceAppName)
	}

	if len(ctx.goWorkUseDirs) == 1 {
		if moduleCoversAppDir(ctx.goWorkUseDirs[0], appDir) {
			return resolveModuleFromDir(ctx.goWorkUseDirs[0], appDir)
		}
		return resolvedModule{}, fmt.Errorf("go.work found, but cannot resolve module for apps/%s", sourceAppName)
	}

	candidates := make([]string, 0, len(ctx.goWorkUseDirs))
	for _, useDir := range ctx.goWorkUseDirs {
		if moduleCoversAppDir(useDir, appDir) {
			candidates = append(candidates, useDir)
		}
	}

	if len(candidates) == 1 {
		return resolveModuleFromDir(candidates[0], appDir)
	}

	return resolvedModule{}, fmt.Errorf("go.work found, but cannot resolve module for apps/%s", sourceAppName)
}

func resolveRootModule(ctx projectContext, appDir string) (resolvedModule, bool, error) {
	rootGoModPath := filepath.Join(ctx.rootDir, "go.mod")
	if _, err := os.Stat(rootGoModPath); err != nil {
		if os.IsNotExist(err) {
			return resolvedModule{}, false, nil
		}
		return resolvedModule{}, false, err
	}

	if !moduleCoversAppDir(ctx.rootDir, appDir) {
		return resolvedModule{}, false, nil
	}

	modulePath, err := readModulePath(rootGoModPath)
	if err != nil {
		return resolvedModule{}, false, err
	}

	return resolvedModule{
		rootDir:         ctx.rootDir,
		goModPath:       rootGoModPath,
		modulePath:      modulePath,
		oldImportPrefix: buildOldImportPrefix(modulePath, ctx.rootDir, appDir),
	}, true, nil
}

func resolveModuleFromDir(moduleDir, appDir string) (resolvedModule, error) {
	goModPath := filepath.Join(moduleDir, "go.mod")
	modulePath, err := readModulePath(goModPath)
	if err != nil {
		return resolvedModule{}, err
	}
	return resolvedModule{
		rootDir:         moduleDir,
		goModPath:       goModPath,
		modulePath:      modulePath,
		oldImportPrefix: buildOldImportPrefix(modulePath, moduleDir, appDir),
	}, nil
}

func buildOldImportPrefix(modulePath, moduleDir, appDir string) string {
	relPath, err := filepath.Rel(filepath.Clean(moduleDir), filepath.Clean(appDir))
	if err != nil || relPath == "." {
		return modulePath
	}
	return modulePath + "/" + filepath.ToSlash(relPath)
}

func dirContains(baseDir, targetDir string) bool {
	rel, err := filepath.Rel(filepath.Clean(baseDir), filepath.Clean(targetDir))
	if err != nil {
		return false
	}
	return rel == "." || (rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)))
}

func moduleCoversAppDir(moduleDir, appDir string) bool {
	cleanModuleDir := filepath.Clean(moduleDir)
	cleanAppDir := filepath.Clean(appDir)
	return cleanModuleDir == cleanAppDir || dirContains(cleanModuleDir, cleanAppDir)
}

// findProjectRoot 查找项目根目录
// 1. 如果存在 go.work，使用 workspace 模式
// 2. 否则向上遍历找 go.mod
// 返回: rootDir, isWorkspace, modulePath, error
func findProjectRoot(currentDir string) (string, bool, string, error) {
	ctx, err := detectProjectContext(currentDir)
	if err != nil {
		return "", false, "", err
	}

	if ctx.goWorkPath != "" {
		if ctx.goModPath == "" {
			return "", false, "", fmt.Errorf("workspace found but no matching module root found for current directory")
		}

		modulePath, err := readModulePath(ctx.goModPath)
		if err != nil {
			return "", false, "", err
		}
		return ctx.rootDir, true, modulePath, nil
	}

	modulePath, err := readModulePath(ctx.goModPath)
	if err != nil {
		return "", false, "", err
	}
	return ctx.rootDir, false, modulePath, nil
}
