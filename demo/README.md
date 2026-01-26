# Demo Polylith Repository

This directory contains a demo polylith-style Terraform repository structure for testing and demonstration purposes.

## Structure

```
demo/
├── .motf.yml              # motf configuration
├── components/            # Reusable terraform components
│   └── azurerm/
│       ├── storage-account/
│       └── key-vault/
├── bases/                 # Composable base configurations
│   └── k8s-argocd/
└── projects/              # Deployable projects
    └── prod-infra/
```

## Usage

From the `demo` directory, you can run motf commands:

```bash
cd demo

# Format a component
motf fmt -c storage-account

# Initialize a base
motf init -b k8s-argocd

# Validate a project (with init)
motf val -i -p prod-infra

# Use explicit path
motf fmt --path components/azurerm/key-vault

# Pass extra arguments to terraform
motf init -c storage-account -a -upgrade
```

## Components

- **storage-account**: Azure Storage Account configuration
- **key-vault**: Azure Key Vault configuration

## Bases

- **k8s-argocd**: Kubernetes ArgoCD deployment base

## Projects

- **prod-infra**: Production infrastructure project
