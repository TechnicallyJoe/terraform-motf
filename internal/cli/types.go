package cli

// Module directory constants
const (
	DirComponents = "components"
	DirBases      = "bases"
	DirProjects   = "projects"
)

// Subdirectory constants
const (
	DirExamples  = "examples"
	DirModules   = "modules"
	DirTests     = "tests"
	DirSpacelift = ".spacelift"
)

// File constants
const (
	FileSpaceliftConfig = "config.yml"
)

// Module type constants
const (
	TypeComponent = "component"
	TypeBase      = "base"
	TypeProject   = "project"
)

// ModuleDirs contains all module directory names
var ModuleDirs = []string{DirComponents, DirBases, DirProjects}

// ModuleTypeOrder defines the sorting order for module types
var ModuleTypeOrder = map[string]int{
	TypeComponent: 1,
	TypeBase:      2,
	TypeProject:   3,
}

// ModuleInfo holds information about a discovered module
type ModuleInfo struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Path    string `json:"path"`
	Version string `json:"version,omitempty"`
}
