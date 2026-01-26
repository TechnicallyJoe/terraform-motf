# Copilot Instructions for motf

## Project Overview

`motf` is a Go CLI tool for managing polylith-style Terraform repositories. It wraps terraform/tofu commands (`init`, `fmt`, `validate`, `test`) to work with modules organized into three categories: **components**, **bases**, and **projects**.

## Architecture

```
cmd/           → Cobra CLI commands (root.go, init.go, fmt.go, validate.go, test.go, list.go, show.go)
internal/
  config/      → .motf.yml configuration loading and validation
  finder/      → Module discovery via recursive directory walking
  terraform/   → Terraform/tofu command execution wrapper
demo/          → Test fixture with polylith structure (components/, bases/, projects/)
e2e/           → End-to-end tests that build the binary and run against demo/
```

### Key Data Flow
1. `cmd/root.go` → Loads config via `internal/config`, creates `terraform.Runner`
2. `cmd/helpers.go` → `resolveTargetPath()` uses `internal/finder` to locate modules by name
3. `internal/terraform/terraform.go` → Executes terraform/tofu with configured binary

### Module Types (defined in `cmd/types.go`)
- **components**: Reusable Terraform modules (e.g., `storage-account`)
- **bases**: Composable base configurations (e.g., `k8s-argocd`)
- **projects**: Deployable infrastructure projects

## Build & Test Commands

```bash
# Build
go build -o motf .

# Run all unit tests
go test ./...

# Run specific package tests
go test ./cmd/... -v
go test ./internal/... -v

# Run e2e tests (requires terraform in PATH)
cd e2e && go test -v

# Quick manual test against demo
./motf list          # from repo root with demo/ present
./motf show storage-account
```

## Code Conventions

### CLI Commands
- Each command lives in `cmd/<command>.go` with matching `cmd/<command>_test.go`
- Commands use `cobra.Command` with `RunE` returning errors (not `Run`)
- Register commands in `init()` via `rootCmd.AddCommand()`
- Flags: use `StringVarP`/`BoolVar` in `init()`, store in package-level vars

### Error Handling
Always wrap errors with context using `fmt.Errorf("...: %w", err)`:
```go
// ✓ Correct - wraps with context
return fmt.Errorf("failed to read config file: %w", err)

// ✗ Avoid - loses error chain
return fmt.Errorf("failed to read config file: %s", err.Error())
```

### Testing Patterns

**Unit tests** (`cmd/*_test.go`, `internal/*_test.go`):
- Test individual functions in isolation
- Use `t.TempDir()` for isolated filesystems
- Reset package-level flag vars after use (e.g., `argsFlag = []string{}`)
- Write tests to be **human-readable** and **easily refactorable** - prefer explicit assertions over clever abstractions

**E2E tests** (`e2e/e2e_test.go`):
- Test complete CLI commands from start to finish
- Build the binary fresh via `buildMotf(t)` helper
- Run against the `demo/` directory as real-world fixture
- Skip tests requiring terraform: `skipIfNoTerraform(t)`

### Demo Directory

The `demo/` directory is a **real polylith-style Terraform repository** used for testing:

```
demo/
├── components/azurerm/     # Nested component modules
│   ├── storage-account/    # Has tests/ and examples/
│   ├── key-vault/
│   └── resource-group/
├── bases/k8s-argocd/       # Base module
└── projects/prod-infra/    # Project module
```

**When to modify `demo/`:**
- Adding a new command that needs a fixture to test against
- Testing edge cases (name clashes, nested directories, missing files)
- Validating module discovery behavior

**Important:** Keep `demo/` minimal but representative. Each module should have at least a `main.tf` with valid (even if empty) Terraform.

### Module Discovery
- Modules must contain `.tf` or `.tf.json` files to be recognized
- Skip directories: `.terraform`, `.git`, `node_modules`, `examples`, `modules`, `tests`
- Name clashes (same module name in multiple locations) produce explicit errors

### Configuration
- Config file: `.motf.yml` (optional, searched up to git root)
- Valid binaries: `terraform` or `tofu` only
- Test engines: `terratest` (runs `go test ./...`), `terraform`, or `tofu`

## Key Files Reference

| File | Purpose |
|------|---------|
| [cmd/root.go](cmd/root.go) | Entry point, global flags, config loading |
| [cmd/helpers.go](cmd/helpers.go) | `resolveTargetPath()`, module type detection |
| [cmd/types.go](cmd/types.go) | Constants for module dirs/types, `ModuleInfo` struct |
| [internal/finder/finder.go](internal/finder/finder.go) | `FindModule()`, `ListAllModules()` |
| [internal/terraform/terraform.go](internal/terraform/terraform.go) | `Runner` with `RunInit/Fmt/Validate/Test` |
| [demo/](demo/) | Test fixture - always test changes against this |

## Common Tasks

### Adding a New Command
1. Create `cmd/<name>.go` with `var <name>Cmd = &cobra.Command{...}`
2. Add `init()` with `rootCmd.AddCommand(<name>Cmd)`
3. Create `cmd/<name>_test.go` with unit tests
4. Add e2e test case in `e2e/e2e_test.go`

### Modifying Module Discovery
Edit [internal/finder/finder.go](internal/finder/finder.go). Update `skipDirs` map if new directories should be excluded.

### Adding Configuration Options
1. Add field to `Config` struct in [internal/config/config.go](internal/config/config.go)
2. Update `DefaultConfig()` with default value
3. Add validation in `Load()` if needed

## Commit Message Convention

This project uses [Conventional Commits](https://www.conventionalcommits.org/). All commits should follow this format:

```
<type>(<scope>): <description>

[optional body]
```

### Types
| Type | Description | Example |
|------|-------------|---------|
| `feat` | New feature | `feat(tasks): add shell configuration support` |
| `fix` | Bug fix | `fix(finder): handle symlinks correctly` |
| `docs` | Documentation | `docs: update README with examples` |
| `chore` | Maintenance | `chore(deps): update cobra to v1.9.0` |
| `refactor` | Code refactoring | `refactor(cmd): extract helper functions` |
| `test` | Tests | `test(e2e): add plan command tests` |

### Breaking Changes
Add `!` after type or include `BREAKING CHANGE:` in footer:
```
feat!: change task command argument order
```

### Why Conventional Commits
- GoReleaser auto-generates changelogs grouped by type
- Clear git history for contributors
- Enables automated versioning if needed

## Maintaining This Document

**Update this file when refactoring yields specific patterns.** Examples of patterns worth documenting:
- Error handling conventions (like the `%w` wrapping pattern)
- Structural decisions (like separating each command into its own `.go` file)
- Testing approaches that improve readability or maintainability
- New skip directories added to module discovery

