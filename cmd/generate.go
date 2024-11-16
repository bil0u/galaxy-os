package cmd

import (
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate files and configurations",
}

func init() {
	generateCmd.AddCommand(genBotConfigCmd)
	generateCmd.AddCommand(genGuildConfigCmd)
}

var genBotConfigCmd = &cobra.Command{
	Use:   "bot-config",
	Short: "Generate a bot configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		panic("not implemented yet")
	},
}

var genGuildConfigCmd = &cobra.Command{
	Use:   "guild-config",
	Short: "Generate a guild configuration file",
	Run: func(cmd *cobra.Command, args []string) {
		panic("not implemented yet")
	},
}
