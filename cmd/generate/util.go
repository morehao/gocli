package generate

import (
	"bufio"
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

const (
	TplFuncIsBuiltInField      = "isBuiltInField"
	TplFuncIsSysField          = "isSysField"
	TplFuncIsDefaultModelLayer = "isDefaultModelLayer"
	TplFuncIsDefaultDaoLayer   = "isDefaultDaoLayer"
)

func IsBuiltInField(name string) bool {
	buildInFieldMap := map[string]struct{}{
		"ID":        {},
		"CreatedAt": {},
		"UpdatedAt": {},
		"DeletedAt": {},
	}
	_, ok := buildInFieldMap[name]
	return ok
}

func IsSysField(name string) bool {
	sysFieldMap := map[string]struct{}{
		"ID":        {},
		"CreatedAt": {},
		"CreatedBy": {},
		"UpdatedAt": {},
		"UpdatedBy": {},
		"DeletedAt": {},
		"DeletedBy": {},
	}
	_, ok := sysFieldMap[name]
	return ok
}

func IsDefaultModelLayer(name string) bool {
	return name == "model"
}

func IsDefaultDaoLayer(name string) bool {
	return name == "dao"
}

// RemoveTablePrefixFromStructName 从结构体名中去除表名前缀
// 例如：表名 iam_users，前缀 iam_，结构体名 IamUsers -> Users
// 参数：
//   - structName: 原始结构体名（如 IamUsers）
//   - tableName: 原始表名（如 iam_users）
//   - prefix: 要去除的前缀（如 iam_）
// 返回：去除前缀后的结构体名（如 Users）
func RemoveTablePrefixFromStructName(structName, tableName, prefix string) string {
	if prefix == "" {
		return structName
	}
	
	// 如果表名以指定前缀开头，则从结构体名中去除对应的前缀部分
	if strings.HasPrefix(tableName, prefix) {
		// 将前缀转换为对应的结构体名格式（去除下划线，每个单词首字母大写）
		// 例如：iam_ -> Iam, sys_ -> Sys
		prefixWithoutUnderscore := strings.TrimSuffix(prefix, "_")
		if prefixWithoutUnderscore == "" {
			return structName
		}
		
		// 将前缀转换为 PascalCase
		prefixParts := strings.Split(prefixWithoutUnderscore, "_")
		var prefixStructName string
		for _, part := range prefixParts {
			if part != "" {
				// 将每个部分的首字母大写，其余小写
				if len(part) > 0 {
					prefixStructName += strings.ToUpper(part[:1]) + strings.ToLower(part[1:])
				}
			}
		}
		
		// 如果结构体名以此前缀开头，则去除
		if strings.HasPrefix(structName, prefixStructName) {
			remaining := strings.TrimPrefix(structName, prefixStructName)
			// 确保剩余部分首字母大写
			if remaining != "" {
				return strings.ToUpper(remaining[:1]) + remaining[1:]
			}
			// 如果去除前缀后为空，返回原结构体名（这种情况不应该发生，但为了安全）
			return structName
		}
	}
	
	return structName
}

// RemoveTablePrefixFromFilename 从文件名中去除表名前缀
// 例如：文件名 iam_user.go，前缀 iam_，处理后 -> user.go
// 参数：
//   - filename: 原始文件名（如 iam_user.go）
//   - tableName: 原始表名（如 iam_users）
//   - prefix: 要去除的前缀（如 iam_）
// 返回：去除前缀后的文件名（如 user.go）
func RemoveTablePrefixFromFilename(filename, tableName, prefix string) string {
	if prefix == "" {
		return filename
	}
	
	// 如果表名以指定前缀开头，则从文件名中去除对应的前缀部分
	if strings.HasPrefix(tableName, prefix) {
		// 分离文件名和扩展名
		ext := filepath.Ext(filename)
		nameWithoutExt := strings.TrimSuffix(filename, ext)
		
		// 将前缀转换为文件名格式（去除下划线）
		prefixWithoutUnderscore := strings.TrimSuffix(prefix, "_")
		if prefixWithoutUnderscore == "" {
			return filename
		}
		
		// 构建前缀在文件名中的形式（snake_case）
		prefixInFilename := prefixWithoutUnderscore + "_"
		
		// 如果文件名以此前缀开头，则去除
		if strings.HasPrefix(nameWithoutExt, prefixInFilename) {
			remaining := strings.TrimPrefix(nameWithoutExt, prefixInFilename)
			if remaining != "" {
				return remaining + ext
			}
		}
	}
	
	return filename
}

// CopyEmbeddedTemplatesToTempDir 将嵌入的模板文件复制到临时目录，并返回该目录的路径。
func CopyEmbeddedTemplatesToTempDir(embeddedFS embed.FS, root string) (string, error) {
	// 创建一个临时目录来存放模板文件
	tempDir, err := os.MkdirTemp("", "codegen_templates")
	if err != nil {
		return "", fmt.Errorf("failed to create temp directory: %v", err)
	}

	// 将嵌入的模板文件复制到临时目录
	err = fs.WalkDir(embeddedFS, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			data, readErr := embeddedFS.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			// 保持目录结构
			relPath, relErr := filepath.Rel(root, path)
			if relErr != nil {
				return relErr
			}
			targetPath := filepath.Join(tempDir, relPath)
			if mkDirErr := os.MkdirAll(filepath.Dir(targetPath), 0755); mkDirErr != nil {
				return mkDirErr
			}
			if writeErr := os.WriteFile(targetPath, data, 0644); writeErr != nil {
				return writeErr
			}
		}
		return nil
	})
	if err != nil {
		// 如果复制失败，清理临时目录
		os.RemoveAll(tempDir)
		return "", fmt.Errorf("failed to copy templates: %v", err)
	}

	return tempDir, nil
}

