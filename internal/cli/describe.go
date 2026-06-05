package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/TechnicallyJoe/terraform-motf/internal/terraform"
	"github.com/spf13/cobra"
)

var describeJsonFlag bool

var describeCmd = &cobra.Command{
	Use:   "describe [module-name]",
	Short: "Describe the interface of a Terraform module",
	Long: `Parse and display the inputs, outputs, and providers of a Terraform module.

Shows the module's required Terraform version, provider dependencies,
input variables (with types, defaults, and descriptions), and outputs.`,
	Example: `  motf describe storage-account       # Describe storage-account module
  motf describe k8s-argocd --json     # Output as JSON
  motf describe --path ./my-module    # Describe module at explicit path`,
	Args: cobra.MaximumNArgs(1),
	RunE: runDescribe,
}

func init() {
	describeCmd.Flags().BoolVar(&describeJsonFlag, "json", false, "Output in JSON format")
	describeCmd.Flags().BoolVar(&changedFlag, "changed", false, "Run on modules changed compared to --ref")
	describeCmd.Flags().StringVar(&refFlag, "ref", "", "Git ref for --changed (default: auto-detect from origin/HEAD)")
	describeCmd.Flags().BoolVarP(&parallelFlag, "parallel", "p", false, "Run commands in parallel")
	describeCmd.Flags().IntVar(&maxParallelFlag, "max-parallel", 0, "Maximum parallel jobs (default: number of CPU cores)")
	rootCmd.AddCommand(describeCmd)
}

func runDescribe(cmd *cobra.Command, args []string) error {
	if changedFlag {
		if len(args) > 0 {
			return cobra.MaximumNArgs(0)(cmd, args)
		}
		if describeJsonFlag {
			return describeChangedJSON()
		}
		return runOnChangedModulesWithPath(func(moduleAbsPath string, stdout, stderr io.Writer) error {
			schema, err := terraform.LoadModuleSchema(moduleAbsPath, getRoot())
			if err != nil {
				_, _ = fmt.Fprintf(stderr, "failed to parse module: %v\n", err)
				return err
			}
			printSchemaToWriter(stdout, schema)
			return nil
		})
	}

	targetPath, err := resolveTargetPath(args)
	if err != nil {
		return err
	}

	schema, err := terraform.LoadModuleSchema(targetPath, getRoot())
	if err != nil {
		return fmt.Errorf("failed to parse module: %w", err)
	}

	if describeJsonFlag {
		return printSchemaJSON(cmd, schema)
	}

	printSchema(cmd, schema)
	return nil
}

func describeChangedJSON() error {
	modules, err := detectChangedModules(refFlag)
	if err != nil {
		return err
	}

	basePath, err := getBasePath()
	if err != nil {
		return err
	}

	schemas := make([]*terraform.ModuleSchema, 0)
	for _, mod := range modules {
		absPath := filepath.Join(basePath, mod.Path)
		schema, err := terraform.LoadModuleSchema(absPath, getRoot())
		if err != nil {
			return fmt.Errorf("failed to parse module %s: %w", mod.Path, err)
		}
		schemas = append(schemas, schema)
	}

	output, err := json.MarshalIndent(schemas, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(output))
	return nil
}

func printSchemaJSON(cmd *cobra.Command, schema *terraform.ModuleSchema) error {
	output, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	cmd.Println(string(output))
	return nil
}

func printSchemaToWriter(w io.Writer, schema *terraform.ModuleSchema) {
	_, _ = fmt.Fprintf(w, "Module: %s\n", schema.Name)
	_, _ = fmt.Fprintf(w, "Path:   %s\n", schema.Path)

	if schema.TerraformVersion != "" {
		_, _ = fmt.Fprintf(w, "\nTerraform: %s\n", schema.TerraformVersion)
	}

	printExampleToWriter(w, schema)

	if len(schema.Providers) > 0 {
		_, _ = fmt.Fprintln(w, "\nProviders:")
		_, _ = fmt.Fprintf(w, "  %-20s %s\n", "NAME", "VERSION")
		for _, p := range schema.Providers {
			version := p.Version
			if version == "" {
				version = "(any)"
			}
			_, _ = fmt.Fprintf(w, "  %-20s %s\n", p.Name, version)
		}
	}

	if len(schema.Variables) > 0 {
		_, _ = fmt.Fprintln(w, "\nVariables:")
		_, _ = fmt.Fprintf(w, "  %-25s %-15s %-15s %s\n", "NAME", "TYPE", "DEFAULT", "DESCRIPTION")
		for _, v := range schema.Variables {
			typeStr := normalizeType(v.Type)
			defaultStr := v.DefaultString()
			descLines := wrapText(v.Description, 60)
			firstDesc := ""
			if len(descLines) > 0 {
				firstDesc = descLines[0]
			}
			_, _ = fmt.Fprintf(w, "  %-25s %-15s %-15s %s\n", truncate(v.Name, 25), truncate(typeStr, 15), truncate(defaultStr, 15), firstDesc)
			for i := 1; i < len(descLines); i++ {
				_, _ = fmt.Fprintf(w, "  %-25s %-15s %-15s %s\n", "", "", "", descLines[i])
			}
		}
	}

	if len(schema.Outputs) > 0 {
		_, _ = fmt.Fprintln(w, "\nOutputs:")
		_, _ = fmt.Fprintf(w, "  %-25s %s\n", "NAME", "DESCRIPTION")
		for _, o := range schema.Outputs {
			desc := o.Description
			if o.Sensitive {
				if desc != "" {
					desc += " (sensitive)"
				} else {
					desc = "(sensitive)"
				}
			}
			descLines := wrapText(desc, 60)
			firstDesc := ""
			if len(descLines) > 0 {
				firstDesc = descLines[0]
			}
			_, _ = fmt.Fprintf(w, "  %-25s %s\n", truncate(o.Name, 25), firstDesc)
			for i := 1; i < len(descLines); i++ {
				_, _ = fmt.Fprintf(w, "  %-25s %s\n", "", descLines[i])
			}
		}
	}
}

