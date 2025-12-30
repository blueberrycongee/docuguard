# DocuGuard

[![Build Status](https://github.com/yourname/docuguard/workflows/CI/badge.svg)](https://github.com/yourname/docuguard/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/yourname/docuguard)](https://goreportcard.com/report/github.com/yourname/docuguard)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

English | [中文](README_CN.md)

A lightweight CLI tool for checking documentation-code consistency using LLM semantic analysis.

## Features

- Detects inconsistencies between documentation and code implementation
- **PR Bot Mode**: Automatically checks PRs for documentation that may need updates
- Supports binding annotations in Markdown files
- Multiple output formats: text, JSON, GitHub Actions
- Configurable via YAML
- CI/CD integration ready

## Requirements

- Go 1.21+

## Installation

### From Source

```bash
go install github.com/yourname/docuguard/cmd/docuguard@latest
```

### Binary

Download from [Releases](https://github.com/yourname/docuguard/releases).

## Quick Start

DocuGuard supports two modes:

### Mode 1: Annotation-Based Check

Manually bind documentation to code using annotations.

#### 1. Initialize

```bash
docuguard init
```

#### 2. Add Binding

In your Markdown file:

```markdown
<!-- docuguard:start -->
<!-- docuguard:bindCode path="src/payment.go" func="CalculateShipping" -->

Orders over $100 qualify for free shipping.

<!-- docuguard:end -->
```

#### 3. Run Check

```bash
docuguard check docs/api.md
```

### Mode 2: PR Bot Mode (Recommended)

Automatically detect code changes and find related documentation.

#### Local Development

```bash
# Compare current branch vs main
docuguard pr

# Specify base branch
docuguard pr --base main

# Compare last 3 commits
docuguard pr --base HEAD~3

# Only show detected changes (dry run)
docuguard pr --dry-run

# Skip LLM, use keyword matching only
docuguard pr --skip-llm
```

#### GitHub CI

```bash
# Check specific PR
docuguard pr --github --pr 123

# Post comment on PR
docuguard pr --github --pr 123 --comment
```

## Configuration

Create `.docuguard.yaml`:

```yaml
version: "1.0"

llm:
  provider: "openai"
  model: "gpt-4"
  timeout: "60s"

scan:
  include:
    - "docs/**/*.md"
  exclude: []

rules:
  fail_on_inconsistent: true
  confidence_threshold: 0.8

output:
  format: "text"
  color: true
```

Set your API key:

```bash
export OPENAI_API_KEY=your-api-key
```

## Commands

### docuguard check

Check documentation-code consistency using annotations.

```bash
docuguard check [files...]
docuguard check --all
docuguard check --format json docs/api.md
```

### docuguard pr

Check documentation consistency for PR changes.

```bash
# Local mode
docuguard pr [flags]

Flags:
  --base string     Base branch for comparison (default "main")
  --docs strings    Documentation patterns to scan (default [README.md,docs/**/*.md])
  --dry-run         Only show detected changes
  --skip-llm        Skip LLM check, use keyword matching only
  --format string   Output format: text, json (default "text")

# GitHub mode
docuguard pr --github [flags]

Flags:
  --pr int          PR number (required)
  --token string    GitHub token (or use GITHUB_TOKEN env)
  --repo string     Repository owner/repo (auto-detected)
  --comment         Post comment on PR
```

### docuguard init

Initialize configuration file.

```bash
docuguard init
```

## Supported Bindings

| Type | Syntax |
|------|--------|
| Function | `func="FunctionName"` |
| Struct | `struct="StructName"` |
| Const | `const="ConstName"` |
| Var | `var="VarName"` |

## Output Formats

- `text` - Human-readable output (default)
- `json` - Machine-readable JSON
- `github-actions` - GitHub Actions annotations

## CI/CD Integration

### GitHub Actions (PR Bot)

Add to your workflow (`.github/workflows/docuguard.yml`):

```yaml
name: Documentation Check

on:
  pull_request:
    types: [opened, synchronize]

jobs:
  docuguard:
    runs-on: ubuntu-latest
    permissions:
      contents: read
      pull-requests: write
    
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.21"

      - name: Install DocuGuard
        run: go install github.com/yourname/docuguard/cmd/docuguard@latest

      - name: Run DocuGuard
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
        run: |
          docuguard pr \
            --github \
            --pr ${{ github.event.pull_request.number }} \
            --comment
```

### GitHub Actions (Annotation Mode)

```yaml
name: Documentation Check

on:
  push:
    paths:
      - "docs/**"
      - "**.go"

jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: "1.21"

      - name: Install DocuGuard
        run: go install github.com/yourname/docuguard/cmd/docuguard@latest

      - name: Check Documentation
        run: docuguard check --all --format github-actions
        env:
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
```

## How PR Bot Works

1. Detects code changes from git diff
2. Extracts changed Go symbols (functions, structs, etc.)
3. Scans documentation files (README.md, docs/*.md)
4. Uses keyword matching to find related documentation
5. Optionally uses LLM to verify consistency
6. Reports findings or posts PR comment

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT
