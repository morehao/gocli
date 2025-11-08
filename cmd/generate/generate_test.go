package generate

import (
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
	// 执行命令，需要指定 app 名称
	_, err := ExecuteCommand(Cmd, "--mode", "model", "--app", "demoapp")
	if err != nil {
		t.Errorf("Failed to execute command with config: %v", err)
	}
	t.Log(gutils.ToJsonString(cfg))
}
