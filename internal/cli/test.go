package cli

import (
	"github.com/spf13/cobra"
)

// testCmd represents the test command
var testCmd = &cobra.Command{
	Use:   "test [module-name]",
	Short: "Run tests on a component, base, or project",
	Long: `Run tests on a component, base, or project using the configured test engine.

The test engine (e.g., terratest, terraform, tofu) is configured in .motf.yml under the 'test' section.
By default, terratest is used, which runs 'go test ./...' in the module directory.

Examples:
  motf test storage-account                    # Run tests on storage-account module
  motf test storage-account -a -v              # Run tests with verbose output
  motf test storage-account -a -timeout=30m    # Run tests with custom timeout`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if changedFlag {
			if len(args) > 0 {
				return cobra.MaximumNArgs(0)(cmd, args)
			}
			return runOnChangedModules(func(moduleAbsPath string) error {
				return runner.RunTest(moduleAbsPath, argsFlag...)
			})
		}

		targetPath, err := resolveTargetPath(args)
		if err != nil {
			return err
		}

		return runner.RunTest(targetPath, argsFlag...)
	},
}

func init() {
	testCmd.Flags().BoolVar(&changedFlag, "changed", false, "Run on modules changed compared to --ref")
	testCmd.Flags().StringVar(&refFlag, "ref", "", "Git ref for --changed (default: auto-detect from origin/HEAD)")
	rootCmd.AddCommand(testCmd)
}
