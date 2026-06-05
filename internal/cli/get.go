package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/TechnicallyJoe/terraform-motf/internal/finder"
	"github.com/TechnicallyJoe/terraform-motf/internal/spacelift"
	"github.com/spf13/cobra"
)

// getJsonFlag controls JSON output for get command
var getJsonFlag bool

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get [module-name]",
	Short: "Get details about a component, base, or project",
	Long: `Get detailed information about a module including its type, path,
whether it has submodules, tests, examples, and its Spacelift registry version.

Use the --json flag to output in JSON format for scripting.

Examples:
  motf get storage-account      # Get details for storage-account
  motf get --path ./my-module   # Get details for module at explicit path
  motf get storage-account --json  # Output as JSON`,
	Args: cobra.MaximumNArgs(1),
	RunE: runGet,
}

func init() {
	getCmd.Flags().BoolVar(&getJsonFlag, "json", false, "Output in JSON format")
	getCmd.Flags().BoolVar(&changedFlag, "changed", false, "Run on modules changed compared to --ref")
	getCmd.Flags().StringVar(&refFlag, "ref", "", "Git ref for --changed (default: auto-detect from origin/HEAD)")
	getCmd.Flags().BoolVarP(&parallelFlag, "parallel", "p", false, "Run commands in parallel")
	getCmd.Flags().IntVar(&maxParallelFlag, "max-parallel", 0, "Maximum parallel jobs (default: number of CPU cores)")
	rootCmd.AddCommand(getCmd)
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

func runGet(cmd *cobra.Command, args []string) error {
	if changedFlag {
		if len(args) > 0 {
			return cobra.MaximumNArgs(0)(cmd, args)
		}
		if getJsonFlag {
			return getChangedJSON()
		}
		return runOnChangedModulesWithPath(func(moduleAbsPath string, stdout, stderr io.Writer) error {
			details, err := getModuleDetails(moduleAbsPath)
			if err != nil {
				_, _ = fmt.Fprintf(stderr, "failed to get module details: %v\n", err)
				return err
			}
			printModuleDetailsToWriter(stdout, details)
			return nil
		})
	}

	targetPath, err := resolveTargetPath(args)
	if err != nil {
		return err
	}

	details, err := getModuleDetails(targetPath)
	if err != nil {
		return err
	}

	if getJsonFlag {
		return printModuleDetailsJSON(details)
	}

	printModuleDetails(details)
	return nil
}

func getChangedJSON() error {
	modules, err := detectChangedModules(refFlag)
	if err != nil {
		return err
	}

	basePath, err := getBasePath()
	if err != nil {
		return err
	}

	results := make([]*ModuleDetails, 0)
	for _, mod := range modules {
		absPath := filepath.Join(basePath, mod.Path)
		details, err := getModuleDetails(absPath)
		if err != nil {
			continue
		}
		results = append(results, details)
	}

	output, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(output))
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

func printModuleDetailsToWriter(w io.Writer, details *ModuleDetails) {
	_, _ = fmt.Fprintf(w, "Name:                  %s\n", details.Name)
	_, _ = fmt.Fprintf(w, "Type:                  %s\n", formatType(details.Type))
	_, _ = fmt.Fprintf(w, "Path:                  %s\n", details.Path)
	_, _ = fmt.Fprintf(w, "Spacelift Version:     %s\n", details.SpaceliftVersion)
	_, _ = fmt.Fprintf(w, "Has Submodules:        %s\n", formatBool(details.HasSubmodules))
	_, _ = fmt.Fprintf(w, "Has Tests:             %s\n", formatBool(details.HasTests))
	_, _ = fmt.Fprintf(w, "Has Examples:          %s\n", formatBool(details.HasExamples))

	if len(details.Submodules) > 0 {
		_, _ = fmt.Fprintln(w, "\nSubmodules:")
		for _, ex := range details.Submodules {
			_, _ = fmt.Fprintf(w, "  - %s (%s)\n", ex.Name, ex.Path)
		}
	}

	if len(details.Examples) > 0 {
		_, _ = fmt.Fprintln(w, "\nExamples:")
		for _, ex := range details.Examples {
			_, _ = fmt.Fprintf(w, "  - %s (%s)\n", ex.Name, ex.Path)
		}
	}

	if len(details.Tests) > 0 {
		_, _ = fmt.Fprintln(w, "\nTests:")
		for _, ex := range details.Tests {
			_, _ = fmt.Fprintf(w, "  - %s (%s)\n", ex.Name, ex.Path)
		}
	}
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
