# DocuGuard

[![Build Status](https://github.com/yourname/docuguard/workflows/CI/badge.svg)](https://github.com/yourname/docuguard/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/yourname/docuguard)](https://goreportcard.com/report/github.com/yourname/docuguard)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

English | [中文](README_CN.md)

A lightweight CLI tool for checking documentation-code consistency using LLM semantic analysis.

## Features

- Detects inconsistencies between documentation and code implementation
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

### 1. Initialize

```bash
docuguard init
```

### 2. Add Binding

In your Markdown file:

```markdown
<!-- docuguard:start -->
<!-- docuguard:bindCode path="src/payment.go" func="CalculateShipping" -->

Orders over $100 qualify for free shipping.

<!-- docuguard:end -->
```

### 3. Run Check

```bash
docuguard check docs/api.md
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

### GitHub Actions

Add to your workflow (`.github/workflows/docs.yml`):

```yaml
name: Documentation Check

on:
  push:
    paths:
      - "docs/**"
      - "**.go"
  pull_request:
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

Or download binary directly:

```yaml
- name: Install DocuGuard
  run: |
    curl -sL https://github.com/yourname/docuguard/releases/latest/download/docuguard-linux-amd64 -o docuguard
    chmod +x docuguard
    sudo mv docuguard /usr/local/bin/
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## License

MIT
