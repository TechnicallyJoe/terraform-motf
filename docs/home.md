# motf - Terraform Monorepo Orchestrator

**motf** (pronounced *motif* or *mo-tif*) is a CLI tool for working with Terraform monorepos. It simplifies running terraform/tofu commands across modules organized in a [polylith-style](https://polylith.gitbook.io/poly/) repository structure.

![motf demo](assets/demo.gif)

## Why motf?

Managing Terraform in a monorepo can be cumbersome. You need to navigate to the right directory, run `terraform init`, then run your command. When working on multiple modules, this becomes tedious.

**motf** solves this by:

- **Finding modules by name** — Run `motf fmt storage-account` from anywhere in your repo
- **Smart discovery** — Recursively searches `components/`, `bases/`, and `projects/`
- **Consistent interface** — Same commands work across all module types
- **Change detection** — Run commands only on modules that changed (`--changed`)
- **CI-friendly** — JSON output, exit codes, and `--names` flag for scripting

## Core Concepts

### Repository Structure

motf expects a **polylith-style** Terraform monorepo with three module categories:

```
repository-root/
├── .motf.yml              # Optional configuration
├── components/            # Reusable Terraform modules
│   └── azurerm/
│       ├── storage-account/
│       ├── key-vault/
│       └── resource-group/
├── bases/                 # Composable base configurations
│   └── k8s-argocd/
└── projects/              # Deployable infrastructure
    └── prod-infra/
```

### Module Types

| Type | Directory | Purpose |
|------|-----------|---------|
| **Component** | `components/` | Reusable, single-purpose modules (e.g., `storage-account`, `vpc`) |
| **Base** | `bases/` | Compositions of components for specific platforms (e.g., `k8s-argocd`) |
| **Project** | `projects/` | Deployable infrastructure that references components and bases |

### Module Discovery

motf recursively searches for modules in nested subdirectories. A directory is recognized as a module if it contains `.tf` or `.tf.json` files.

```bash
# These all work regardless of nesting depth
motf fmt storage-account     # Finds components/azurerm/storage-account/
motf val k8s-argocd          # Finds bases/k8s-argocd/
motf init prod-infra         # Finds projects/prod-infra/
```

**Skipped directories:** `.terraform`, `.git`, `examples`, `modules`, `tests`, `.spacelift`

## Quick Start

### Installation

```bash
# Using go install
go install github.com/TechnicallyJoe/terraform-motf/cmd/motf@latest

# Or build from source
git clone https://github.com/TechnicallyJoe/terraform-motf.git
cd terraform-motf
go build -o motf ./cmd/motf
```

### Requirements

- Go 1.25+ (for building from source)
- `terraform` or `tofu` CLI in PATH

### First Commands

```bash
# List all modules in your repo
motf list

# Get details about a specific module
motf get storage-account

# Format a module
motf fmt storage-account

# Validate with init
motf val -i storage-account

# See what changed
motf changed
```

## Documentation

| Page | Description |
|------|-------------|
| [Commands](commands) | Full reference for all CLI commands |
| [Configuration](configuration) | `.motf.yml` options and custom tasks |
| [CI Integration](ci) | GitHub Actions and pipeline examples |

## Features at a Glance

| Feature | Description |
|---------|-------------|
| **Simple commands** | `init`, `fmt`, `validate`, `plan`, `test` on any module |
| **Module inspection** | `get` and `describe` for detailed module info |
| **Example targeting** | Run commands on `examples/` subdirectories with `-e` |
| **Change detection** | `--changed` flag to run only on modified modules |
| **Custom tasks** | Define shell commands in `.motf.yml` |
| **Multiple binaries** | Support for both `terraform` and `tofu` |
| **JSON output** | `--json` flag for scripting and CI |
| **Name clash detection** | Clear errors when module names conflict |

## Getting Help

```bash
# General help
motf --help

# Command-specific help
motf fmt --help
motf changed --help

# Show version
motf --version
```