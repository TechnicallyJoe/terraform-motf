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
