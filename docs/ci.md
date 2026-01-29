# Continuous Integration

motf is designed to work well in CI/CD pipelines. This page covers common patterns for integrating motf with GitHub Actions and other CI systems.

## Key Features

| Feature | Description |
|---------|-------------|
| `--changed` flag | Run commands only on modules that changed |
| `--ref` flag | Specify the base branch for comparison |
| `--names` flag | Output module names for scripting |
| `--json` flag | Machine-readable output |
| Exit codes | Non-zero exit on failure |
| `-a --check` | Formatting check mode (no modifications) |

---

## GitHub Actions

### Basic Workflow

This workflow runs on every push and pull request, checking formatting and validating only the modules that changed:

```yaml
name: Terraform

on:
  push:
    branches: [master]
  pull_request:
    branches: [master]

permissions:
  contents: read

jobs:
  terraform-ci:
    name: Terraform CI
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Required for git diff

      - name: Install Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'
          cache: true

      - name: Set up Terraform
        uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: '1.10.0'

      - name: Install motf
        run: go install github.com/TechnicallyJoe/terraform-motf/cmd/motf@latest

      - name: Check formatting
        run: motf fmt --changed --ref origin/${{ github.base_ref || 'master' }} -a --check

      - name: Validate modules
        run: motf val -i --changed --ref origin/${{ github.base_ref || 'master' }}
```

### Important: fetch-depth

Always use `fetch-depth: 0` when using `--changed`:

```yaml
- uses: actions/checkout@v4
  with:
    fetch-depth: 0  # Fetch all history for accurate git diff
```

Without this, the shallow clone won't have the base branch history needed for comparison.

---

## Complete CI Example

This is the workflow used by the motf project itself:

```yaml
name: Terraform

on:
  push:
    branches: [master]
  pull_request:
    branches: [master]

permissions:
  contents: read

jobs:
  terraform-ci:
    name: Terraform CI
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Install Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.25'
          cache: true

      - name: Set up Terraform
        uses: hashicorp/setup-terraform@v3
        with:
          terraform_version: '1.10.0'

      - name: Install motf
        run: go install github.com/TechnicallyJoe/terraform-motf/cmd/motf@latest

      - name: Check formatting
        run: motf fmt -a --check --changed --ref origin/${{ github.base_ref || 'master' }}

      - name: Validate modules
        run: motf val -i --changed --ref origin/${{ github.base_ref || 'master' }}
```

---

## Advanced Patterns

### Run Custom Tasks in CI

```yaml
- name: Run lint task on changed modules
  run: motf task -t lint --changed --ref origin/${{ github.base_ref || 'master' }}

- name: Run docs task on changed modules
  run: motf task -t docs --changed --ref origin/${{ github.base_ref || 'master' }}

```

### Skip CI When No Modules Changed

```yaml
- name: Check for changes
  id: changes
  run: |
    count=$(motf changed --names | wc -l)
    echo "count=$count" >> $GITHUB_OUTPUT

- name: Validate
  if: steps.changes.outputs.count != '0'
  run: motf val -i --changed
```

---

## Using with OpenTofu

If your project uses OpenTofu instead of Terraform:

```yaml
- name: Set up OpenTofu
  uses: opentofu/setup-opentofu@v1
  with:
    tofu_version: '1.8.0'
```

And configure `.motf.yml`:

```yaml
binary: tofu
```

---

## Exit Codes

motf uses standard exit codes:

| Code | Meaning |
|------|---------|
| 0 | Success (all modules passed) |
| 1 | Failure (at least one module failed) |

When using `--changed`:
- If no changes, exit code is 0
- If any module fails, exit code is 1 (but all modules are attempted)

---

## Tips

1. **Always fetch full history** for `--changed` to work correctly
2. **Use `--ref`** explicitly in CI to avoid auto-detection issues
3. **Combine `-i` with `val`** to ensure modules are initialized before validation
6. **Cache Go modules** when building motf from source
