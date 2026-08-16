package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// capFirstS 首字母大写（局部工具，避免与其他包冲突）。
func capFirstS(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// AppendAppConnector 在 pkg 中为 appName 追加连接件（与 oldApp 并存，不覆盖）。
// appName 的多种大小写形态由 oldApp 形态经 ReplaceTokensInTree 统一替换得出：
//   - demo → user（小写）
//   - Demo → User（大写）
//   - demoapp → userapp（复合 token）
//
// 必须覆盖新 app 会引用到的全部 dbclient 连接件：gorm(DB)、es(ES)、以及 testsetup(常量+初始器)。
// 原因：新 app 的 svchealth 同时引用 dbclient.<X>DB 和 dbclient.<X>ES，两者缺一即编译失败。
func AppendAppConnector(pkgDir, oldApp, appName string) error {
	cap := capFirstS(appName)
	newCap := capFirstS(appName + "app") // Userapp，对应 Demoapp

	// 1) constant.go 追加 AppNameUser = "user"
	constantPath := filepath.Join(pkgDir, "testsetup", "constant.go")
	if err := appendConstantEntry(constantPath, cap, appName); err != nil {
		return err
	}
	// 2) dbclient/gorm.go 追加 dbNameUser 常量与 UserDB(ctx)
	gormPath := filepath.Join(pkgDir, "dbclient", "gorm.go")
	if err := appendDBClientEntry(gormPath, cap, appName); err != nil {
		return err
	}
	// 2b) dbclient/es.go 追加 ESServiceUser 常量、var UserES、switch case；否则新 app svchealth 引用 UserES 编译失败
	esPath := filepath.Join(pkgDir, "dbclient", "es.go")
	if err := appendESClientEntry(esPath, cap, appName); err != nil {
		return err
	}
	// 3) 从 initializer_demo.go 复制生成 initializer_user.go，并把 demo 的 token 全部替换
	src := filepath.Join(pkgDir, "testsetup", "initializer_"+oldApp+".go")
	dst := filepath.Join(pkgDir, "testsetup", "initializer_"+appName+".go")
	if err := copyInitializer(src, dst, oldApp, appName, newCap); err != nil {
		return err
	}
	// 4) init.go 注册 AppNameUser -> newUserappInitializer
	initPath := filepath.Join(pkgDir, "testsetup", "init.go")
	if err := appendInitEntry(initPath, cap, newCap); err != nil {
		return err
	}
	return nil
}

func appendConstantEntry(path, cap, appName string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	marker := "AppName" + cap
	if strings.Contains(string(content), marker) {
		return nil
	}
	// 在 const 块内 AppNameDemo 行后追加 AppNameUser = "user"
	demoAnchor := "AppName" + capFirstS("demo")
	if idx := strings.Index(string(content), demoAnchor); idx >= 0 {
		nl := strings.Index(string(content[idx:]), "\n")
		insertAt := idx + nl + 1
		line := "\t" + marker + " = " + strconv.Quote(appName) + "\n"
		return os.WriteFile(path, []byte(string(content[:insertAt])+line+string(content[insertAt:])), 0o644)
	}
	line := "\t" + marker + " = " + strconv.Quote(appName) + "\n"
	return os.WriteFile(path, append(content, []byte(line)...), 0o644)
}

func appendDBClientEntry(path, cap, appName string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	dbMarker := "dbName" + cap
	if strings.Contains(string(content), dbMarker) {
		return nil
	}
	// 常量：在 dbNameDemo 行后插入 dbNameUser = "user"
	dbDemo := "dbName" + capFirstS("demo")
	if idx := strings.Index(string(content), dbDemo); idx >= 0 {
		nl := strings.Index(string(content[idx:]), "\n")
		insertAt := idx + nl + 1
		content = []byte(string(content[:insertAt]) + "\tdbName"+cap+" = \""+appName+"\"\n" + string(content[insertAt:]))
	} else {
		return fmt.Errorf("appendDBClientEntry: cannot find %s in %s", dbDemo, path)
	}
	// 追加访问器 UserDB(ctx)
	block := "\nfunc " + cap + "DB(ctx context.Context) *gorm.DB {\n" +
		"\treturn GetDB(ctx, dbName" + cap + ")\n}\n"
	return os.WriteFile(path, append(content, []byte(block)...), 0o644)
}

// appendESClientEntry 在 es.go 中追加 appName 的 ES 连接件：
//   - `var <Cap>ES *elasticsearch.Client` 追加到 var 块（DemoES 行后）
//   - `ESService<Cap> = "<appName>"` 追加到 const 块（ESServiceDemo 行后）
//   - switch 中加 `case ESService<Cap>: <Cap>ES = client`
//
// 等三处，保证新 app 的 svchealth 引用的 dbclient.<Cap>ES 存在且被初始化。
func appendESClientEntry(path, cap, appName string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	text := string(content)
	esMarker := cap + "ES"
	if strings.Contains(text, esMarker+" *elasticsearch.Client") {
		return nil
	}
	capService := "ESService" + cap
	lineSEnd := "\n\t" + esMarker + " *elasticsearch.Client"
	// var 块：DemoES 声明后追加
	oldVar := "DemoES *elasticsearch.Client"
	if idx := strings.Index(text, oldVar); idx >= 0 {
		nl := strings.Index(text[idx:], "\n")
		insertAt := idx + nl + 1
		text = text[:insertAt] + lineSEnd + "\n" + text[insertAt:]
	} else {
		return fmt.Errorf("appendESClientEntry: cannot find var %q in %s", oldVar, path)
	}
	// const 块：ESServiceDemo 后追加 ESServiceUser = "user"
	oldConst := "ESServiceDemo = \"demo\""
	if idx := strings.Index(text, oldConst); idx >= 0 {
		nl := strings.Index(text[idx:], "\n")
		insertAt := idx + nl + 1
		text = text[:insertAt] + "\t" + capService + " = \"" + appName + "\"\n" + text[insertAt:]
	} else {
		return fmt.Errorf("appendESClientEntry: cannot find const %q in %s", oldConst, path)
	}
	// switch：case ESServiceDemo 后追加 case ESServiceUser: UserES = client
	oldCase := "case ESServiceDemo:"
	if idx := strings.Index(text, oldCase); idx >= 0 {
		// 用一个哨兵字符串实现等价多行插入（保持缩进一致）
		replacement := "case ESServiceDemo:\n" +
			"\t\t" + "case " + capService + ":" + "\n" +
			"\t\t" + cap + "ES = client"
		text = strings.Replace(text, oldCase+"\n\t\t\tDemoES = client", replacement+"\n\t\t\tDemoES = client", 1)
	} else {
		return fmt.Errorf("appendESClientEntry: cannot find switch case %q in %s", oldCase, path)
	}
	return os.WriteFile(path, []byte(text), 0o644)
}

// copyInitializer 复制 initializer_<oldApp>.go 为 initializer_<appName>.go，
// 并把其中的 <oldApp> 相关 token 替换为 appName 形态（demoapp→userapp、Demoapp→Userapp、AppNameDemo→AppNameUser）。
func copyInitializer(src, dst, oldApp, appName, newCap string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	text := string(content)
	oldAppToken := oldApp + "app"    // demoapp
	oldCap := capFirstS(oldAppToken) // Demoapp
	text = strings.ReplaceAll(text, oldAppToken, appName+"app")
	text = strings.ReplaceAll(text, oldCap, newCap)
	text = strings.ReplaceAll(text, "AppName"+capFirstS(oldApp), "AppName"+capFirstS(appName))
	return os.WriteFile(dst, []byte(text), 0o644)
}

func appendInitEntry(path, cap, newCap string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	key := "AppName" + cap
	if strings.Contains(string(content), key+":") {
		return nil
	}
	// demo 的注册行：AppNameDemo: newDemoappInitializer（注意值首字母大写 Demoapp）
	demoLine := "AppName" + capFirstS("demo") + ": new" + capFirstS("demo") + "appInitializer"
	entry := "\t\t" + key + ": new" + newCap + "Initializer,\n"
	if idx := strings.Index(string(content), demoLine); idx >= 0 {
		// 找到该行结尾换行并插入
		nl := strings.Index(string(content[idx:]), "\n")
		insertAt := idx + nl + 1
		content = []byte(string(content[:insertAt]) + entry + string(content[insertAt:]))
		return os.WriteFile(path, content, 0o644)
	}
	return fmt.Errorf("appendInitEntry: cannot find anchor %q in %s", demoLine, path)
}