func printSchema(cmd *cobra.Command, schema *terraform.ModuleSchema) {
	cmd.Printf("Module: %s\n", schema.Name)
	cmd.Printf("Path:   %s\n", schema.Path)

	if schema.TerraformVersion != "" {
		cmd.Printf("\nTerraform: %s\n", schema.TerraformVersion)
	}

	// Example module block
	printExample(cmd, schema)

	// Providers table
	if len(schema.Providers) > 0 {
		cmd.Println("\nProviders:")
		cmd.Printf("  %-20s %s\n", "NAME", "VERSION")
		for _, p := range schema.Providers {
			version := p.Version
			if version == "" {
				version = "(any)"
			}
			cmd.Printf("  %-20s %s\n", p.Name, version)
		}
	}

	// Variables table
	if len(schema.Variables) > 0 {
		cmd.Println("\nVariables:")
		cmd.Printf("  %-25s %-15s %-15s %s\n", "NAME", "TYPE", "DEFAULT", "DESCRIPTION")
		for _, v := range schema.Variables {
			typeStr := normalizeType(v.Type)
			defaultStr := v.DefaultString()
			descLines := wrapText(v.Description, 60)

			// First line with all columns
			firstDesc := ""
			if len(descLines) > 0 {
				firstDesc = descLines[0]
			}
			cmd.Printf("  %-25s %-15s %-15s %s\n", truncate(v.Name, 25), truncate(typeStr, 15), truncate(defaultStr, 15), firstDesc)

			// Continuation lines for description
			for i := 1; i < len(descLines); i++ {
				cmd.Printf("  %-25s %-15s %-15s %s\n", "", "", "", descLines[i])
			}
		}
	}

	// Outputs table
	if len(schema.Outputs) > 0 {
		cmd.Println("\nOutputs:")
		cmd.Printf("  %-25s %s\n", "NAME", "DESCRIPTION")
		for _, o := range schema.Outputs {
			desc := o.Description
			if o.Sensitive {
				if desc != "" {
					desc += " (sensitive)"
				} else {
					desc = "(sensitive)"
				}
			}
			descLines := wrapText(desc, 60)

			firstDesc := ""
			if len(descLines) > 0 {
				firstDesc = descLines[0]
			}
			cmd.Printf("  %-25s %s\n", truncate(o.Name, 25), firstDesc)

			for i := 1; i < len(descLines); i++ {
				cmd.Printf("  %-25s %s\n", "", descLines[i])
			}
		}
	}
}

func printExample(cmd *cobra.Command, schema *terraform.ModuleSchema) {
	cmd.Println("\nExample:")
	cmd.Printf("  module \"%s\" {\n", schema.Name)
	cmd.Printf("    source = \"%s\"\n", schema.Path)

	// Find the longest required variable name for alignment
	maxLen := 0
	for _, v := range schema.Variables {
		if v.Required && len(v.Name) > maxLen {
			maxLen = len(v.Name)
		}
	}

	// Print required variables (if any)
	if maxLen > 0 {
		cmd.Println()
		for _, v := range schema.Variables {
			if v.Required {
				cmd.Printf("    %-*s = %s\n", maxLen, v.Name, v.EmptyValueForType())
			}
		}
	}

	cmd.Println("  }")
}

func printExampleToWriter(w io.Writer, schema *terraform.ModuleSchema) {
	_, _ = fmt.Fprintln(w, "\nExample:")
	_, _ = fmt.Fprintf(w, "  module \"%s\" {\n", schema.Name)
	_, _ = fmt.Fprintf(w, "    source = \"%s\"\n", schema.Path)

	maxLen := 0
	for _, v := range schema.Variables {
		if v.Required && len(v.Name) > maxLen {
			maxLen = len(v.Name)
		}
	}

	if maxLen > 0 {
		_, _ = fmt.Fprintln(w)
		for _, v := range schema.Variables {
			if v.Required {
				_, _ = fmt.Fprintf(w, "    %-*s = %s\n", maxLen, v.Name, v.EmptyValueForType())
			}
		}
	}

	_, _ = fmt.Fprintln(w, "  }")
}

// normalizeType simplifies complex type definitions for display in tables.
// Complex types like object({...}) and tuple([...]) are simplified to just show the wrapper.
func normalizeType(t string) string {
	t = strings.TrimSpace(t)

	// Simplify object types to object({...})
	if strings.HasPrefix(t, "object(") {
		return "object({...})"
	}

	// Simplify tuple types to tuple([...])
	if strings.HasPrefix(t, "tuple(") {
		return "tuple([...])"
	}

	return t
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

func wrapText(text string, maxWidth int) []string {
	if text == "" {
		return nil
	}

	var lines []string
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}

	currentLine := words[0]
	for _, word := range words[1:] {
		if len(currentLine)+1+len(word) <= maxWidth {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	lines = append(lines, currentLine)

	return lines
}
