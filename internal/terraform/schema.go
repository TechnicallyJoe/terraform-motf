package terraform

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-config-inspect/tfconfig"
)

// ProviderInfo represents a required provider
type ProviderInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// VariableInfo represents a module variable
type VariableInfo struct {
	Name        string `json:"name"`
	Type        string `json:"type,omitempty"`
	Default     any    `json:"default,omitempty"`
	Required    bool   `json:"required"`
	Description string `json:"description,omitempty"`
}

// DefaultString returns a formatted string representation of the variable's default value.
// Complex types (objects, maps, lists) are simplified for table display.
func (v VariableInfo) DefaultString() string {
	if v.Required {
		return "(required)"
	}
	if v.Default == nil {
		return "null"
	}
	switch val := v.Default.(type) {
	case string:
		if val == "" {
			return `""`
		}
		return fmt.Sprintf(`"%s"`, val)
	case bool:
		return fmt.Sprintf("%t", val)
	case float64:
		return fmt.Sprintf("%g", val)
	case []any:
		if len(val) == 0 {
			return "[]"
		}
		return "[...]"
	case map[string]any:
		if len(val) == 0 {
			return "{}"
		}
		return "{...}"
	default:
		// For other complex types, check if it's a slice or map
		if b, err := json.Marshal(val); err == nil {
			s := string(b)
			if s == "[]" || s == "{}" {
				return s
			}
			if strings.HasPrefix(s, "[") {
				return "[...]"
			}
			if strings.HasPrefix(s, "{") {
				return "{...}"
			}
		}
		return fmt.Sprintf("%v", val)
	}
}

// FullDefaultString returns the complete default value without truncation.
// Used for generating example module blocks.
func (v VariableInfo) FullDefaultString() string {
	if v.Default == nil {
		return "null"
	}
	switch val := v.Default.(type) {
	case string:
		if val == "" {
			return `""`
		}
		return fmt.Sprintf(`"%s"`, val)
	case bool:
		return fmt.Sprintf("%t", val)
	case float64:
		return fmt.Sprintf("%g", val)
	default:
		// For complex types, format as HCL-like representation
		if b, err := json.Marshal(val); err == nil {
			return string(b)
		}
		return fmt.Sprintf("%v", val)
	}
}

// EmptyValueForType returns an appropriate placeholder value for the given Terraform type.
func (v VariableInfo) EmptyValueForType() string {
	switch {
	case v.Type == "string" || v.Type == "":
		return `""`
	case v.Type == "number":
		return "0"
	case v.Type == "bool":
		return "false"
	case strings.HasPrefix(v.Type, "list") || strings.HasPrefix(v.Type, "set"):
		return "[]"
	case strings.HasPrefix(v.Type, "map") || strings.HasPrefix(v.Type, "object"):
		return "{}"
	default:
		return "null"
	}
}

// OutputInfo represents a module output
type OutputInfo struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Sensitive   bool   `json:"sensitive,omitempty"`
}

// ModuleSchema represents the parsed Terraform module schema
type ModuleSchema struct {
	Name             string         `json:"name"`
	Path             string         `json:"path"`
	TerraformVersion string         `json:"terraform_version,omitempty"`
	Providers        []ProviderInfo `json:"providers,omitempty"`
	Variables        []VariableInfo `json:"variables,omitempty"`
	Outputs          []OutputInfo   `json:"outputs,omitempty"`
}

// LoadModuleSchema parses a Terraform module and returns its schema.
// If rootPath is provided, the schema.Path will be made relative to it.
func LoadModuleSchema(modulePath string, rootPath string) (*ModuleSchema, error) {
	module, diags := tfconfig.LoadModule(modulePath)
	if diags.HasErrors() {
		return nil, diags.Err()
	}

	return buildModuleSchema(module, modulePath, rootPath), nil
}

func buildModuleSchema(module *tfconfig.Module, modulePath string, rootPath string) *ModuleSchema {
	schema := &ModuleSchema{
		Name: filepath.Base(modulePath),
		Path: modulePath,
	}

	// Make path relative to rootPath if possible
	if rootPath != "" {
		if rel, err := filepath.Rel(rootPath, modulePath); err == nil {
			schema.Path = rel
		}
	}

	// Required Terraform version
	if len(module.RequiredCore) > 0 {
		schema.TerraformVersion = strings.Join(module.RequiredCore, ", ")
	}

	// Required providers (sorted by name)
	schema.Providers = buildProviderList(module.RequiredProviders)

	// Variables (sorted: required first, then alphabetically)
	schema.Variables = buildVariableList(module.Variables)

	// Outputs (sorted by name)
	schema.Outputs = buildOutputList(module.Outputs)

	return schema
}

func buildProviderList(providers map[string]*tfconfig.ProviderRequirement) []ProviderInfo {
	names := make([]string, 0, len(providers))
	for name := range providers {
		names = append(names, name)
	}
	sort.Strings(names)

	result := make([]ProviderInfo, 0, len(providers))
	for _, name := range names {
		req := providers[name]
		version := ""
		if len(req.VersionConstraints) > 0 {
			version = strings.Join(req.VersionConstraints, ", ")
		}
		result = append(result, ProviderInfo{
			Name:    name,
			Version: version,
		})
	}
	return result
}

func buildVariableList(variables map[string]*tfconfig.Variable) []VariableInfo {
	// Collect all variables
	vars := make([]VariableInfo, 0, len(variables))
	for name, v := range variables {
		vars = append(vars, VariableInfo{
			Name:        name,
			Type:        v.Type,
			Default:     v.Default,
			Required:    v.Required,
			Description: v.Description,
		})
	}

	// Sort: required first, then alphabetically by name
	sort.Slice(vars, func(i, j int) bool {
		if vars[i].Required != vars[j].Required {
			return vars[i].Required // required (true) comes before optional (false)
		}
		return vars[i].Name < vars[j].Name
	})

	return vars
}

func buildOutputList(outputs map[string]*tfconfig.Output) []OutputInfo {
	names := make([]string, 0, len(outputs))
	for name := range outputs {
		names = append(names, name)
	}
	sort.Strings(names)

	result := make([]OutputInfo, 0, len(outputs))
	for _, name := range names {
		o := outputs[name]
		result = append(result, OutputInfo{
			Name:        name,
			Description: o.Description,
			Sensitive:   o.Sensitive,
		})
	}
	return result
}