// GetAppInfo 应用模块路径信息
// 输入示例：/Users/morehao/xxx/go-gin-web/apps/demoapp
// 或者：/Users/morehao/xxx/gocli/cmd/generate/_example/apps/demoapp
func GetAppInfo(workDir string) (*AppInfo, error) {
	cleanPath := filepath.Clean(workDir)
	segments := strings.Split(cleanPath, string(filepath.Separator))

	// 查找 "apps/{appName}" 结构
	var appsIndex = -1
	for i := 0; i < len(segments)-1; i++ {
		if segments[i] == "apps" {
			appsIndex = i
			break
		}
	}
	if appsIndex == -1 {
		return nil, fmt.Errorf("invalid structure: path does not contain /apps/{appName}")
	}

	// apps 目录前面至少需要有一个父级目录（projectName）
	if appsIndex < 1 {
		return nil, fmt.Errorf("invalid structure: apps directory must have at least one parent directory")
	}

	// 解析 app 名称
	appName := segments[appsIndex+1]

	// 解析项目名和相对路径
	// 项目名是 apps 的直接父目录
	projectName := segments[appsIndex-1]

	// 构建从项目根到app的相对路径
	// 如果 apps 直接在项目根下：apps/demoapp
	appPathInProject := filepath.Join("apps", appName)

	// 获取项目根目录的绝对路径
	// 项目根目录是 apps 的父目录
	projectRootPath := filepath.Join(segments[:appsIndex]...)
	if len(projectRootPath) == 0 {
		projectRootPath = string(filepath.Separator)
	} else {
		projectRootPath = string(filepath.Separator) + projectRootPath
	}

	// 读取 go.mod 文件获取模块路径
	modulePath, err := getModulePath(projectRootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get module path: %v", err)
	}

	return &AppInfo{
		AppPathInProject: appPathInProject,
		ProjectName:      projectName,
		AppName:          appName,
		ProjectRootPath:  projectRootPath,
		ModulePath:       modulePath,
	}, nil
}

// getModulePath 从 go.mod 文件中读取模块路径
func getModulePath(projectRootPath string) (string, error) {
	goModPath := filepath.Join(projectRootPath, "go.mod")
	
	file, err := os.Open(goModPath)
	if err != nil {
		return "", fmt.Errorf("failed to open go.mod: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			// 提取模块路径，格式：module github.com/morehao/go-gin-web
			modulePath := strings.TrimPrefix(line, "module ")
			modulePath = strings.TrimSpace(modulePath)
			return modulePath, nil
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading go.mod: %v", err)
	}

	return "", fmt.Errorf("module declaration not found in go.mod")
}

// ExecuteCommand 执行命令并捕获输出
func ExecuteCommand(root *cobra.Command, args ...string) (output string, err error) {
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetErr(buf)
	root.SetArgs(args)
	err = root.Execute()
	return buf.String(), err
}
