package scaffold

import (
	"fmt"
	"regexp"
	"strings"
)

// appNamePattern 合法的 app 名称：小写字母开头，仅含小写字母与数字。
// 保证替换后 Go 包名、import 路径安全。
var appNamePattern = regexp.MustCompile(`^[a-z][a-z0-9]*$`)

// modulePathPattern 宽松校验模块路径：非空、无空白、不以 / 结尾。
var modulePathPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._~/\-]*$`)

// ValidateAppName 校验 app 名称。
func ValidateAppName(name string) error {
	if name == "" {
		return fmt.Errorf("app name cannot be empty")
	}
	if !appNamePattern.MatchString(name) {
		return fmt.Errorf("invalid app name %q: must match ^[a-z][a-z0-9]*$ (lowercase letters and digits, starting with a letter)", name)
	}
	return nil
}

// ValidateModulePath 校验模块路径。
func ValidateModulePath(modulePath string) error {
	if modulePath == "" {
		return fmt.Errorf("module path cannot be empty")
	}
	if !modulePathPattern.MatchString(modulePath) {
		return fmt.Errorf("invalid module path %q", modulePath)
	}
	if strings.HasSuffix(modulePath, "/") {
		return fmt.Errorf("invalid module path %q: must not end with /", modulePath)
	}
	return nil
}
