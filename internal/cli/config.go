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
		fmt.Printf("  Root:   %s\n", cfg.Root)
		fmt.Printf("  Binary: %s\n", cfg.Binary)
		if cfg.ConfigPath != "" {
			fmt.Printf("  Config: %s\n", cfg.ConfigPath)
		} else {
			fmt.Printf("  Config: none (using defaults)\n")
		}
		fmt.Println("\nTest configuration:")
		fmt.Printf("  Engine: %s\n", cfg.Test.Engine)
		if cfg.Test.Args != "" {
			fmt.Printf("  Args:   %s\n", cfg.Test.Args)
		} else {
			fmt.Printf("  Args:   (none)\n")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
}
