package generate

import (
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
	// 如果 apps 直接在项目根下：go-gin-web/apps/demoapp
	// 如果 apps 在子目录下：generate/_example/apps/demoapp
	appPathInProject := filepath.Join(projectName, "apps", appName)

	// 获取项目根目录的绝对路径
	// 项目根目录是 apps 的父目录
	projectRootPath := filepath.Join(segments[:appsIndex]...)
	if len(projectRootPath) == 0 {
		projectRootPath = string(filepath.Separator)
	} else {
		projectRootPath = string(filepath.Separator) + projectRootPath
	}

	return &AppInfo{
		AppPathInProject: appPathInProject,
		ProjectName:      projectName,
		AppName:          appName,
		ProjectRootPath:  projectRootPath,
	}, nil
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
