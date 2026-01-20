# tfpl - Terraform Polylith CLI

A command-line tool for working with polylith-style Terraform repositories. `tfpl` makes it easy to run terraform/tofu commands on components, bases, and projects organized in a polylith structure.

## Features

- **Simple commands**: Run `init`, `fmt`, and `validate` on components, bases, or projects
- **Configurable**: Support for both `terraform` and `tofu` via `.tfpl.yml`
- **Smart discovery**: Recursively finds modules in nested subdirectories
- **Clash detection**: Warns when multiple modules share the same name

## Installation

### Using go install

```bash
go install github.com/TechnicallyJoe/sturdy-parakeet@latest
```

The binary will be named `sturdy-parakeet`. You can rename it to `tfpl`:

```bash
mv $(go env GOPATH)/bin/sturdy-parakeet $(go env GOPATH)/bin/tfpl
```

### Building from source

```bash
git clone https://github.com/TechnicallyJoe/sturdy-parakeet.git
cd sturdy-parakeet
go build -o tfpl .
```

## Requirements

- Go 1.25 or later (for building)
- `terraform` or `tofu` CLI installed and available in PATH

## Usage

### Commands

#### `tfpl init`
Run `terraform init` or `tofu init` on a module:

```bash
tfpl init storage-account       # Init storage-account (searches components, bases, and projects)
tfpl init k8s-argocd           # Init k8s-argocd
tfpl init spacelift-modules    # Init spacelift-modules
```

#### `tfpl fmt`
Run `terraform fmt` or `tofu fmt` on a module:

```bash
tfpl fmt storage-account       # Format storage-account
tfpl fmt k8s-argocd           # Format k8s-argocd
tfpl fmt -i storage-account   # Init then format storage-account
```

#### `tfpl val` (or `validate`)
Run `terraform validate` or `tofu validate` on a module:

```bash
tfpl val storage-account       # Validate storage-account
tfpl validate k8s-argocd      # Validate k8s-argocd
tfpl val -i spacelift-modules # Init then validate spacelift-modules
```

#### `tfpl config`
Show current configuration:

```bash
tfpl config
```

Output:
```
Current configuration:
  Root:   iac
  Binary: terraform
```

### Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--path` | | Explicit path (mutually exclusive with module name) |
| `--init` | `-i` | Run init before the command (for `fmt` and `val`) |
| `--args` | `-a` | Extra arguments to pass to terraform/tofu (can be specified multiple times) |
| `--version` | `-v` | Show version |
| `--help` | `-h` | Show help |

### Examples

```bash
# Format a component
tfpl fmt storage-account

# Validate a base (with init first)
tfpl val -i k8s-argocd

# Use explicit path
tfpl fmt --path iac/components/azurerm/storage-account

# Init a project
tfpl init spacelift-modules

# Pass extra arguments
tfpl init storage-account -a -upgrade -a -reconfigure

# Show version
tfpl -v
```

## Configuration File

Create a `.tfpl.yml` file in your repository root to configure `tfpl`:

```yaml
# The root directory containing the polylith structure (components, bases, projects)
# Default: "" (repository root)
root: iac

# The Terraform binary to use: "terraform" or "tofu"
# Default: "terraform"
binary: terraform
```

### Configuration Options

| Option | Type | Default | Description |
|--------|------|---------|-------------|
| `root` | string | `""` | Directory containing components/bases/projects (relative to repo root) |
| `binary` | string | `"terraform"` | Binary to use: `"terraform"` or `"tofu"` |

The configuration file is optional. If not present, `tfpl` will use default values (empty root, "terraform" binary).

## Expected Directory Structure

`tfpl` expects a polylith-style repository structure:

```
repository-root/
├── .tfpl.yml              # Configuration file (optional)
└── iac/                   # Root directory (if configured as "root: iac")
    ├── components/
    │   ├── aws/
    │   │   └── s3-bucket/
    │   └── azurerm/
    │       ├── storage-account/
    │       └── naming/
    ├── bases/
    │   ├── azsloth/
    │   └── azsloth-docker-translator/
    └── projects/
        └── spacelift-modules/
```

Each module directory (component, base, or project) should contain at least one `.tf` or `.tf.json` file.

## Edge Cases

### Nested Subfolders

`tfpl` recursively searches for modules in nested subdirectories. For example:

```
iac/components/
├── azurerm/
│   ├── storage-account/
│   └── naming/
└── aws/
    └── s3-bucket/
```

You can refer to modules by name regardless of their nesting:

```bash
tfpl fmt storage-account  # Finds iac/components/azurerm/storage-account
tfpl fmt s3-bucket       # Finds iac/components/aws/s3-bucket
```

### Name Clashes

If multiple modules share the same name in different locations, `tfpl` will detect the clash and provide a helpful error:

```
Error: multiple modules named 'naming' found - name clash detected:
  1. /path/to/repo/iac/components/azurerm/naming
  2. /path/to/repo/iac/components/aws/naming

Please use --path to specify the exact path
```

To resolve this, use the `--path` flag with an explicit path:

```bash
tfpl fmt --path iac/components/azurerm/naming
```

### Module Type Detection

`tfpl` automatically searches for modules across all three directories (components, bases, and projects). You don't need to specify the module type - just provide the name:

```bash
tfpl fmt storage-account  # Searches in components, bases, and projects
tfpl val k8s-argocd      # Automatically finds it in bases/
tfpl init prod-infra     # Automatically finds it in projects/
```

### Mutual Exclusivity

The following combinations are not allowed:

- Cannot use `--path` together with a module name argument

Example error:

```bash
tfpl fmt storage --path iac/components/storage
# Error: --path is mutually exclusive with module name argument
```

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
go build -o tfpl .
```

### Project Structure

```
/
├── cmd/
│   └── root.go           # Main command definitions using cobra
├── internal/
│   ├── config/
│   │   └── config.go     # YAML configuration loading
│   ├── finder/
│   │   └── finder.go     # Module discovery with recursive search
│   └── terraform/
│       └── terraform.go  # Terraform/tofu command execution
├── main.go               # Entry point
├── go.mod                # Go modules file
└── README.md             # This file
```

## License

[Add your license here]

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
