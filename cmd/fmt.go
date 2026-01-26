package cmd

import (
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
	rootCmd.AddCommand(fmtCmd)
}
