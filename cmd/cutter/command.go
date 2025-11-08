package cutter

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// Cmd represents the base command for cloning entire Go projects
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
		if err := cloneProject(destination); err != nil {
			fmt.Println("Error running cutter:", err)
			os.Exit(1)
		}
		fmt.Println("Successfully created Go project at", destination)
	},
}

// AppCmd represents the app subcommand for cloning app within the same project
var AppCmd = &cobra.Command{
	Use:   "app",
	Short: "Clone an app within the same Go project (e.g., clone demoapp to newapp)",
	Run: func(cmd *cobra.Command, args []string) {
		source, err := cmd.Flags().GetString("source")
		if err != nil {
			fmt.Println("Error retrieving source flag:", err)
			os.Exit(1)
		}
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			fmt.Println("Error retrieving name flag:", err)
			os.Exit(1)
		}
		if source == "" {
			fmt.Println("Source app name is empty")
			os.Exit(1)
		}
		if name == "" {
			fmt.Println("New app name is empty")
			os.Exit(1)
		}
		if err := cloneApp(source, name); err != nil {
			fmt.Println("Error cloning app:", err)
			os.Exit(1)
		}
		fmt.Printf("Successfully cloned %s to %s\n", source, name)
	},
}

func init() {
	// Flags for project cloning
	Cmd.Flags().StringP("destination", "d", "", "Destination directory for the new project. For example: ./your/project/path")

	// Flags for app cloning
	AppCmd.Flags().StringP("source", "s", "demoapp", "Source app name to clone from")
	AppCmd.Flags().StringP("name", "n", "", "New app name")

	// Add app subcommand to main command
	Cmd.AddCommand(AppCmd)
}

