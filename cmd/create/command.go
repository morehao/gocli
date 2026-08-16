// Package create 提供基于内置模板快速创建 Go 项目/应用的能力：
//
//	gocli create project [dir] -m <module-path>   # 创建后端 monorepo 项目
//	gocli create app -n <app-name>                # 在既有 monorepo 中新增 app
package create

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Cmd represents the create command for scaffolding projects and apps from built-in templates.
var Cmd = &cobra.Command{
	Use:   "create",
	Short: "Scaffold a backend monorepo project or add an app to an existing monorepo",
	Long: `Create commands scaffold projects and apps from the built-in template:

  gocli create project my-backend -m github.com/acme/backend
  gocli create app -n userapp`,
}

var projectCmd = &cobra.Command{
	Use:   "project [dir]",
	Short: "Create a new backend monorepo project from the built-in template",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		modulePath, err := cmd.Flags().GetString("module")
		if err != nil {
			fmt.Println("Error retrieving module flag:", err)
			os.Exit(1)
		}
		gitInit, err := cmd.Flags().GetBool("git")
		if err != nil {
			fmt.Println("Error retrieving git flag:", err)
			os.Exit(1)
		}
		skipTidy, err := cmd.Flags().GetBool("no-tidy")
		if err != nil {
			fmt.Println("Error retrieving no-tidy flag:", err)
			os.Exit(1)
		}

		projName, err := cmd.Flags().GetString("name")
		if err != nil {
			fmt.Println("Error retrieving name flag:", err)
			os.Exit(1)
		}

		dir := ""
		if len(args) == 1 {
			dir = args[0]
		}
		if err := createProjectWithOpts(CreateOptions{
			Dir:         dir,
			ModulePath:  modulePath,
			ProjectName: projName,
			GitInit:     gitInit,
			Tidy:        !skipTidy,
		}); err != nil {
			fmt.Println("Error creating project:", err)
			os.Exit(1)
		}
	},
}

var appCmd = &cobra.Command{
	Use:   "app",
	Short: "Add a new app to the current monorepo from the ark demo app",
	Run: func(cmd *cobra.Command, args []string) {
		appName, err := cmd.Flags().GetString("name")
		if err != nil {
			fmt.Println("Error retrieving name flag:", err)
			os.Exit(1)
		}
		skipTidy, err := cmd.Flags().GetBool("no-tidy")
		if err != nil {
			fmt.Println("Error retrieving no-tidy flag:", err)
			os.Exit(1)
		}

		if err := createAppX(appName, !skipTidy); err != nil {
			fmt.Println("Error creating app:", err)
			os.Exit(1)
		}
	},
}

func init() {
	// project flags
	projectCmd.Flags().StringP("module", "m", "", "Root module path, e.g. github.com/acme/backend (required)")
	projectCmd.Flags().StringP("name", "p", "", "Project name (defaults to dir basename or module last segment)")
	projectCmd.Flags().Bool("git", false, "Initialize a git repository after creation")
	projectCmd.Flags().Bool("no-tidy", false, "Skip `go work sync` after creation (runs by default)")

	// app flags
	appCmd.Flags().StringP("name", "n", "", "New app name, e.g. userapp (required)")
	appCmd.Flags().Bool("no-tidy", false, "Skip `go mod tidy` for the new app (runs by default)")

	Cmd.AddCommand(projectCmd, appCmd)
}
