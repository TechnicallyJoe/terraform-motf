# Copilot Instructions for motf

## Project Overview

`motf` is a Go CLI tool for managing polylith-style Terraform repositories. It wraps terraform/tofu commands (`init`, `fmt`, `validate`, `test`) to work with modules organized into three categories: **components**, **bases**, and **projects**.

## Architecture

```
cmd/
  motf/        → Main entrypoint (imports internal/cli)
internal/
  cli/         → Cobra CLI commands (root.go, init.go, fmt.go, validate.go, test.go, plan.go, list.go, get.go, describe.go, changed.go, task.go)
  config/      → .motf.yml configuration loading and validation
  finder/      → Module discovery via recursive directory walking
  git/         → Git operations for change detection (uses go-git library)
  spacelift/   → Spacelift stack configuration discovery
  tasks/       → Custom task configuration loading from .motf.yml
  terraform/   → Terraform/tofu command execution wrapper
demo/          → Test fixture with polylith structure (components/, bases/, projects/)
e2e/           → End-to-end tests that build the binary and run against demo/
```

### Key Data Flow
1. `internal/cli/root.go` → Loads config via `internal/config`, creates `terraform.Runner`
2. `internal/cli/helpers.go` → `resolveTargetPath()` uses `internal/finder` to locate modules by name
3. `internal/terraform/terraform.go` → Executes terraform/tofu with configured binary
4. `internal/git/diff.go` → Detects changed modules via go-git (committed + uncommitted changes)
5. `internal/cli/changed_runner.go` → Runs commands on changed modules when `--changed` flag is used

### Module Types (defined in `internal/cli/types.go`)
- **components**: Reusable Terraform modules (e.g., `storage-account`)
- **bases**: Composable base configurations (e.g., `k8s-argocd`)
- **projects**: Deployable infrastructure projects

## Build & Test Commands

```bash
# Build
go build -o motf ./cmd/motf

# Run all unit tests
go test ./...

# Run specific package tests
go test ./cmd/... -v
go test ./internal/... -v

# Run e2e tests (requires terraform in PATH)
cd e2e && go test -v

# Quick manual test against demo
./motf list          # from repo root with demo/ present
./motf get storage-account

# Linting test
golangci-lint run
```

## Code Conventions

### CLI Commands
- Each command lives in `internal/cli/<command>.go` with matching `internal/cli/<command>_test.go`
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

**Unit tests** (`internal/cli/*_test.go`, `internal/*_test.go`):
- Test individual functions in isolation
- Use `t.TempDir()` for isolated filesystems
- Reset package-level flag vars after use (e.g., `argsFlag = []string{}`)
- Write tests to be **human-readable** and **easily refactorable** - prefer explicit assertions over clever abstractions

**E2E tests** (`e2e/e2e_test.go`):
- Test complete CLI commands from start to finish
- Build the binary fresh via `buildMotf(t)` helper
- Run against the `demo/` directory as real-world fixture
- Skip tests requiring tofu: `skipIfNoTofu(t)`
- For git-related tests, use `setupCleanGitRepo(t)` to create isolated git repositories

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
- Writing new e2e tests that require specific module setups
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
- Custom tasks: defined in `.motf.yml` under `tasks:` key

### Dependencies
- **go-git** (`github.com/go-git/go-git/v5`): Pure Go git library for change detection
- **cobra** (`github.com/spf13/cobra`): CLI framework
- **yaml.v3** (`gopkg.in/yaml.v3`): YAML parsing for config

## Key Files Reference

| File | Purpose |
|------|---------|
| [cmd/motf/main.go](cmd/motf/main.go) | Main entrypoint |
| [internal/cli/root.go](internal/cli/root.go) | CLI root, global flags, config loading |
| [internal/cli/helpers.go](internal/cli/helpers.go) | `resolveTargetPath()`, module type detection |
| [internal/cli/types.go](internal/cli/types.go) | Constants for module dirs/types, `ModuleInfo` struct |
| [internal/cli/changed.go](internal/cli/changed.go) | `motf changed` command implementation |
| [internal/cli/changed_runner.go](internal/cli/changed_runner.go) | Helper for `--changed` flag on commands |
| [internal/finder/finder.go](internal/finder/finder.go) | `FindModule()`, `ListAllModules()` |
| [internal/git/diff.go](internal/git/diff.go) | Git change detection with go-git library |
| [internal/tasks/tasks.go](internal/tasks/tasks.go) | Custom task loading from `.motf.yml` |
| [internal/terraform/terraform.go](internal/terraform/terraform.go) | `Runner` with `RunInit/Fmt/Validate/Test/Plan` |
| [demo/](demo/) | Test fixture - always test changes against this |

## Common Tasks

### Adding a New Command
1. Create `internal/cli/<name>.go` with `var <name>Cmd = &cobra.Command{...}`
2. Add `init()` with `rootCmd.AddCommand(<name>Cmd)`
3. Create `internal/cli/<name>_test.go` with unit tests
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

## Documentation

The `/docs/` directory contains user-facing documentation that is automatically published to the GitHub Wiki via the `publish-wiki.yml` workflow.

### Documentation Structure

| File | Wiki Page | Content |
|------|-----------|---------|
| `docs/home.md` | Home | Main landing page: intro to motf, core concepts (polylith architecture, module types), quick start guide, feature overview, links to other pages |
| `docs/commands.md` | Commands | Complete CLI reference: all 12 commands with flags, examples, output samples, compatibility matrix |
| `docs/configuration.md` | Configuration | `.motf.yml` reference: all options, defaults, test engines, custom tasks definition and examples |
| `docs/ci.md` | CI | CI/CD integration: GitHub Actions examples, `--changed` usage, matrix strategies, scripting patterns |
| `docs/assets/` | - | Images and media (demo.gif, etc.) synced to wiki for embedding |

### Updating Documentation

1. Edit files in `/docs/` directory
2. The `publish-wiki.yml` workflow automatically syncs changes to the GitHub Wiki

### Writing Guidelines

- Use lowercase filenames (`commands.md`, not `Commands.md`)
- Link between wiki pages using relative links: `[Commands](commands)`, `[Configuration](configuration#custom-tasks)`
- Reference images with `assets/` prefix: `![demo](assets/demo.gif)`
- Keep README.md high-level with links to wiki for detailed documentation

## Maintaining This Document

Ensure this instructions file stays up to date with architectural and coding conventions as the project evolves.
Examples of patterns worth documenting:

- When refactoring yields specific patterns
  - Error handling conventions (like the `%w` wrapping pattern)
  - Structural decisions (like separating each command into its own `.go` file)
  - Testing approaches that improve readability or maintainability
  - New skip directories added to module discovery
- Changes to configuration options or validation rules
- Updates to commit message conventions or release processes
- Update to documentation
- New commands added to the CLI (motf)
