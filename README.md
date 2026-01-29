# motf - Terraform Monorepo Orchestrator

A command-line tool for working with Terraform monorepos. `motf` (pronounced *motif*) makes it easy to run terraform/tofu commands on components, bases, and projects organized in a monorepo structure.

![motf demo](docs/assets/demo.gif)

## Features

- **Simple commands**: Run `init`, `fmt`, `validate`, `plan`, and `test` on any module by name
- **Smart discovery**: Recursively finds modules in nested subdirectories
- **Change detection**: Run commands only on modified modules with `--changed`
- **Module inspection**: View detailed module info with `get` and `describe`
- **Custom tasks**: Define shell commands in `.motf.yml`
- **CI-friendly**: JSON output, exit codes, and scripting support

## Installation

```bash
# Using go install
go install github.com/TechnicallyJoe/terraform-motf/cmd/motf@latest
```

## Building

```bash
# Or build from source
git clone https://github.com/TechnicallyJoe/terraform-motf.git
cd terraform-motf
go build -o motf ./cmd/motf
```

## Requirements

- Go 1.25+ (for building)
- `terraform` or `tofu` CLI in PATH

## Quick Start

```bash
$ motf help

motf (Terraform Monorepo Orchestrator) is a CLI tool for working with Terraform monorepos.

It supports running terraform/tofu commands on components, bases, and projects organized
in a structured monorepo.

Usage:
  motf [command]

Examples:
  motf fmt storage-account         # Run fmt on storage-account (searches all types)
  motf val k8s-argocd              # Run validate on k8s-argocd
  motf val -i k8s-argocd           # Run init then validate on k8s-argocd
  motf init k8s-argocd             # Run init on k8s-argocd
  motf fmt --path iac/components/azurerm/storage-account  # Run fmt on explicit path
  motf init storage-account -a -upgrade -a -reconfigure  # Run init with extra args

Available Commands:
  changed     List modules with changes compared to a base branch
  completion  Generate the autocompletion script for the specified shell
  config      Show current configuration
  describe    Describe the interface of a Terraform module
  fmt         Run terraform/tofu fmt on a component, base, or project
  get         Get details about a component, base, or project
  help        Help about any command
  init        Run terraform/tofu init on a component, base, or project
  list        List all modules (components, bases, and projects)
  plan        Run terraform/tofu plan on a component, base, or project
  task        Run a custom task from .motf.yml
  test        Run tests on a component, base, or project
  val         Run terraform/tofu validate on a component, base, or project
  version     Print version information

Flags:
  -a, --args stringArray   Extra arguments to pass to terraform/tofu (can be specified multiple times)
  -c, --config string      Path to config file (default: searches for .motf.yml)
  -h, --help               help for motf
      --path string        Explicit path (mutually exclusive with module name)
  -v, --version            version for motf

Use "motf [command] --help" for more information about a command.

```

## Documentation

For comprehensive documentation, see the **[Wiki](https://github.com/TechnicallyJoe/terraform-motf/wiki)**:

## Repository Structure

motf expects a polylith-style monorepo for types to work:

```
repository-root/
├── .motf.yml              # Optional configuration
├── components/            # Reusable Terraform modules
├── bases/                 # Composable base configurations
└── projects/              # Deployable infrastructure
```

Create `.motf.yml` in your repository root. This can be used to customize settings like module root, binary choice (terraform vs tofu), and define custom tasks.

See [Configuration](https://github.com/TechnicallyJoe/terraform-motf/wiki/configuration) for all options.

## Releases

Download the latest release from the [Releases page](https://github.com/TechnicallyJoe/terraform-motf/releases).

Releases are automated via [GoReleaser](https://goreleaser.com/) when a version tag is pushed:

```bash
git tag v1.0.0
git push origin v1.0.0
```

## Contributing

Contributions are welcome! This project uses [Conventional Commits](https://www.conventionalcommits.org/):

### Types
| Type | Description | Example |
|------|-------------|---------|
| `feat` | New feature | `feat(tasks): add shell configuration support` |
| `fix` | Bug fix | `fix(finder): handle symlinks correctly` |
| `docs` | Documentation | `docs: update README with examples` |
| `chore` | Maintenance | `chore(deps): update cobra to v1.9.0` |
| `refactor` | Code refactoring | `refactor(cmd): extract helper functions` |
| `test` | Tests | `test(e2e): add plan command tests` |

## License

MIT License - see [LICENSE](LICENSE) for details.
