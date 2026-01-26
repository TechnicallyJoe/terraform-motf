package cmd

import (
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Long:  "Print the version, commit, and build date of motf.",
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Print(versionTemplate())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
