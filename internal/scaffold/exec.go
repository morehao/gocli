package scaffold

import (
	"fmt"
	"os"
	"os/exec"
)

// RunCmd 在 dir 目录下执行命令，输出透传到终端。
func RunCmd(dir, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RunGoWorkSync 在 dir 下执行 go work sync，用于同步 workspace 依赖。
// 失败时返回错误（由调用方决定是否降级为提示）。
func RunGoWorkSync(dir string) error {
	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("go command not found in PATH: %w", err)
	}
	return RunCmd(dir, "go", "work", "sync")
}

// RunGoModTidy 在 dir 下执行 go mod tidy。
func RunGoModTidy(dir string) error {
	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("go command not found in PATH: %w", err)
	}
	return RunCmd(dir, "go", "mod", "tidy")
}

// RunGitInit 在 dir 下执行 git init。
func RunGitInit(dir string) error {
	if _, err := exec.LookPath("git"); err != nil {
		return fmt.Errorf("git command not found in PATH: %w", err)
	}
	return RunCmd(dir, "git", "init")
}
