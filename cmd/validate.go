package cmd

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
	rootCmd.AddCommand(valCmd)
}
