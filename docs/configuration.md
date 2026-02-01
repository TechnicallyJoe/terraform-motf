# Configuration

motf uses a `.motf.yml` configuration file to customize its behavior. The configuration file is optional—motf works with sensible defaults if no config is present.

## Config File Location

motf searches for `.motf.yml` by walking up the directory tree from your current working directory until it reaches the git repository root. You can also specify a config file explicitly:

```bash
motf --config /path/to/.motf.yml list
```

## Configuration Options

### Full Example

```yaml
# Root directory containing components/, bases/, projects/
# Default: "" (repository root)
root: iac

# Terraform binary to use: "terraform" or "tofu"
# Default: "terraform"
binary: terraform

# Test configuration
test:
  # Test engine: "terratest", "terraform", or "tofu"
  # Default: "terratest"
  engine: terratest

  # Additional arguments passed to the test command
  # Default: ""
  args: "-v -timeout=30m"

# Parallelism configuration for --parallel flag
parallelism:
  # Maximum number of parallel jobs
  # Default: 0 (auto-detect based on CPU cores)
  max_jobs: 4

# Custom tasks (see Custom Tasks section below)
tasks:
  lint:
    description: "Run tflint on the module"
    command: "tflint --init && tflint"

  docs:
    description: "Generate terraform-docs"
    shell: bash
    command: |
      terraform-docs markdown table . > README.md
      echo "Documentation updated"
```

### Options Reference

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `root` | string | `""` | Directory containing `components/`, `bases/`, `projects/`. Relative paths are resolved from the config file location. |
| `binary` | string | `"terraform"` | Binary to use: `"terraform"` or `"tofu"` |
| `test.engine` | string | `"terratest"` | Test engine: `"terratest"`, `"terraform"`, or `"tofu"` |
| `test.args` | string | `""` | Additional arguments passed to the test command |
| `parallelism.max_jobs` | int | `0` | Maximum parallel jobs. `0` means auto-detect (number of CPU cores) |
| `tasks` | map | `{}` | Custom task definitions (see below) |

### Root Directory

The `root` option specifies where your module directories are located:

```yaml
# Modules are at repo-root/iac/components/, repo-root/iac/bases/, etc.
root: iac
```

```yaml
# Modules are at repo-root/components/, repo-root/bases/, etc.
root: ""
```

If `root` is a relative path, it's resolved relative to the config file location (not the current working directory).

### Binary Selection

Choose between `terraform` and `tofu`:

```yaml
binary: tofu
```

Only `terraform` and `tofu` are valid values. Any other value will cause a configuration error.

### Test Configuration

Configure how `motf test` runs tests:

```yaml
test:
  engine: terratest
  args: "-v -timeout=30m"
```

#### Test Engines

| Engine | Command Executed | Use Case |
|--------|------------------|----------|
| `terratest` | `go test ./... <args>` | Go-based Terratest tests |
| `terraform` | `terraform test <args>` | Native Terraform test files (`.tftest.hcl`) |
| `tofu` | `tofu test <args>` | Native OpenTofu test files |

#### Test Arguments

Arguments are combined in this order:
1. `test.args` from config (space-separated)
2. `-a` flags from command line

```yaml
test:
  engine: terratest
  args: "-v -timeout=30m"
```

```bash
# Executes: go test ./... -v -timeout=30m -run=TestBasic
motf test storage-account -a -run=TestBasic
```

---

## Parallelism Configuration

Configure default parallel execution behavior for commands with `--parallel` flag:

```yaml
parallelism:
  max_jobs: 4
```

### Options

| Option | Default | Description |
|--------|---------|-------------|
| `max_jobs` | `0` | Maximum concurrent jobs. `0` = auto-detect (uses number of CPU cores) |

### Priority Order

The effective max parallel jobs is determined in this order:

1. `--max-parallel` CLI flag (highest priority)
2. `parallelism.max_jobs` in `.motf.yml`
3. Auto-detect based on `runtime.NumCPU()` (default)

### Example

```yaml
parallelism:
  max_jobs: 8
```

```bash
# Uses 8 parallel jobs from config
motf fmt --changed --parallel

# Overrides to 2 parallel jobs
motf fmt --changed --parallel --max-parallel 2
```

### Output Format

When running in parallel mode, output is prefixed with module name and timestamp:

