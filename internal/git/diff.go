package git

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

// GetChangedFiles returns a list of files that have changed between the base ref and HEAD,
// including any uncommitted changes in the working directory.
func GetChangedFiles(repoRoot, base string) ([]string, error) {
	repo, err := git.PlainOpen(repoRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}

	fileSet := make(map[string]bool)

	// Get committed changes between base and HEAD
	committedFiles, err := getCommittedChanges(repo, base)
	if err != nil {
		// If we can't get committed changes (e.g., base doesn't exist), continue with uncommitted only
		// This allows the command to work even on initial commits
		if !strings.Contains(err.Error(), "reference not found") {
			return nil, err
		}
	}
	for _, f := range committedFiles {
		fileSet[f] = true
	}

	// Get uncommitted changes (staged + unstaged)
	uncommittedFiles, err := getUncommittedChanges(repo)
	if err != nil {
		return nil, err
	}
	for _, f := range uncommittedFiles {
		fileSet[f] = true
	}

	// Convert set to slice
	var files []string
	for f := range fileSet {
		files = append(files, f)
	}

	return files, nil
}

// getCommittedChanges returns files changed between base ref and HEAD.
func getCommittedChanges(repo *git.Repository, base string) ([]string, error) {
	// Resolve base reference
	baseHash, err := repo.ResolveRevision(plumbing.Revision(base))
	if err != nil {
		return nil, fmt.Errorf("failed to resolve base ref '%s': %w", base, err)
	}

	// Get HEAD
	headRef, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Get commits
	baseCommit, err := repo.CommitObject(*baseHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get base commit: %w", err)
	}

	headCommit, err := repo.CommitObject(headRef.Hash())
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD commit: %w", err)
	}

	// Get trees
	baseTree, err := baseCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get base tree: %w", err)
	}

	headTree, err := headCommit.Tree()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD tree: %w", err)
	}

	// Compute diff
	changes, err := baseTree.Diff(headTree)
	if err != nil {
		return nil, fmt.Errorf("failed to compute diff: %w", err)
	}

	var files []string
	for _, change := range changes {
		// Include both old and new paths (for renames)
		if change.From.Name != "" {
			files = append(files, change.From.Name)
		}
		if change.To.Name != "" && change.To.Name != change.From.Name {
			files = append(files, change.To.Name)
		}
	}

	return files, nil
}

// getUncommittedChanges returns files with uncommitted changes (staged + unstaged).
func getUncommittedChanges(repo *git.Repository) ([]string, error) {
	worktree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return nil, fmt.Errorf("failed to get status: %w", err)
	}

	var files []string
	for file, s := range status {
		// Include any file that has changes (staged or unstaged)
		if s.Staging != git.Unmodified || s.Worktree != git.Unmodified {
			files = append(files, file)
		}
	}

	return files, nil
}

// GetRepoRoot returns the root directory of the git repository.
func GetRepoRoot() (string, error) {
	// Start from current directory and walk up to find .git
	repo, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return "", fmt.Errorf("not a git repository: %w", err)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	return worktree.Filesystem.Root(), nil
}

// GetDefaultBranch attempts to determine the default branch of the repository.
// It checks origin/HEAD first, then falls back to common defaults (main, master).
func GetDefaultBranch() (string, error) {
	repo, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{
		DetectDotGit: true,
	})
	if err != nil {
		return "", fmt.Errorf("not a git repository: %w", err)
	}

	// Try origin/HEAD (symbolic ref that points to default branch)
	ref, err := repo.Reference(plumbing.NewRemoteHEADReferenceName("origin"), false)
	if err == nil && ref.Type() == plumbing.SymbolicReference {
		// ref.Target() returns the ref it points to, e.g., "refs/remotes/origin/main"
		target := ref.Target().String()
		parts := strings.Split(target, "/")
		if len(parts) > 0 {
			return "origin/" + parts[len(parts)-1], nil
		}
	}

	// Fallback: check if origin/main or origin/master exists
	for _, branch := range []string{"origin/main", "origin/master"} {
		refName := plumbing.NewRemoteReferenceName("origin", strings.TrimPrefix(branch, "origin/"))
		if _, err := repo.Reference(refName, true); err == nil {
			return branch, nil
		}
	}

	return "", fmt.Errorf("could not determine default branch")
}

// MapFilesToModules takes a list of changed files and returns a list of module directories
// that contain those files. It filters to only include paths that are within the given
// module directories (e.g., components/, bases/, projects/).
func MapFilesToModules(changedFiles []string, moduleDirs []string) []string {
	moduleSet := make(map[string]bool)

	for _, file := range changedFiles {
		// Normalize path separators
		file = filepath.ToSlash(file)

		// Check if file is within any of the module directories
		for _, moduleDir := range moduleDirs {
			if strings.HasPrefix(file, moduleDir+"/") {
				// Extract the module path: find the deepest directory that contains .tf files
				// For now, we'll extract the immediate subdirectory structure
				// e.g., components/azurerm/storage-account/main.tf -> components/azurerm/storage-account
				modulePath := extractModulePath(file, moduleDir)
				if modulePath != "" {
					moduleSet[modulePath] = true
				}
				break
			}
		}
	}

	// Convert set to sorted slice
	var modules []string
	for module := range moduleSet {
		modules = append(modules, module)
	}

	return modules
}

// extractModulePath extracts the module directory path from a file path.
// It returns the directory containing the file (the module directory).
func extractModulePath(filePath, moduleDir string) string {
	// Get the directory containing the file
	dir := filepath.Dir(filePath)

	// Normalize to forward slashes for consistency
	dir = filepath.ToSlash(dir)

	// The module path should be at least moduleDir + one more component
	if dir == moduleDir || !strings.HasPrefix(dir, moduleDir+"/") {
		return ""
	}

	return dir
}
