package scaffold

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"strings"
)

// RewriteGoContent 解析 Go 源码并重写：
//   - 若 oldAppName 非空，包名中出现 oldAppName 的位置替换为 newAppName
//   - import 路径按 prefixMappings 做最长前缀匹配替换
//
// 重写后重新格式化；未发生任何改动时返回原始内容。
func RewriteGoContent(filename string, content []byte, oldAppName, newAppName string, prefixMappings map[string]string) ([]byte, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, filename, content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse file %s fail: %w", filename, err)
	}

	changed := false
	if oldAppName != "" && node.Name != nil && strings.Contains(node.Name.Name, oldAppName) {
		node.Name.Name = strings.Replace(node.Name.Name, oldAppName, newAppName, -1)
		changed = true
	}

	ast.Inspect(node, func(n ast.Node) bool {
		importSpec, ok := n.(*ast.ImportSpec)
		if !ok {
			return true
		}
		importPath := strings.Trim(importSpec.Path.Value, `"`)
		updatedPath, ok := RewriteModulePrefix(importPath, prefixMappings)
		if !ok {
			return true
		}
		importSpec.Path.Value = fmt.Sprintf("%q", updatedPath)
		changed = true
		return true
	})

	if !changed {
		return content, nil
	}

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, node); err != nil {
		return nil, fmt.Errorf("format file %s fail: %w", filename, err)
	}
	return buf.Bytes(), nil
}

// RewriteModulePrefix 按映射对模块路径做最长前缀匹配替换。
// 例如 mappings 为 {github.com/example: github.com/acme/backend} 时：
// github.com/example/pkg -> github.com/acme/backend/pkg
func RewriteModulePrefix(modulePath string, mappings map[string]string) (string, bool) {
	bestOld := ""
	bestNew := ""
	for oldPath, newPath := range mappings {
		if !HasModulePathPrefix(modulePath, oldPath) {
			continue
		}
		if len(oldPath) > len(bestOld) {
			bestOld = oldPath
			bestNew = newPath
		}
	}
	if bestOld == "" {
		return modulePath, false
	}
	if modulePath == bestOld {
		return bestNew, true
	}
	return bestNew + strings.TrimPrefix(modulePath, bestOld), true
}

// HasModulePathPrefix 判断 path 是否等于 prefix 或以 prefix/ 开头。
func HasModulePathPrefix(path, prefix string) bool {
	return path == prefix || strings.HasPrefix(path, prefix+"/")
}
