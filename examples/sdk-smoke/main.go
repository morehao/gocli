// Command sdk-smoke 演示并验证：外部 Go 程序 import pkg/codegen 直接调用生成能力（不经过 CLI）。
//
// 用法：go run ./examples/sdk-smoke <projectRoot> <dsn> <app> <table> <package> <mode>
// 例：
//
//	go run ./examples/sdk-smoke /path/to/ark-go-starter \
//	  'postgresql://postgres:123456@127.0.0.1:5432/demo?sslmode=disable' \
//	  api demo_tag demotag backend_module
package main

import (
	"fmt"
	"os"

	"github.com/morehao/gocli/pkg/codegen"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	if len(os.Args) < 7 {
		fmt.Println("usage: sdk-smoke <projectRoot> <dsn> <app> <table> <package> <mode>")
		os.Exit(2)
	}
	root, dsn, app, table, pkg, mode := os.Args[1], os.Args[2], os.Args[3], os.Args[4], os.Args[5], os.Args[6]

	dbCfg, err := codegen.ParseDatabaseDSN(dsn)
	if err != nil {
		panic(err)
	}
	var db *gorm.DB
	switch dbCfg.Type {
	case codegen.DBTypePostgres:
		db, err = gorm.Open(postgres.Open(dbCfg.ConnStr), &gorm.Config{})
	case codegen.DBTypeMySQL:
		db, err = gorm.Open(mysql.Open(dbCfg.ConnStr), &gorm.Config{})
	default:
		panic("unsupported db type: " + dbCfg.Type)
	}
	if err != nil {
		panic(err)
	}

	g, err := codegen.New(db, codegen.Options{
		AppName:     app,
		ProjectRoot: root,
		Table:       table,
		PackageName: pkg,
		Mode:        codegen.Mode(mode),
		DryRun:      os.Getenv("DRY_RUN") == "1",
		Web:         codegen.WebConfig{RegisterMenu: true, MenuGroup: "SDK 演示"},
	})
	if err != nil {
		panic(err)
	}
	res, err := g.Generate()
	if err != nil {
		fmt.Println("generate failed:", err)
		os.Exit(1)
	}
	fmt.Println("[sdk-smoke] OK, mode =", mode)
	for _, f := range res.WrittenFiles {
		fmt.Println("  wrote:", f)
	}
	for _, w := range res.Warnings {
		fmt.Println("  warn:", w)
	}
}
