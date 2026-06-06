package cli

import (
	"io"

	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:   "apply [module-name]",
	Short: "Run terraform/tofu apply on a component, base, or project",
	Long: `Run terraform/tofu apply on a component, base, or project.

Use the --example/-e flag to run apply on a specific example instead of the module itself.

Examples:
  motf apply storage-account                   # Run apply on storage-account module
  motf apply storage-account -e basic          # Run apply on the 'basic' example
  motf apply -i storage-account                # Run init then apply
  motf apply storage-account -a -auto-approve  # Apply without confirmation
  motf apply --changed -a -auto-approve        # Apply all changed modules without confirmation`,
	Args:    cobra.MaximumNArgs(1),
	PreRunE: requiresBinary,
	RunE: func(cmd *cobra.Command, args []string) error {
		if changedFlag {
			if len(args) > 0 {
				return cobra.MaximumNArgs(0)(cmd, args)
			}
			return runOnChangedModulesWithPath(func(moduleAbsPath string, stdout, stderr io.Writer) error {
				if initFlag {
					if err := runner.RunInitWithOutput(moduleAbsPath, stdout, stderr); err != nil {
						return err
					}
				}
				return runner.RunApplyWithOutput(moduleAbsPath, stdout, stderr, argsFlag...)
			})
		}

		targetPath, err := resolveTargetWithExample(args, exampleFlag)
		if err != nil {
			return err
		}

		if initFlag {
			if err := runner.RunInit(targetPath); err != nil {
				return err
			}
		}

		return runner.RunApply(targetPath, argsFlag...)
	},
}

func init() {
	applyCmd.Flags().BoolVarP(&initFlag, "init", "i", false, "Run init before the command")
	applyCmd.Flags().StringVarP(&exampleFlag, "example", "e", "", "Run on a specific example instead of the module")
	applyCmd.Flags().BoolVar(&changedFlag, "changed", false, "Run on modules changed compared to --ref")
	applyCmd.Flags().StringVar(&refFlag, "ref", "", "Git ref for --changed (default: auto-detect from origin/HEAD)")
	applyCmd.Flags().BoolVarP(&parallelFlag, "parallel", "p", false, "Run commands in parallel")
	applyCmd.Flags().IntVar(&maxParallelFlag, "max-parallel", 0, "Maximum parallel jobs (default: number of CPU cores)")
	rootCmd.AddCommand(applyCmd)
}
