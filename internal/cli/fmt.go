package cli

import (
	"io"

	"github.com/spf13/cobra"
)

// fmtCmd represents the fmt command
var fmtCmd = &cobra.Command{
	Use:   "fmt [module-name]",
	Short: "Run terraform/tofu fmt on a component, base, or project",
	Long: `Run terraform/tofu fmt on a component, base, or project.

Use the --example/-e flag to run fmt on a specific example instead of the module itself.

Examples:
  motf fmt storage-account              # Run fmt on storage-account module
  motf fmt storage-account -e basic     # Run fmt on the 'basic' example
  motf fmt -i storage-account -e basic  # Run init then fmt on the 'basic' example`,
	Args: cobra.MaximumNArgs(1),
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
				return runner.RunFmtWithOutput(moduleAbsPath, stdout, stderr, argsFlag...)
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

		return runner.RunFmt(targetPath, argsFlag...)
	},
}

func init() {
	fmtCmd.Flags().BoolVarP(&initFlag, "init", "i", false, "Run init before the command")
	fmtCmd.Flags().StringVarP(&exampleFlag, "example", "e", "", "Run on a specific example instead of the module")
	fmtCmd.Flags().BoolVar(&changedFlag, "changed", false, "Run on modules changed compared to --ref")
	fmtCmd.Flags().StringVar(&refFlag, "ref", "", "Git ref for --changed (default: auto-detect from origin/HEAD)")
	fmtCmd.Flags().BoolVarP(&parallelFlag, "parallel", "p", false, "Run commands in parallel")
	fmtCmd.Flags().IntVar(&maxParallelFlag, "max-parallel", 0, "Maximum parallel jobs (default: number of CPU cores)")
	rootCmd.AddCommand(fmtCmd)
}
