package cmd

import (
	"fmt"
	"sort"

	"github.com/TechnicallyJoe/terraform-motf/internal/tasks"
	"github.com/spf13/cobra"
)

var (
	taskFlag     string
	listTaskFlag bool
)

var taskCmd = &cobra.Command{
	Use:   "task [module-name]",
	Short: "Run a custom task from .motf.yml",
	Long: `Run a custom task defined in .motf.yml on a module.

Tasks are shell commands configured in your .motf.yml file under the 'tasks' section.
By default, or with --list, shows all available tasks.

Examples:
  motf task storage-account                    # List available tasks
  motf task storage-account --list             # List available tasks
  motf task storage-account -t hello-world     # Run 'hello-world' task
  motf task storage-account --task lint        # Run 'lint' task
  motf task --path ./modules/x -t docs         # Run task on explicit path`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// If no task specified, list tasks
		if taskFlag == "" || listTaskFlag {
			return listTasks()
		}

		// Resolve module path
		targetPath, err := resolveTargetPath(args)
		if err != nil {
			return err
		}

		// Run the task
		taskRunner := tasks.NewRunner(cfg.Tasks)
		return taskRunner.Run(taskFlag, targetPath)
	},
}

func listTasks() error {
	if len(cfg.Tasks) == 0 {
		fmt.Println("No tasks defined in .motf.yml")
		return nil
	}

	fmt.Println("Available tasks:")

	// Sort task names for consistent output
	names := make([]string, 0, len(cfg.Tasks))
	for name := range cfg.Tasks {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		task := cfg.Tasks[name]
		if task.Description != "" {
			fmt.Printf("  %-20s %s\n", name, task.Description)
		} else {
			fmt.Printf("  %s\n", name)
		}
	}
	return nil
}

func init() {
	taskCmd.Flags().StringVarP(&taskFlag, "task", "t", "", "Task name to run")
	taskCmd.Flags().BoolVarP(&listTaskFlag, "list", "l", false, "List available tasks")
	rootCmd.AddCommand(taskCmd)
}
