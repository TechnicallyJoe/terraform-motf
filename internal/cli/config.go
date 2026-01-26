package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show current configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("Current configuration:")
		fmt.Printf("  root:   %s\n", cfg.Root)
		fmt.Printf("  binary: %s\n", cfg.Binary)
		if cfg.ConfigPath != "" {
			fmt.Printf("  config: %s\n", cfg.ConfigPath)
		} else {
			fmt.Printf("  config: none (using defaults)\n")
		}
		fmt.Println("\nTest configuration:")
		fmt.Printf("  engine: %s\n", cfg.Test.Engine)
		if cfg.Test.Args != "" {
			fmt.Printf("  args:   %s\n", cfg.Test.Args)
		} else {
			fmt.Printf("  args:   (none)\n")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
