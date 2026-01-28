package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

// setupTestRepo creates a temporary git repository for testing.
// It returns the repo path and a cleanup function.
func setupTestRepo(t *testing.T) string {
	t.Helper()

	tmpDir := t.TempDir()

	// Initialize git repo
	runGit(t, tmpDir, "init")
	runGit(t, tmpDir, "config", "user.email", "test@example.com")
	runGit(t, tmpDir, "config", "user.name", "Test User")

	return tmpDir
}

// runGit runs a git command in the given directory.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v\nOutput: %s", args, err, output)
	}
}

// writeFile creates a file with the given content.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("failed to create dir %s: %v", dir, err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file %s: %v", path, err)
	}
}

func TestGetChangedFiles_CommittedChanges(t *testing.T) {
	repoDir := setupTestRepo(t)

	// Create initial commit
	writeFile(t, filepath.Join(repoDir, "initial.txt"), "initial content")
	runGit(t, repoDir, "add", "-A")
	runGit(t, repoDir, "commit", "-m", "initial commit")

	// Create a branch to use as base
	runGit(t, repoDir, "branch", "base")

	// Make changes on HEAD
	writeFile(t, filepath.Join(repoDir, "components", "storage", "main.tf"), "# storage module")
	runGit(t, repoDir, "add", "-A")
	runGit(t, repoDir, "commit", "-m", "add storage component")

	// Get changed files
	files, err := GetChangedFiles(repoDir, "base")
	if err != nil {
		t.Fatalf("GetChangedFiles failed: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("expected 1 changed file, got %d: %v", len(files), files)
	}

	expected := "components/storage/main.tf"
	if len(files) > 0 && files[0] != expected {
		t.Errorf("expected %q, got %q", expected, files[0])
	}
}

func TestGetChangedFiles_UncommittedChanges(t *testing.T) {
	repoDir := setupTestRepo(t)

	// Create initial commit
	writeFile(t, filepath.Join(repoDir, "initial.txt"), "initial content")
	runGit(t, repoDir, "add", "-A")
	runGit(t, repoDir, "commit", "-m", "initial commit")

	// Make uncommitted changes (staged)
	writeFile(t, filepath.Join(repoDir, "staged.tf"), "# staged")
	runGit(t, repoDir, "add", "staged.tf")

	// Make uncommitted changes (unstaged)
	writeFile(t, filepath.Join(repoDir, "unstaged.tf"), "# unstaged")

	// Get changed files (compare HEAD to HEAD, so only uncommitted show)
	files, err := GetChangedFiles(repoDir, "HEAD")
	if err != nil {
		t.Fatalf("GetChangedFiles failed: %v", err)
	}

	sort.Strings(files)
	expected := []string{"staged.tf", "unstaged.tf"}

	if !reflect.DeepEqual(files, expected) {
		t.Errorf("expected %v, got %v", expected, files)
	}
}

func TestGetChangedFiles_InvalidRef(t *testing.T) {
	repoDir := setupTestRepo(t)

	// Create initial commit
	writeFile(t, filepath.Join(repoDir, "initial.txt"), "initial content")
	runGit(t, repoDir, "add", "-A")
	runGit(t, repoDir, "commit", "-m", "initial commit")

	// Try to get changes against non-existent ref
	// Should not error, just return uncommitted changes only
	_, err := GetChangedFiles(repoDir, "nonexistent-branch")
	if err != nil {
		t.Errorf("expected no error for missing ref, got: %v", err)
	}
}

func TestGetChangedFiles_InvalidRepo(t *testing.T) {
	tmpDir := t.TempDir()

	_, err := GetChangedFiles(tmpDir, "HEAD")
	if err == nil {
		t.Error("expected error for non-git directory")
	}
}

func TestGetRepoRoot(t *testing.T) {
	repoDir := setupTestRepo(t)

	// Change to a subdirectory
	subDir := filepath.Join(repoDir, "subdir", "nested")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	// Save current dir and change to subdir
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()

	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	root, err := GetRepoRoot()
	if err != nil {
		t.Fatalf("GetRepoRoot failed: %v", err)
	}

	// The root should be the repoDir
	if root != repoDir {
		t.Errorf("expected %q, got %q", repoDir, root)
	}
}

func TestGetDefaultBranch_FallbackToMain(t *testing.T) {
	repoDir := setupTestRepo(t)

	// Create initial commit on main
	writeFile(t, filepath.Join(repoDir, "initial.txt"), "content")
	runGit(t, repoDir, "add", "-A")
	runGit(t, repoDir, "commit", "-m", "initial")

	// Rename to main
	runGit(t, repoDir, "branch", "-M", "main")

	// Add a remote (fake, just for refs)
	runGit(t, repoDir, "remote", "add", "origin", repoDir)
	runGit(t, repoDir, "fetch", "origin")

	// Save current dir and change to repo
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()

	if err := os.Chdir(repoDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	branch, err := GetDefaultBranch()
	if err != nil {
		t.Fatalf("GetDefaultBranch failed: %v", err)
	}

	if branch != "origin/main" {
		t.Errorf("expected origin/main, got %q", branch)
	}
}

func TestGetDefaultBranch_FallbackToMaster(t *testing.T) {
	repoDir := setupTestRepo(t)

	// Create initial commit on master
	writeFile(t, filepath.Join(repoDir, "initial.txt"), "content")
	runGit(t, repoDir, "add", "-A")
	runGit(t, repoDir, "commit", "-m", "initial")

	// Ensure branch is named master
	runGit(t, repoDir, "branch", "-M", "master")

	// Add a remote
	runGit(t, repoDir, "remote", "add", "origin", repoDir)
	runGit(t, repoDir, "fetch", "origin")

	// Save current dir and change to repo
	origDir, _ := os.Getwd()
	defer func() { _ = os.Chdir(origDir) }()

	if err := os.Chdir(repoDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	branch, err := GetDefaultBranch()
	if err != nil {
		t.Fatalf("GetDefaultBranch failed: %v", err)
	}

	if branch != "origin/master" {
		t.Errorf("expected origin/master, got %q", branch)
	}
}

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
