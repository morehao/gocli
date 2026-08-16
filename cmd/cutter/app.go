/*
 * @Author: morehao morehao@qq.com
 * @Date: 2025-11-08 22:32:22
 * @LastEditors: morehao morehao@qq.com
 * @LastEditTime: 2025-11-08 22:34:26
 * @FilePath: /golib/Users/morehao/Documents/practice/go/gocli/cmd/cutter/app.go
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
package cutter

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// cloneApp 在同一项目内克隆一个app
func cloneApp(sourceAppName, newAppName string) error {
	if sourceAppName == "" || newAppName == "" {
		return fmt.Errorf("source app name and new app name cannot be empty")
	}

	if sourceAppName == newAppName {
		return fmt.Errorf("source app name and new app name cannot be the same")
	}

	// 获取当前工作目录（应该是项目根目录）
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory fail: %w", err)
	}

	// 查找项目上下文并解析源 app 的模块信息
	ctx, err := detectProjectContext(currentDir)
	if err != nil {
		return fmt.Errorf("find project root fail: %w", err)
	}

	resolvedModule, err := resolveAppModulePath(ctx, sourceAppName)
	if err != nil {
		return fmt.Errorf("resolve source app module fail: %w", err)
	}

	fmt.Printf("Project root: %s (workspace: %v)\n", ctx.rootDir, ctx.goWorkPath != "")

	// 确认 apps 目录存在（兼容 ark 的 backend/apps 与经典 apps 两种结构）
	appsDir := ctx.appsDir()
	if _, err := os.Stat(appsDir); os.IsNotExist(err) {
		return fmt.Errorf("apps directory does not exist: %s", appsDir)
	}

	// 确认源 app 存在
	sourceAppDir := filepath.Join(appsDir, sourceAppName)
	if _, err := os.Stat(sourceAppDir); os.IsNotExist(err) {
		return fmt.Errorf("source app does not exist: %s", sourceAppDir)
	}

	// 确认新 app 不存在
	newAppDir := filepath.Join(appsDir, newAppName)
	if _, err := os.Stat(newAppDir); !os.IsNotExist(err) {
		return fmt.Errorf("new app already exists: %s", newAppDir)
	}

	// 创建新 app 目录
	if err := os.MkdirAll(newAppDir, os.ModePerm); err != nil {
		return fmt.Errorf("create new app directory fail: %w", err)
	}

	fmt.Printf("Cloning %s to %s...\n", sourceAppName, newAppName)

	// 复制并替换内容
	if err := copyAndReplaceApp(sourceAppDir, newAppDir, sourceAppName, newAppName, resolvedModule); err != nil {
		// 如果出错，清理已创建的目录
		os.RemoveAll(newAppDir)
		return fmt.Errorf("copy and replace app fail: %w", err)
	}

	if err := maybeModifyAppGoMod(newAppDir, sourceAppName, newAppName); err != nil {
		os.RemoveAll(newAppDir)
		return fmt.Errorf("modify app go.mod fail: %w", err)
	}

	return nil
}

// copyAndReplaceApp 复制app目录并替换相关的包名和import路径
func copyAndReplaceApp(srcDir, dstDir, oldAppName, newAppName string, resolvedModule resolvedModule) error {
	err := filepath.Walk(srcDir, func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 获取相对于源目录的路径
		relativePath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		// 检查是否应该忽略
		if relativePath != "." && shouldIgnore(relativePath) {
			fmt.Println("Excluding:", path)
			if fileInfo.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 创建目标路径
		targetPath := filepath.Join(dstDir, relativePath)

		// 如果是目录，创建目录
		if fileInfo.IsDir() {
			return os.MkdirAll(targetPath, fileInfo.Mode())
		}

		// 如果是 .go 文件，需要替换内容
		if strings.HasSuffix(fileInfo.Name(), ".go") {
			return copyAndReplaceGoFileInApp(path, targetPath, oldAppName, newAppName, resolvedModule)
		}

		// 如果是 .yaml 或 .yml 配置文件，也需要替换内容
		if strings.HasSuffix(fileInfo.Name(), ".yaml") || strings.HasSuffix(fileInfo.Name(), ".yml") {
			return copyAndReplaceTextFile(path, targetPath, oldAppName, newAppName)
		}

		// 其他文件直接复制
		return copyFile(path, targetPath)
	})
	return err
}

// copyAndReplaceGoFileInApp 复制并替换 Go 文件中的包名和 import 路径，
// 以及代码中对被重写 import 的包标识符引用。
func copyAndReplaceGoFileInApp(srcFile, dstFile, oldAppName, newAppName string, resolvedModule resolvedModule) error {
	fs := token.NewFileSet()
	node, err := parser.ParseFile(fs, srcFile, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse file %s fail: %w", srcFile, err)
	}

	// 替换包名中的 app 名称
	if node.Name != nil && strings.Contains(node.Name.Name, oldAppName) {
		node.Name.Name = strings.Replace(node.Name.Name, oldAppName, newAppName, -1)
	}

	// 收集被重写 import 的别名映射：旧别名 -> 新别名
	// import 默认别名取路径末段（如 .../fixapp -> fixapp）；显式别名用 importSpec.Name。
	// 重写路径后新天然小写末段为新 app 名；显式别名则按 oldAppName 同名替换。
	aliases := make(map[string]string)
	// 遍历文件中的所有 import 语句，替换路径中的 oldAppName 为 newAppName
	ast.Inspect(node, func(n ast.Node) bool {
		importSpec, ok := n.(*ast.ImportSpec)
		if !ok {
			return true
		}
		importPath := strings.Trim(importSpec.Path.Value, `"`)
		// 只替换当前 app 旧模块前缀
		if !hasModulePathPrefix(importPath, resolvedModule.oldImportPrefix) {
			return true
		}
		updatedImportPath := renameImportedAppPath(importPath, resolvedModule.oldImportPrefix, oldAppName, newAppName)
		importSpec.Path.Value = fmt.Sprintf(`"%s"`, updatedImportPath)

		if importSpec.Name != nil {
			alias := importSpec.Name.Name
			if alias != "_" && alias != "." && strings.Contains(alias, oldAppName) {
				newAlias := strings.Replace(alias, oldAppName, newAppName, -1)
				importSpec.Name.Name = newAlias
				aliases[alias] = newAlias
			}
			return true
		}
		// 只有 app 根模块（import 路径等于旧模块前缀，如 <module>/fixapp）的引用
		// 才需要在重写包名时同步替换代码中的包标识符；子包（如 objuser）保持自身包名。
		if importPath == resolvedModule.oldImportPrefix {
			oldAlias := pathBase(importPath)
			if oldAlias != "" && oldAlias != "_" {
				aliases[oldAlias] = newAppName
			}
		}
		return true
	})
	// 重写代码中对被替换 import 包名的引用（如 fixapp.Routers -> cutapp.Routers）
	if len(aliases) > 0 {
		ast.Inspect(node, func(n ast.Node) bool {
			sel, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			x, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}
			if newAlias, ok := aliases[x.Name]; ok {
				x.Name = newAlias
			}
			return true
		})
	}

	// 将更新后的代码写入目标文件
	file, err := os.Create(dstFile)
	if err != nil {
		return fmt.Errorf("create file %s fail: %w", dstFile, err)
	}
	defer file.Close()

	if err := format.Node(file, fs, node); err != nil {
		return fmt.Errorf("format and write file %s fail: %w", dstFile, err)
	}
	return nil
}

// pathBase 返回 import 路径的最后一段作为默认包名（去除路径）。
func pathBase(importPath string) string {
	if idx := strings.LastIndex(importPath, "/"); idx >= 0 {
		return importPath[idx+1:]
	}
	return importPath
}

func renameImportedAppPath(importPath, oldImportPrefix, oldAppName, newAppName string) string {
	if importPath == oldImportPrefix {
		return renameAppModulePath(importPath, oldImportPrefix, oldAppName, newAppName)
	}

	if strings.HasPrefix(importPath, oldImportPrefix+"/") {
		newPrefix := renameAppModulePath(oldImportPrefix, oldImportPrefix, oldAppName, newAppName)
		return newPrefix + strings.TrimPrefix(importPath, oldImportPrefix)
	}

	return importPath
}

// copyAndReplaceTextFile 复制并替换文本文件中的app名称
func copyAndReplaceTextFile(srcFile, dstFile, oldAppName, newAppName string) error {
	content, err := os.ReadFile(srcFile)
	if err != nil {
		return fmt.Errorf("read file %s fail: %w", srcFile, err)
	}

	// 替换所有出现的旧app名称
	newContent := strings.ReplaceAll(string(content), oldAppName, newAppName)

	// 写入新文件
	err = os.WriteFile(dstFile, []byte(newContent), 0644)
	if err != nil {
		return fmt.Errorf("write file %s fail: %w", dstFile, err)
	}
	return nil
}
