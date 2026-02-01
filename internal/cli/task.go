package cli

import (
	"fmt"
	"io"
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
  motf task storage-account -t lint -e basic   # Run 'lint' task on 'basic' example
  motf task --path ./modules/x -t docs         # Run task on explicit path
  motf task -t lint --changed                  # Run 'lint' task on changed modules
  motf task -t lint --changed --parallel       # Run 'lint' task on changed modules in parallel`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// If no task specified, list tasks
		if taskFlag == "" || listTaskFlag {
			return listTasks()
		}

		if changedFlag {
			if exampleFlag != "" {
				return fmt.Errorf("--changed cannot be used with --example")
			}
			if len(args) > 0 {
				return cobra.MaximumNArgs(0)(cmd, args)
			}
			return runOnChangedModulesWithPath(func(moduleAbsPath string, stdout, stderr io.Writer) error {
				taskRunner := tasks.NewRunner(cfg.Tasks)
				return taskRunner.RunWithOutput(taskFlag, moduleAbsPath, stdout, stderr)
			})
		}

		// Resolve module path (with optional example)
		targetPath, err := resolveTargetWithExample(args, exampleFlag)
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
	taskCmd.Flags().StringVarP(&exampleFlag, "example", "e", "", "Run on a specific example instead of the module")
	taskCmd.Flags().BoolVar(&changedFlag, "changed", false, "Run on modules changed compared to --ref")
	taskCmd.Flags().StringVar(&refFlag, "ref", "", "Git ref for --changed (default: auto-detect from origin/HEAD)")
	taskCmd.Flags().BoolVarP(&parallelFlag, "parallel", "p", false, "Run commands in parallel")
	taskCmd.Flags().IntVar(&maxParallelFlag, "max-parallel", 0, "Maximum parallel jobs (default: number of CPU cores)")
	rootCmd.AddCommand(taskCmd)
}
