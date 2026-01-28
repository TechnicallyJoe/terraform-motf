package cli

import (
	"github.com/spf13/cobra"
)

// valCmd represents the validate command
var valCmd = &cobra.Command{
	Use:     "val [module-name]",
	Aliases: []string{"validate"},
	Short:   "Run terraform/tofu validate on a component, base, or project",
	Long: `Run terraform/tofu validate on a component, base, or project.

Use the --example/-e flag to run validate on a specific example instead of the module itself.

Examples:
  motf val storage-account              # Run validate on storage-account module
  motf val storage-account -e basic     # Run validate on the 'basic' example
  motf val -i storage-account -e basic  # Run init then validate on the 'basic' example`,
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
				return runner.RunValidate(moduleAbsPath, argsFlag...)
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

		return runner.RunValidate(targetPath, argsFlag...)
	},
}

func init() {
	valCmd.Flags().BoolVarP(&initFlag, "init", "i", false, "Run init before the command")
	valCmd.Flags().StringVarP(&exampleFlag, "example", "e", "", "Run on a specific example instead of the module")
	valCmd.Flags().BoolVar(&changedFlag, "changed", false, "Run on modules changed compared to --ref")
	valCmd.Flags().StringVar(&refFlag, "ref", "", "Git ref for --changed (default: auto-detect from origin/HEAD)")
	rootCmd.AddCommand(valCmd)
}
