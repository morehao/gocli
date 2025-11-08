package generate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/morehao/golib/gutils"
)

// 测试模板文件加载
func TestLoadTemplates(t *testing.T) {
	// 测试模板目录是否存在
	dirs := []string{"template/module", "template/model", "template/api"}
	for _, dir := range dirs {
		entries, err := TemplatesFS.ReadDir(dir)
		if err != nil {
			t.Errorf("Failed to read directory %s: %v", dir, err)
			continue
		}
		if len(entries) == 0 {
			t.Errorf("Directory %s is empty", dir)
		}
		t.Logf("Directory %s is not empty", dir)
	}
}

// 测试配置加载
func TestConfigLoading(t *testing.T) {
	// 切换到 example 目录作为项目根目录
	exampleDir, _ := filepath.Abs("example")
	originalDir, _ := os.Getwd()
	if err := os.Chdir(exampleDir); err != nil {
		t.Skipf("Skipping test: example directory not found: %v", err)
		return
	}
	defer os.Chdir(originalDir)

	// 执行命令，需要指定 app 名称
	_, err := ExecuteCommand(Cmd, "--mode", "model", "--app", "demoapp")
	if err != nil {
		t.Errorf("Failed to execute command with config: %v", err)
	}
	t.Log(gutils.ToJsonString(cfg))
}

// TestGenerateModelCode 测试生成 model 层代码
// 运行前请确保：
// 1. MySQL 数据库可访问
// 2. example/apps/demoapp/config/code_gen.yaml 中的配置正确
// 3. example 目录存在并包含完整的示例项目结构
func TestGenerateModelCode(t *testing.T) {
	// 切换到 example 目录作为项目根目录
	exampleDir, _ := filepath.Abs("example")
	originalDir, _ := os.Getwd()
	if err := os.Chdir(exampleDir); err != nil {
		t.Skipf("Skipping test: example directory not found: %v", err)
		return
	}
	defer os.Chdir(originalDir)

	_, err := ExecuteCommand(Cmd, "--mode", "model", "--app", "demoapp")
	if err != nil {
		t.Errorf("Failed to execute model command: %v", err)
	}
}

// TestGenerateModuleCode 测试生成完整模块代码
// 运行前请确保：
// 1. MySQL 数据库可访问
// 2. example/apps/demoapp/config/code_gen.yaml 中的配置正确
// 3. example 目录存在并包含完整的示例项目结构
func TestGenerateModuleCode(t *testing.T) {
	exampleDir, _ := filepath.Abs("example")
	originalDir, _ := os.Getwd()
	if err := os.Chdir(exampleDir); err != nil {
		t.Skipf("Skipping test: example directory not found: %v", err)
		return
	}
	defer os.Chdir(originalDir)

	_, err := ExecuteCommand(Cmd, "--mode", "module", "--app", "demoapp")
	if err != nil {
		t.Errorf("Failed to execute module command: %v", err)
	}
}

// TestGenerateApiCode 测试生成 API 代码
// 运行前请确保：
// 1. MySQL 数据库可访问
// 2. example/apps/demoapp/config/code_gen.yaml 中的配置正确
// 3. example 目录存在并包含完整的示例项目结构
func TestGenerateApiCode(t *testing.T) {
	exampleDir, _ := filepath.Abs("example")
	originalDir, _ := os.Getwd()
	if err := os.Chdir(exampleDir); err != nil {
		t.Skipf("Skipping test: example directory not found: %v", err)
		return
	}
	defer os.Chdir(originalDir)

	_, err := ExecuteCommand(Cmd, "--mode", "api", "--app", "demoapp")
	if err != nil {
		t.Errorf("Failed to execute api command: %v", err)
	}
}
