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

	"golang.org/x/mod/modfile"
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

	// 确认是 Go 项目
	if !isGoProject(currentDir) {
		return fmt.Errorf("%s is not a Go project (no go.mod found)", currentDir)
	}

	// 读取 go.mod 获取模块名
	modFilePath := filepath.Join(currentDir, "go.mod")
	modContent, err := os.ReadFile(modFilePath)
	if err != nil {
		return fmt.Errorf("read go.mod fail: %w", err)
	}
	modFile, err := modfile.Parse(modFilePath, modContent, nil)
	if err != nil {
		return fmt.Errorf("parse go.mod fail: %w", err)
	}
	modulePath := modFile.Module.Mod.Path

	// 确认 apps 目录存在
	appsDir := filepath.Join(currentDir, "apps")
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
	if err := copyAndReplaceApp(sourceAppDir, newAppDir, sourceAppName, newAppName, modulePath); err != nil {
		// 如果出错，清理已创建的目录
		os.RemoveAll(newAppDir)
		return fmt.Errorf("copy and replace app fail: %w", err)
	}

	return nil
}

// copyAndReplaceApp 复制app目录并替换相关的包名和import路径
func copyAndReplaceApp(srcDir, dstDir, oldAppName, newAppName, modulePath string) error {
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
			return copyAndReplaceGoFileInApp(path, targetPath, oldAppName, newAppName, modulePath)
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

// copyAndReplaceGoFileInApp 复制并替换 Go 文件中的包名和 import 路径
func copyAndReplaceGoFileInApp(srcFile, dstFile, oldAppName, newAppName, modulePath string) error {
	fs := token.NewFileSet()
	node, err := parser.ParseFile(fs, srcFile, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("parse file %s fail: %w", srcFile, err)
	}

	// 替换包名中的 app 名称
	if node.Name != nil && strings.Contains(node.Name.Name, oldAppName) {
		node.Name.Name = strings.Replace(node.Name.Name, oldAppName, newAppName, -1)
	}

	// 遍历文件中的所有 import 语句，替换路径中的 oldAppName 为 newAppName
	ast.Inspect(node, func(n ast.Node) bool {
		importSpec, ok := n.(*ast.ImportSpec)
		if ok {
			importPath := strings.Trim(importSpec.Path.Value, `"`)
			// 只替换包含当前模块路径和旧app名称的import
			if strings.Contains(importPath, modulePath+"/apps/"+oldAppName) {
				updatedImportPath := strings.Replace(importPath, "/apps/"+oldAppName, "/apps/"+newAppName, -1)
				importSpec.Path.Value = fmt.Sprintf(`"%s"`, updatedImportPath)
			}
		}
		return true
	})

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
