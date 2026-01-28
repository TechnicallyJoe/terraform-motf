package cli

import (
	"github.com/spf13/cobra"
)

// planCmd represents the plan command
var planCmd = &cobra.Command{
	Use:   "plan [module-name]",
	Short: "Run terraform/tofu plan on a component, base, or project",
	Long: `Run terraform/tofu plan on a component, base, or project.

Use the --example/-e flag to run plan on a specific example instead of the module itself.

Examples:
  motf plan storage-account              # Run plan on storage-account module
  motf plan storage-account -e basic     # Run plan on the 'basic' example
  motf plan storage-account --example basic
  motf plan -i storage-account           # Run init then plan`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if changedFlag {
			if len(args) > 0 {
				return cobra.MaximumNArgs(0)(cmd, args)
			}
			return runOnChangedModules(func(moduleAbsPath string) error {
				if initFlag {
					if err := runner.RunInit(moduleAbsPath); err != nil {
						return err
					}
				}
				return runner.RunPlan(moduleAbsPath, argsFlag...)
			})
		}

		targetPath, err := resolveTargetWithExample(args, exampleFlag)
		if err != nil {
			return err
		}

		// Run init first if flag is set
		if initFlag {
			if err := runner.RunInit(targetPath); err != nil {
				return err
			}
		}

		return runner.RunPlan(targetPath, argsFlag...)
	},
}

func init() {
	planCmd.Flags().BoolVarP(&initFlag, "init", "i", false, "Run init before the command")
	planCmd.Flags().StringVarP(&exampleFlag, "example", "e", "", "Run on a specific example instead of the module")
	planCmd.Flags().BoolVar(&changedFlag, "changed", false, "Run on modules changed compared to --ref")
	planCmd.Flags().StringVar(&refFlag, "ref", "", "Git ref for --changed (default: auto-detect from origin/HEAD)")
	rootCmd.AddCommand(planCmd)
}
