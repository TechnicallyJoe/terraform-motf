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
		if cfg.ConfigPath != "" {
			fmt.Printf("Config file: %s\n\n", cfg.ConfigPath)
		} else {
			fmt.Print("Config file: none (using defaults)\n")
		}

		fmt.Println("\nSettings:")
		fmt.Printf("  root:   %s\n", valueOrDefault(cfg.Root, "(current directory)"))
		fmt.Printf("  binary: %s\n", cfg.Binary)

		fmt.Println("\nTest:")
		fmt.Printf("  engine: %s\n", cfg.Test.Engine)
		fmt.Printf("  args:   %s\n", valueOrDefault(cfg.Test.Args, "(none)"))

		fmt.Println("\nParallelism:")
		fmt.Printf("  max_jobs: %d\n", cfg.Parallelism.GetMaxJobs())

		if len(cfg.Tasks) > 0 {
			fmt.Println("\nTasks:")

			// Determine the maximum task name length to align descriptions nicely.
			nameWidth := 15
			for name := range cfg.Tasks {
				if l := len(name); l > nameWidth {
					nameWidth = l
				}
			}

			for name, task := range cfg.Tasks {
				fmt.Printf(" - %-*s %s\n", nameWidth, name, valueOrDefault(task.Description, "(no description)"))
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
