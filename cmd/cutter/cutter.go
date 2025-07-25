package cutter

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"

	ignore "github.com/morehao/go-gitignore"
	"github.com/spf13/cobra"
	"golang.org/x/mod/modfile"
)

// Cmd represents the base command when called without any subcommands
var Cmd = &cobra.Command{
	Use:   "cutter",
	Short: "A CLI for rapidly scaffolding Go projects with templates or by cloning existing structures.",
	Run: func(cmd *cobra.Command, args []string) {
		destination, err := cmd.Flags().GetString("destination")
		if err != nil {
			fmt.Println("Error retrieving destination flag:", err)
			os.Exit(1)
		}
		if destination == "" {
			fmt.Println("Destination directory is empty")
			os.Exit(1)
		}
		if err := cutter(destination); err != nil {
			fmt.Println("Error running cutter:", err)
			os.Exit(1)
		}
		fmt.Println("Successfully created Go project at", destination)
	},
}

func init() {
	Cmd.Flags().StringP("destination", "d", "", "Destination directory for the new project. For example: ./your/project/path")
}

func cutter(newProjectPath string) error {
	if newProjectPath == "" {
		return fmt.Errorf("new project path is empty")
	}

	// 获取当前执行目录，确认它是Go项目
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("get current directory fail, err: %w", err)
	}
	if !isGoProject(currentDir) {
		return fmt.Errorf("%s is not a Go project", currentDir)
	}

	// 获取模板项目名称
	templateName := filepath.Base(currentDir)
	newProjectName := filepath.Base(newProjectPath)

	// 确认新项目目录不存在或为空
	if _, err := os.Stat(newProjectPath); !os.IsNotExist(err) {
		return fmt.Errorf("%s already exists", newProjectPath)
	}

	// 创建新项目目录
	if err := os.MkdirAll(newProjectPath, os.ModePerm); err != nil {
		return fmt.Errorf("create new project directory: %w", err)
	}

	// 读取.gitignore文件
	gitignorePath := filepath.Join(currentDir, ".gitignore")
	gitignore, err := readGitignore(gitignorePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read .gitignore file fail, err: %w", err)
	}

	// 复制模板项目到新项目目录，并替换import路径
	if err := copyAndReplace(currentDir, newProjectPath, templateName, newProjectName, gitignore); err != nil {
		return fmt.Errorf("copy and replace fail, err: %w", err)
	}
	if err := removeGitDir(newProjectPath); err != nil {
		return fmt.Errorf("remove .git dir fail, err: %w", err)
	}
	return nil
}

// isGoProject 检查指定路径是否为Go项目（是否包含go.mod文件）
func isGoProject(path string) bool {
	_, err := os.Stat(filepath.Join(path, "go.mod"))
	return !os.IsNotExist(err)
}

// readGitignore 读取.gitignore文件并返回IgnoreParser
func readGitignore(filename string) (*ignore.GitIgnore, error) {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return ignore.CompileIgnoreLines(), nil
	}
	ig, err := ignore.CompileIgnoreFile(filename)
	if err != nil {
		return nil, err
	}
	return ig, nil
}

// copyAndReplace copy指定目录，并替换import路径
func copyAndReplace(srcDir, dstDir, oldName, newName string, ig *ignore.GitIgnore) error {
	err := filepath.Walk(srcDir, func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 获取相对于源目录的路径
		relativePath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}

		// 检查是否排除
		if ig.MatchesPath(relativePath) {
			fmt.Println("Excluding:", path)
			if fileInfo.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// 创建目标目录
		targetPath := strings.Replace(path, srcDir, dstDir, 1)
		if fileInfo.IsDir() {
			return os.MkdirAll(targetPath, fileInfo.Mode())
		}

		// 复制文件并替换 import 路径
		if strings.HasSuffix(fileInfo.Name(), ".go") {
			return copyAndReplaceGoFile(path, targetPath, oldName, newName)
		}

		// 复制其他文件
		return copyFile(path, targetPath)
	})
	if err != nil {
		return err
	}
	if err := modifyGoMod(dstDir, newName); err != nil {
		return err
	}
	return err
}

// copyAndReplaceGoFile 复制并替换 Go 文件中的 import 路径
func copyAndReplaceGoFile(srcFile, dstFile, oldName, newName string) error {
	fs := token.NewFileSet()
	node, err := parser.ParseFile(fs, srcFile, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	// 遍历文件中的所有 import 语句，替换路径中的 oldName 为 newName
	ast.Inspect(node, func(n ast.Node) bool {
		importSpec, ok := n.(*ast.ImportSpec)
		if ok {
			importPath := strings.Trim(importSpec.Path.Value, `"`)
			if strings.Contains(importPath, oldName) {
				updatedImportPath := strings.Replace(importPath, oldName, newName, -1)
				importSpec.Path.Value = fmt.Sprintf(`"%s"`, updatedImportPath)
			}
		}
		return true
	})

	// 将更新后的代码写入目标文件
	file, err := os.Create(dstFile)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := format.Node(file, fs, node); err != nil {
		return err
	}
	return nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

// 修改go.mod中的包名
func modifyGoMod(dstDir, moduleName string) error {
	// 读取go.mod文件
	modFilepath := filepath.Join(dstDir, "go.mod")
	content, err := os.ReadFile(modFilepath)
	if err != nil {
		return err
	}

	// 解析go.mod文件
	modFile, err := modfile.Parse(modFilepath, content, nil)
	if err != nil {
		return err
	}

	// 修改模块名称
	if err := modFile.AddModuleStmt(moduleName); err != nil {
		return err
	}

	// 将修改后的内容格式化回字节切片
	newContent, err := modFile.Format()
	if err != nil {
		return err
	}

	// 写入新的go.mod文件
	err = os.WriteFile(modFilepath, newContent, 0644)
	if err != nil {
		return err
	}
	return nil
}

// 删除.git目录
// removeGitDir 删除指定目录下的.git文件夹
func removeGitDir(dstDir string) error {
	gitDir := filepath.Join(dstDir, ".git")
	err := os.RemoveAll(gitDir)
	if err != nil {
		return err
	}
	return nil
}
