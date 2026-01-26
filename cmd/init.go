package cmd

import (
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
		targetPath, err := resolveTargetWithExample(args, exampleFlag)
		if err != nil {
			return err
		}

		return runner.RunInit(targetPath, argsFlag...)
	},
}

func init() {
	initCmd.Flags().StringVarP(&exampleFlag, "example", "e", "", "Run on a specific example instead of the module")
	rootCmd.AddCommand(initCmd)
}
