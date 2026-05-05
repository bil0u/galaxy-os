package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	bot         string
	version     string
	commit      string
	development = "false"
)

func init() {
	root.AddCommand(botCmd)
}

var root = &cobra.Command{
	Use:   "bot",
	Short: "Launches the bot",
	Long:  `App is a CLI tool to launch the battleground app.`,
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Println("App launched")
	},
}

func Execute() error {
	return root.Execute()
}
