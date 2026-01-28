package git

import (
	"reflect"
	"sort"
	"testing"
)

func TestMapFilesToModules(t *testing.T) {
	tests := []struct {
		name         string
		changedFiles []string
		moduleDirs   []string
		want         []string
	}{
		{
			name: "single component change",
			changedFiles: []string{
				"components/azurerm/storage-account/main.tf",
			},
			moduleDirs: []string{"components", "bases", "projects"},
			want:       []string{"components/azurerm/storage-account"},
		},
		{
			name: "multiple files in same module",
			changedFiles: []string{
				"components/azurerm/storage-account/main.tf",
				"components/azurerm/storage-account/variables.tf",
				"components/azurerm/storage-account/outputs.tf",
			},
			moduleDirs: []string{"components", "bases", "projects"},
			want:       []string{"components/azurerm/storage-account"},
		},
		{
			name: "changes in multiple modules",
			changedFiles: []string{
				"components/azurerm/storage-account/main.tf",
				"components/azurerm/key-vault/main.tf",
				"projects/prod-infra/main.tf",
			},
			moduleDirs: []string{"components", "bases", "projects"},
			want: []string{
				"components/azurerm/key-vault",
				"components/azurerm/storage-account",
				"projects/prod-infra",
			},
		},
		{
			name: "ignores files outside module dirs",
			changedFiles: []string{
				"README.md",
				".github/workflows/ci.yml",
				"components/azurerm/storage-account/main.tf",
			},
			moduleDirs: []string{"components", "bases", "projects"},
			want:       []string{"components/azurerm/storage-account"},
		},
		{
			name: "base module change",
			changedFiles: []string{
				"bases/k8s-argocd/main.tf",
			},
			moduleDirs: []string{"components", "bases", "projects"},
			want:       []string{"bases/k8s-argocd"},
		},
		{
			name:         "no module changes",
			changedFiles: []string{"README.md", "go.mod"},
			moduleDirs:   []string{"components", "bases", "projects"},
			want:         nil,
		},
		{
			name:         "empty input",
			changedFiles: []string{},
			moduleDirs:   []string{"components", "bases", "projects"},
			want:         nil,
		},
		{
			name: "deeply nested component",
			changedFiles: []string{
				"components/azurerm/networking/vnet/main.tf",
			},
			moduleDirs: []string{"components", "bases", "projects"},
			want:       []string{"components/azurerm/networking/vnet"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapFilesToModules(tt.changedFiles, tt.moduleDirs)
			sort.Strings(got)
			sort.Strings(tt.want)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MapFilesToModules() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractModulePath(t *testing.T) {
	tests := []struct {
		name      string
		filePath  string
		moduleDir string
		want      string
	}{
		{
			name:      "standard component path",
			filePath:  "components/azurerm/storage-account/main.tf",
			moduleDir: "components",
			want:      "components/azurerm/storage-account",
		},
		{
			name:      "project path",
			filePath:  "projects/prod-infra/main.tf",
			moduleDir: "projects",
			want:      "projects/prod-infra",
		},
		{
			name:      "file directly in module dir",
			filePath:  "components/main.tf",
			moduleDir: "components",
			want:      "",
		},
		{
			name:      "nested path",
			filePath:  "components/aws/ec2/instance/main.tf",
			moduleDir: "components",
			want:      "components/aws/ec2/instance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractModulePath(tt.filePath, tt.moduleDir)
			if got != tt.want {
				t.Errorf("extractModulePath() = %v, want %v", got, tt.want)
			}
		})
	}
}
