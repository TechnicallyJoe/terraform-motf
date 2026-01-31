# Command Reference

This page documents all motf commands, their flags, and usage examples.

## Global Flags

These flags are available on all commands:

| Flag | Example | Description |
|------|---------|-------------|
| `--config` | `motf config --config /path/to/.motf.yml` | Path to config file (default: searches for `.motf.yml`) |
| `--path` | `motf fmt --path /path/to/module` | Explicit path to module (mutually exclusive with module name) |
| `-a`, `--args` | `motf plan storage-account -a -var="env=prod"` | Extra arguments to pass to terraform/tofu (repeatable) |
| `-h`, `--help` | `motf task -h` | Show help for any command |

## Parallel Execution Flags

These flags are available on commands that support `--changed`:

| Flag | Example | Description |
|------|---------|-------------|
| `-p`, `--parallel` | `motf fmt --changed --parallel` | Run commands in parallel across modules |
| `--max-parallel` | `motf val --changed -p --max-parallel 4` | Maximum parallel jobs (default: number of CPU cores) |

When parallel mode is enabled, output is prefixed with the module name and timestamp for clarity:

```
storage-account | 14:32:01.123 # Running 'terraform fmt'...
argocd-base     | 14:32:01.125 # Running 'terraform fmt'...
storage-account | 14:32:01.456 # Format complete
argocd-base     | 14:32:01.789 # Format complete
```

---

## init

Run `terraform init` or `tofu init` on a module. Relevant commands `fmt, val, plan` support `-i/--init` flag to run init beforehand.

```bash
motf init <module-name> [flags]
```

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--example` | `-e` | Run on a specific example instead of the module |
| `--changed` | | Run on all modules changed compared to `--ref` |
| `--ref` | | Git ref to compare against (default: auto-detect) |
| `--parallel` | `-p` | Run commands in parallel across modules |
| `--max-parallel` | | Maximum parallel jobs (default: number of CPU cores) |

### Examples

```bash
# Init a module by name
motf init storage-account

# Init an example within a module
motf init storage-account -e basic

# Init all changed modules
motf init --changed

# Init changed modules in parallel
motf init --changed --parallel

# Init changed modules compared to specific branch
motf init --changed --ref origin/main

# Pass extra arguments to terraform init
motf init storage-account -a -upgrade -a -reconfigure
```

---

## fmt

Run `terraform fmt` or `tofu fmt` on a module.

```bash
motf fmt <module-name> [flags]
```

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--init` | `-i` | Run init before formatting |
| `--example` | `-e` | Run on a specific example instead of the module |
| `--changed` | | Run on all modules changed compared to `--ref` |
| `--ref` | | Git ref to compare against (default: auto-detect) |
| `--parallel` | `-p` | Run commands in parallel across modules |
| `--max-parallel` | | Maximum parallel jobs (default: number of CPU cores) |

### Examples

```bash
# Format a module
motf fmt storage-account

# Format with init first
motf fmt -i storage-account

# Format an example named 'basic'
motf fmt storage-account -e basic

# Format all changed modules
motf fmt --changed

# Format all changed modules in parallel
motf fmt --changed --parallel

# Check formatting without modifying
motf fmt storage-account -a -check

# Check formatting on all changed modules in parallel
motf fmt --changed -p -a -check
```

---

## validate

Run `terraform validate` or `tofu validate` on a module.

```bash
motf val <module-name> [flags]
motf validate <module-name> [flags]
```

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--init` | `-i` | Run init before validating |
| `--example` | `-e` | Run on a specific example instead of the module |
| `--changed` | | Run on all modules changed compared to `--ref` |
| `--ref` | | Git ref to compare against (default: auto-detect) |
| `--parallel` | `-p` | Run commands in parallel across modules |
| `--max-parallel` | | Maximum parallel jobs (default: number of CPU cores) |

### Examples

```bash
# Validate a module (requires prior init)
motf val storage-account

# Validate with init
motf val -i storage-account

# Validate an example
motf val storage-account -e basic

# Validate all changed modules with init
motf val -i --changed

# Validate all changed modules in parallel
motf val -i --changed --parallel

# Validate against specific ref
motf val --changed --ref origin/develop
```

---

## plan

Run `terraform plan` or `tofu plan` on a module.

```bash
motf plan <module-name> [flags]
```

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--init` | `-i` | Run init before planning |
| `--example` | `-e` | Run on a specific example instead of the module |
| `--changed` | | Run on all modules changed compared to `--ref` |
| `--ref` | | Git ref to compare against (default: auto-detect) |
| `--parallel` | `-p` | Run commands in parallel across modules |
| `--max-parallel` | | Maximum parallel jobs (default: number of CPU cores) |

### Examples

```bash
# Plan a module
motf plan storage-account

# Plan with init
motf plan -i storage-account

# Plan an example
motf plan storage-account -e basic

# Plan all changed modules in parallel
motf plan --changed --parallel

# Plan with extra arguments
motf plan storage-account -a -var="env=prod"
```

---

## test

Run tests on a module using the configured test engine.

```bash
motf test <module-name> [flags]
```

The test engine is configured in `.motf.yml` (default: `terratest`).

