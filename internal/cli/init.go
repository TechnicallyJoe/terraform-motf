package cli

import (
	"io"

	"github.com/spf13/cobra"
)

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init [module-name]",
	Short: "Run terraform/tofu init on a component, base, or project",
	Long: `Run terraform/tofu init on a component, base, or project.

Use the --example/-e flag to run init on a specific example instead of the module itself.

Examples:
  motf init storage-account              # Run init on storage-account module
  motf init storage-account -e basic     # Run init on the 'basic' example`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if changedFlag {
			if len(args) > 0 {
				return cobra.MaximumNArgs(0)(cmd, args)
			}
			return runOnChangedModulesWithPath(func(moduleAbsPath string, stdout, stderr io.Writer) error {
				return runner.RunInitWithOutput(moduleAbsPath, stdout, stderr, argsFlag...)
			})
		}

		targetPath, err := resolveTargetWithExample(args, exampleFlag)
		if err != nil {
			return err
		}

		return runner.RunInit(targetPath, argsFlag...)
	},
}

func init() {
	initCmd.Flags().StringVarP(&exampleFlag, "example", "e", "", "Run on a specific example instead of the module")
	initCmd.Flags().BoolVar(&changedFlag, "changed", false, "Run on modules changed compared to --ref")
	initCmd.Flags().StringVar(&refFlag, "ref", "", "Git ref for --changed (default: auto-detect from origin/HEAD)")
	initCmd.Flags().BoolVarP(&parallelFlag, "parallel", "p", false, "Run commands in parallel")
	initCmd.Flags().IntVar(&maxParallelFlag, "max-parallel", 0, "Maximum parallel jobs (default: number of CPU cores)")
	rootCmd.AddCommand(initCmd)
}
