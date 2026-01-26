package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/TechnicallyJoe/terraform-motf/internal/finder"
	"github.com/TechnicallyJoe/terraform-motf/internal/spacelift"
	"github.com/spf13/cobra"
)

// showJsonFlag controls JSON output for show command
var showJsonFlag bool

// showCmd represents the show command
var showCmd = &cobra.Command{
	Use:   "show [module-name]",
	Short: "Show details about a component, base, or project",
	Long: `Show detailed information about a module including its type, path,
whether it has submodules, tests, examples, and its Spacelift registry version.

Use the --json flag to output in JSON format for scripting.

Examples:
  motf show storage-account      # Show details for storage-account
  motf show --path ./my-module   # Show details for module at explicit path
  motf show storage-account --json  # Output as JSON`,
	Args: cobra.MaximumNArgs(1),
	RunE: runShow,
}

func init() {
	showCmd.Flags().BoolVar(&showJsonFlag, "json", false, "Output in JSON format")
	rootCmd.AddCommand(showCmd)
}

// ModuleDetails contains detailed information about a module
type ModuleDetails struct {
	Name             string     `json:"name"`
	Type             string     `json:"type"`
	Path             string     `json:"path"`
	HasSubmodules    bool       `json:"has_submodules"`
	HasTests         bool       `json:"has_tests"`
	HasExamples      bool       `json:"has_examples"`
	Submodules       []ItemInfo `json:"submodules,omitempty"`
	Examples         []ItemInfo `json:"examples,omitempty"`
	Tests            []ItemInfo `json:"tests,omitempty"`
	SpaceliftVersion string     `json:"spacelift_version,omitempty"`
}

// ItemInfo contains information about an example
type ItemInfo struct {
	Name string `json:"name"`
	Path string `json:"path"`
}

func runShow(cmd *cobra.Command, args []string) error {
	targetPath, err := resolveTargetPath(args)
	if err != nil {
		return err
	}

	details, err := getModuleDetails(targetPath)
	if err != nil {
		return err
	}

	if showJsonFlag {
		return printModuleDetailsJSON(details)
	}

	printModuleDetails(details)
	return nil
}

// getModuleDetails gathers all information about a module
func getModuleDetails(modulePath string) (*ModuleDetails, error) {
	// Get module name from path
	name := filepath.Base(modulePath)

	// Get module type using existing helper
	modType := getModuleType(modulePath)

	// Get relative path from base
	basePath, err := getBasePath()
	if err != nil {
		return nil, err
	}
	relativePath, err := filepath.Rel(basePath, modulePath)
	if err != nil {
		relativePath = modulePath
	}

	// Check for submodules directory
	hasSubmodules := dirHasContent(filepath.Join(modulePath, DirModules))

	// Check for tests directory
	hasTests := dirHasContent(filepath.Join(modulePath, DirTests))

	// Check for examples directory
	hasExamples := dirHasContent(filepath.Join(modulePath, DirExamples))

	// Get list of submodules
	submodules := listItems(filepath.Join(modulePath, DirModules), basePath)

	// Get list of examples
	examples := listItems(filepath.Join(modulePath, DirExamples), basePath)

	// Get list of test files
	tests := listTestFiles(filepath.Join(modulePath, DirTests), basePath)

	// Get Spacelift version
	spaceliftVersion := spacelift.ReadModuleVersion(modulePath)

	return &ModuleDetails{
		Name:             name,
		Type:             modType,
		Path:             relativePath,
		HasSubmodules:    hasSubmodules,
		HasTests:         hasTests,
		HasExamples:      hasExamples,
		Submodules:       submodules,
		Examples:         examples,
		Tests:            tests,
		SpaceliftVersion: spaceliftVersion,
	}, nil
}

// dirHasContent checks if a directory exists and has at least one entry
func dirHasContent(path string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	return len(entries) > 0
}

// listItems returns a list of items (submodules/examples) in the directory
func listItems(path, basePath string) []ItemInfo {
	var items []ItemInfo

	entries, err := os.ReadDir(path)
	if err != nil {
		return items
	}

	for _, entry := range entries {
		// Only include directories that contain .tf files
		if entry.IsDir() {
			dirPath := filepath.Join(path, entry.Name())
			if finder.HasTerraformFiles(dirPath) {
				relativePath, err := filepath.Rel(basePath, dirPath)
				if err != nil {
					relativePath = dirPath
				}
				items = append(items, ItemInfo{
					Name: entry.Name(),
					Path: relativePath,
				})
			}
		}
	}

	return items
}

// listTestFiles returns a list of test files (*_test.go) in the directory
func listTestFiles(path, basePath string) []ItemInfo {
	var items []ItemInfo

	entries, err := os.ReadDir(path)
	if err != nil {
		return items
	}

	for _, entry := range entries {
		// Only include files matching *_test.go pattern
		if !entry.IsDir() {
			name := entry.Name()
			if strings.HasSuffix(name, "_test.go") {
				filePath := filepath.Join(path, name)
				relativePath, err := filepath.Rel(basePath, filePath)
				if err != nil {
					relativePath = filePath
				}
				items = append(items, ItemInfo{
					Name: name,
					Path: relativePath,
				})
			}
		}
	}

	return items
}

// printModuleDetails outputs the module details in a formatted way
func printModuleDetails(details *ModuleDetails) {
	fmt.Printf("Name:                  %s\n", details.Name)
	fmt.Printf("Type:                  %s\n", formatType(details.Type))
	fmt.Printf("Path:                  %s\n", details.Path)
	fmt.Printf("Spacelift Version:     %s\n", details.SpaceliftVersion)
	fmt.Printf("Has Submodules:        %s\n", formatBool(details.HasSubmodules))
	fmt.Printf("Has Tests:             %s\n", formatBool(details.HasTests))
	fmt.Printf("Has Examples:          %s\n", formatBool(details.HasExamples))

	if len(details.Submodules) > 0 {
		fmt.Println("\nSubmodules:")
		for _, ex := range details.Submodules {
			fmt.Printf("  - %s (%s)\n", ex.Name, ex.Path)
		}
	}

	if len(details.Examples) > 0 {
		fmt.Println("\nExamples:")
		for _, ex := range details.Examples {
			fmt.Printf("  - %s (%s)\n", ex.Name, ex.Path)
		}
	}

	if len(details.Tests) > 0 {
		fmt.Println("\nTests:")
		for _, ex := range details.Tests {
			fmt.Printf("  - %s (%s)\n", ex.Name, ex.Path)
		}
	}

}

// formatBool returns "Yes" or "No" for boolean values
func formatBool(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}

// formatType returns a formatted type string or "unknown" if empty
func formatType(t string) string {
	if t == "" {
		return "unknown"
	}
	return t
}

// printModuleDetailsJSON outputs the module details in JSON format
func printModuleDetailsJSON(details *ModuleDetails) error {
	output, err := json.MarshalIndent(details, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(output))
	return nil
}
