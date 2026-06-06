package cli

import (
	"fmt"
	"io"
	"sort"

	"github.com/TechnicallyJoe/terraform-motf/internal/git"
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

Tasks support a 'scope' field to control where they run:
  scope: module  (default) Run per-module as normal
  scope: root    Run once from the configured root path
  scope: git     Run once from the git repository root

Examples:
  motf task storage-account                    # List available tasks
  motf task storage-account --list             # List available tasks
  motf task storage-account -t hello-world     # Run 'hello-world' task
  motf task storage-account --task lint        # Run 'lint' task
  motf task storage-account -t lint -e basic   # Run 'lint' task on 'basic' example
  motf task --path ./modules/x -t docs         # Run task on explicit path
  motf task -t lint --changed                  # Run 'lint' task on changed modules
  motf task -t lint --changed --parallel       # Run 'lint' task on changed modules in parallel
  motf task -t fmt                             # Run scoped task (scope: root or git)`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// If no task specified, list tasks
		if taskFlag == "" || listTaskFlag {
			return listTasks()
		}

		// Get git root (soft fail - empty string if not in git repo)
		gitRoot, _ := git.GetRepoRoot()

		// Non-module scoped tasks run once from a specific directory
		if taskCfg := cfg.Tasks[taskFlag]; taskCfg != nil {
			if err := taskCfg.ValidateScope(); err != nil {
				return fmt.Errorf("task %q: %w", taskFlag, err)
			}

			scope := taskCfg.EffectiveScope()
			if scope != tasks.ScopeModule {
				if exampleFlag != "" || len(args) > 0 {
					return fmt.Errorf("cannot use --example or module name with %s-scoped task %q", scope, taskFlag)
				}
				if pathFlag != "" {
					return fmt.Errorf("cannot use --path with %s-scoped task %q", scope, taskFlag)
				}

				var workDir string
				switch scope {
				case tasks.ScopeGit:
					if gitRoot == "" {
						return fmt.Errorf("task %q has scope %q but not in a git repository", taskFlag, scope)
					}
					workDir = gitRoot
				case tasks.ScopeRoot:
					basePath, err := getBasePath()
					if err != nil {
						return err
					}
					workDir = basePath
				}

				env := tasks.NewEnvBuilder().
					WithGitRoot(gitRoot).
					WithModulePath("").
					WithModuleName("").
					WithConfigPath(cfg.ConfigPath).
					WithBinary(cfg.Binary).
					Build()
				taskRunner := tasks.NewRunner(cfg.Tasks, env)
				return taskRunner.Run(taskFlag, workDir)
			}
		}

		if changedFlag {
			if exampleFlag != "" {
				return fmt.Errorf("--changed cannot be used with --example")
			}
			if len(args) > 0 {
				return cobra.MaximumNArgs(0)(cmd, args)
			}
			return runOnChangedModulesWithPath(func(moduleAbsPath string, stdout, stderr io.Writer) error {
				taskRunner := tasks.NewRunner(cfg.Tasks, buildTaskEnv(gitRoot, moduleAbsPath))
				return taskRunner.RunWithOutput(taskFlag, moduleAbsPath, stdout, stderr)
			})
		}

		// Resolve module path (with optional example)
		targetPath, err := resolveTargetWithExample(args, exampleFlag)
		if err != nil {
			return err
		}

		// Run the task
		taskRunner := tasks.NewRunner(cfg.Tasks, buildTaskEnv(gitRoot, targetPath))
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
		if task == nil {
			return fmt.Errorf("task %q has an empty definition in .motf.yml", name)
		}
		if err := task.ValidateScope(); err != nil {
			return fmt.Errorf("task %q: %w", name, err)
		}
		label := name
		if scope := task.EffectiveScope(); scope != tasks.ScopeModule {
			label = name + " [" + scope + "]"
		}
		if task.Description != "" {
			fmt.Printf("  %-20s %s\n", label, task.Description)
		} else {
			fmt.Printf("  %s\n", label)
		}
	}
	return nil
}

// buildTaskEnv creates the environment variables for task execution.
func buildTaskEnv(gitRoot, modulePath string) []string {
	return tasks.NewEnvBuilder().
		WithGitRoot(gitRoot).
		WithModulePath(modulePath).
		WithModuleName(tasks.ModuleNameFromPath(modulePath)).
		WithConfigPath(cfg.ConfigPath).
		WithBinary(cfg.Binary).
		Build()
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