For more info see [configuration -> test engines](configuration#test-engines)

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--changed` | | Run tests on all modules changed compared to `--ref` |
| `--ref` | | Git ref to compare against (default: auto-detect) |
| `--parallel` | `-p` | Run commands in parallel across modules |
| `--max-parallel` | | Maximum parallel jobs (default: number of CPU cores) |

### Examples

```bash
# Run tests on a module
motf test storage-account

# Run with verbose output
motf test storage-account -a -v

# Run with timeout
motf test storage-account -a -timeout=30m

# Run tests on all changed modules in parallel
motf test --changed --parallel

# Run specific test
motf test storage-account -a -run=TestBasic

# Combine multiple arguments
motf test storage-account -a -v -a -timeout=30m -a -run=TestBasic

# Test all changed modules
motf test --changed
```

Default arguments can also be supplied through the `test.args` configuration in `.motf.yml`.

---

## list

List all modules in the repository.

```bash
motf list [flags]
```

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--search` | `-s` | Filter modules (supports wildcards '*') |
| `--json` | | Output in JSON format |

### Examples

```bash
# List all modules
motf list

# Filter by name
motf list -s storage
motf list -s *account*
motf list -s azure*

# Output as JSON
motf list --json

# Combine search and JSON
motf list -s *storage* --json
```

### Output

```
NAME             TYPE       PATH                                  VERSION
storage-account  component  components/azurerm/storage-account    1.2.3
key-vault        component  components/azurerm/key-vault
resource-group   component  components/azurerm/resource-group
k8s-argocd       base       bases/k8s-argocd
prod-infra       project    projects/prod-infra
```

---

## get

Get detailed information about a module.

```bash
motf get <module-name> [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--json` | Output in JSON format |

### Examples

```bash
# Get module details
motf get storage-account

# Get details for module at explicit path
motf get --path ./my-module

# Output as JSON
motf get storage-account --json
```

### Output

```
Name:                  storage-account
Type:                  component
Path:                  components/azurerm/storage-account
Spacelift Version:     1.2.3
Has Submodules:        No
Has Tests:             Yes
Has Examples:          Yes

Examples:
  - basic (components/azurerm/storage-account/examples/basic)

Tests:
  - basic_test.go (components/azurerm/storage-account/tests/basic_test.go)
```

---

## describe

Describe the interface of a Terraform module (inputs, outputs, providers).

```bash
motf describe <module-name> [flags]
```

### Flags

| Flag | Description |
|------|-------------|
| `--json` | Output in JSON format |

### Examples

```bash
# Describe a module
motf describe storage-account

# Describe module at explicit path
motf describe --path ./my-module

# Output as JSON
motf describe storage-account --json
```

### Output

```
Module: storage-account
Path:   components/azurerm/storage-account

Terraform Version: >= 1.0.0

Example usage:
  module "storage-account" {
    source = "path/to/storage-account"

    name                = <required>
    resource_group_name = <required>
  }

Providers:
  NAME     VERSION
  azurerm  >= 3.0

Variables:
  NAME                 TYPE    DEFAULT  DESCRIPTION
  name                 string  -        The name of the storage account
  resource_group_name  string  -        Name of the resource group
  location             string  eastus   Azure region

Outputs:
  NAME        DESCRIPTION
  id          The ID of the storage account
  primary_key The primary access key
```

---

## changed

List modules that have changed compared to a git ref.

```bash
motf changed [flags]
```

Detects both committed and uncommitted changes. Useful for CI pipelines to run commands only on affected modules.

### Flags

| Flag | Description |
|------|-------------|
| `--ref` | Git ref to compare against (default: auto-detect from `origin/HEAD`) |
| `--json` | Output in JSON format |
| `--names` | Output only module names, one per line |

### Git Ref Examples

| Ref | Description |
|-----|-------------|
| `origin/main` | Compare against main branch |
| `origin/develop` | Compare against develop branch |
| `HEAD~5` | Compare against 5 commits ago |
| `v1.0.0` | Compare against a tag |

### Examples

```bash
# Compare against auto-detected default branch
motf changed

# Compare against specific branch
motf changed --ref origin/main

# Compare against previous commits
motf changed --ref HEAD~3

# Output as JSON
motf changed --json

# Output only names (useful for scripting)
motf changed --names
```

### Output

```
NAME             TYPE       PATH
storage-account  component  components/azurerm/storage-account
resource-group   component  components/azurerm/resource-group
```

### Scripting with --names

```bash
# Loop through changed modules
for module in $(motf changed --names); do
  motf validate -i "$module"
done

# Using xargs
motf changed --names | xargs -I {} motf fmt {}

# Using built-in helper
motf fmt --changed
```

---

## task

Run a custom task defined in `.motf.yml`.

```bash
motf task <module-name> [flags]
```

See [Configuration](configuration#custom-tasks) for how to define tasks.

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--task` | `-t` | Name of the task to run |
| `--list` | `-l` | List available tasks |
| `--changed` | | Run task on all modules changed compared to `--ref` |
| `--ref` | | Git ref to compare against (default: auto-detect) |
| `--parallel` | `-p` | Run commands in parallel across modules |
| `--max-parallel` | | Maximum parallel jobs (default: number of CPU cores) |

### Examples

```bash
# List available tasks
motf task --list

# Run a specific task
motf task storage-account --task lint

# Run task on explicit path
motf task --path ./modules/x --task docs

# Run task on changed modules
motf task --changed --task lint

# Run task on changed modules in parallel
motf task --changed --task lint --parallel
```

---

## config

Show the current configuration.

```bash
motf config
```

### Output

```
Current configuration:
  Config: /path/to/repo/.motf.yml
  Root:   iac
  Binary: terraform

Test configuration:
  Engine: terratest
  Args:   -v -timeout=30m

Tasks:
  - lint: Run tflint on the module
  - docs: Generate documentation
```

---

## version

Print version information.

```bash
motf version
motf --version
motf -v
```

### Output

```
motf version 1.0.0
commit: abc1234
built:  2025-01-15T10:30:00Z
```

---

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | Error (invalid arguments, module not found, terraform error, etc.) |

When using `--changed` with runtime commands, motf exits with code 1 if any module fails.