```
storage-account | 14:32:01.123 # Running 'terraform fmt'...
argocd-base     | 14:32:01.125 # Running 'terraform fmt'...
storage-account | 14:32:01.456 # Format complete
argocd-base     | 14:32:01.789 # Format complete
```

Each module is assigned a unique color for easier visual tracking.

---

## Custom Tasks

Custom tasks let you define shell commands that can be run on modules via `motf task`.

### Defining Tasks

Tasks are defined under the `tasks` key in `.motf.yml`:

```yaml
tasks:
  task-name:
    description: "Optional description shown in 'motf task --list'"
    shell: bash  # Optional, defaults to "sh"
    command: |
      echo "Commands to execute"
      pwd
```

### Task Options

| Option | Required | Default | Description |
|--------|----------|---------|-------------|
| `command` | Yes | - | Shell command(s) to execute |
| `description` | No | `""` | Description shown when listing tasks |
| `shell` | No | `"sh"` | Shell to use for execution |

### Supported Shells

| Shell | Binary | Arguments |
|-------|--------|-----------|
| `sh` | `sh` | `-c` |
| `bash` | `bash` | `-c` |
| `pwsh` | `pwsh` | `-Command` |
| `cmd` | `cmd` | `/C` |

### Examples

#### Simple Command

```yaml
tasks:
  lint:
    description: "Run tflint"
    command: "tflint --init && tflint"
```

```bash
motf task storage-account -t lint
```

#### Multi-line Commands

```yaml
tasks:
  docs:
    description: "Generate and validate documentation"
    shell: bash
    command: |
      echo "Generating docs for $(basename $PWD)"
      terraform-docs markdown table . > README.md
      if git diff --quiet README.md; then
        echo "Documentation is up to date"
      else
        echo "Documentation was updated"
      fi
```

#### Using Environment Variables

Tasks run in the module directory, so you can use the path context:

```yaml
tasks:
  info:
    command: |
      echo "Module: $(basename $PWD)"
      echo "Path: $PWD"
      echo "Files: $(ls *.tf | wc -l) terraform files"
```

### Built-in Variables

MOTF injects the following environment variables into every task execution:

| Variable | Description |
|----------|-------------|
| `MOTF_GIT_ROOT` | Absolute path to the git repository root (empty if not in a git repo) |
| `MOTF_MODULE_PATH` | Absolute path to the current module being processed |
| `MOTF_MODULE_NAME` | Name of the module (last component of the path, e.g., `storage-account`) |
| `MOTF_CONFIG_PATH` | Absolute path to the `.motf.yml` config file (empty if no config) |
| `MOTF_BINARY` | The terraform/tofu binary name (`terraform` or `tofu`) |

Example usage:

```yaml
tasks:
  generate-docs:
    description: "Generate docs using repo-level script"
    command: $MOTF_GIT_ROOT/scripts/generate-docs.sh $MOTF_MODULE_NAME

  lint:
    description: "Run linting with terraform binary"
    command: $MOTF_BINARY fmt -check $MOTF_MODULE_PATH

  info:
    description: "Show module info"
    command: |
      echo "Module: $MOTF_MODULE_NAME"
      echo "Path: $MOTF_MODULE_PATH"
      echo "Git root: $MOTF_GIT_ROOT"
```

#### Continuous Integrations

```yaml
# GitHub Actions example
tasks:
  pre-commit:
    description: "Run all pre-commit checks"
    shell: bash
    command: |
      set -e
      terraform fmt -check
      tflint --init && tflint
      terraform-docs markdown table --output-check . > /dev/null
```

### Running Tasks

```bash
# List available tasks
motf task --list

# Run a task
motf task storage-account --task lint

# Run on explicit path
motf task --path ./modules/x --task docs

# Run on all changed modules
motf task --task lint --changed
```

### Task Output Example

```bash
$ motf task --list

Available tasks:
  lint  Run tflint on the module
  docs  Generate terraform-docs
```

---

## Viewing Current Configuration

Use `motf config` to see the active configuration:

```bash
$ motf config
Config: /home/user/repo/.motf.yml

Settings:
  Root:   iac
  Binary: terraform

Test:
  Engine: terratest
  Args:   -v -timeout=30m

Tasks:
  - lint           Run tflint on the module
  - docs           Generate terraform-docs
```

If no config file is found:

```bash
$ motf config

Settings:
  Config: (none)
  Root:   (repository root)
  Binary: terraform

Test:
  Engine: terratest
  Args:   (none)

Tasks:
  (none)
```
