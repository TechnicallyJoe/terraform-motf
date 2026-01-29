package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Show current configuration",
	Long:  "Display the current configuration values, showing which config file is in use (if any) and the effective settings.",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Show config file status first - most important info
		if cfg.ConfigPath != "" {
			fmt.Printf("Config file: %s\n\n", cfg.ConfigPath)
		} else {
			fmt.Print("Config file: none (using defaults)\n\n")
		}

		fmt.Println("Settings:")
		fmt.Printf("  root:   %s\n", valueOrDefault(cfg.Root, "(current directory)"))
		fmt.Printf("  binary: %s\n", cfg.Binary)

		fmt.Println("\nTest:")
		if cfg.Test != nil {
			fmt.Printf("  engine: %s\n", cfg.Test.Engine)
			fmt.Printf("  args:   %s\n", valueOrDefault(cfg.Test.Args, "(none)"))
		} else {
			fmt.Println("  engine: terratest (default)")
			fmt.Println("  args:   (none)")
		}

		if len(cfg.Tasks) > 0 {
			fmt.Println("\nTasks:")
			for name, task := range cfg.Tasks {
				fmt.Printf(" - %-15s %s\n", name, valueOrDefault(task.Description, "(no description)"))
			}
		}

		return nil
	},
}

// valueOrDefault returns the value if non-empty, otherwise the default string
func valueOrDefault(value, defaultStr string) string {
	if value == "" {
		return defaultStr
	}
	return value
}

func init() {
	rootCmd.AddCommand(configCmd)
}
