<div align="center">

<img src=".github/assets/image.png" alt="DocuGuard" width="800">

# DocuGuard

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Build Status](https://github.com/blueberrycongee/docuguard/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/blueberrycongee/docuguard/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/blueberrycongee/docuguard)](https://goreportcard.com/report/github.com/blueberrycongee/docuguard)
[![Go Reference](https://pkg.go.dev/badge/github.com/blueberrycongee/docuguard.svg)](https://pkg.go.dev/github.com/blueberrycongee/docuguard)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

**English** | [中文](README_CN.md)

</div>

---

A lightweight CLI tool for checking documentation-code consistency using LLM semantic analysis.

## Features

- **Smart Detection** - Detects inconsistencies between documentation and code implementation
- **PR Bot Mode** - Automatically checks PRs for documentation that may need updates
- **Two-Stage Matching** - Broad keyword matching + LLM relevance filtering for accurate results
- **Annotation Support** - Supports binding annotations in Markdown files
- **Multiple Formats** - Output in text, JSON, or GitHub Actions format
- **Configurable** - Flexible YAML configuration
- **CI/CD Ready** - Easy integration with GitHub Actions

## Installation

### From Source

```bash
go install github.com/blueberrycongee/docuguard/cmd/docuguard@latest
```

### Binary

Download from [Releases](https://github.com/blueberrycongee/docuguard/releases).

### Requirements

- Go 1.21+
- OpenAI API key (or compatible API)

## Quick Start

### PR Bot Mode (Recommended)

Automatically detect code changes and find related documentation.

```bash
# Compare current branch vs main
docuguard pr

# Use two-stage matching for better accuracy
docuguard pr --two-stage

# Specify base branch
docuguard pr --base develop

# Only show detected changes (dry run)
docuguard pr --dry-run
```

### Annotation-Based Check

Manually bind documentation to code using annotations.

```markdown
<!-- docuguard:start -->
<!-- docuguard:bindCode path="src/payment.go" func="CalculateShipping" -->

Orders over $100 qualify for free shipping.

<!-- docuguard:end -->
```

```bash
docuguard check docs/api.md
```

## Configuration

Create `.docuguard.yaml` in your project root:

```yaml
version: "1.0"

llm:
  provider: "openai"        # openai, anthropic, ollama
  model: "gpt-4"
  base_url: ""              # Optional: custom API endpoint
  timeout: "60s"

scan:
  include:
    - "README.md"
    - "docs/**/*.md"
  exclude: []

rules:
  fail_on_inconsistent: true
  confidence_threshold: 0.8

output:
  format: "text"            # text, json, github-actions
  color: true
```

Set your API key:

```bash
export OPENAI_API_KEY=your-api-key

# Or use custom endpoint (e.g., Azure OpenAI, SiliconFlow)
export OPENAI_API_BASE=https://your-api-endpoint
```

## Commands

### `docuguard pr`

Check documentation consistency for PR/code changes.

```bash
docuguard pr [flags]

Flags:
  --base string       Base branch for comparison (default "main")
  --docs strings      Documentation patterns (default [README.md,docs/**/*.md])
  --dry-run           Only show detected changes, skip LLM check
  --skip-llm          Skip LLM, use keyword matching only
  --two-stage         Use two-stage matching (broad match + LLM filter)
  --format string     Output format: text, json (default "text")

GitHub Mode:
  --github            Enable GitHub mode
  --pr int            PR number (required in GitHub mode)
  --token string      GitHub token (or use GITHUB_TOKEN env)
  --repo string       Repository owner/repo (auto-detected)
  --comment           Post comment on PR
```

### `docuguard check`

Check documentation using binding annotations.

```bash
docuguard check [files...]
docuguard check --all
docuguard check --format json docs/api.md
```

### `docuguard init`

Initialize configuration file.

```bash
docuguard init
```

## How It Works

### PR Bot Mode

```
PR Diff → Parse Changed Lines → Extract Symbols & Code → Match Documents → LLM Consistency Check
```

1. **Get PR Diff**: Fetches diff from GitHub API (or local git diff)
2. **Extract Symbols**: Parses diff lines to extract changed symbol names and their code directly from the diff (not from local files)
3. **Match Documents**: Finds related documentation using keyword matching
4. **LLM Check**: Verifies if documentation matches the new code implementation

### Two-Stage Matching (with `--two-stage` flag)

For better accuracy, use the two-stage matching mode:

```
Symbols → [Stage 1: Broad Match] → Candidates → [Stage 2: LLM Filter] → Relevant Docs → Consistency Check
```

1. **Stage 1 - Broad Match**: Uses multiple strategies (exact name, code blocks, keywords, partial match) to find candidate documents
2. **Stage 2 - LLM Filter**: Batch checks candidates with LLM to filter truly relevant documents
3. **Consistency Check**: Verifies if documentation matches code implementation

### Supported Bindings

| Type | Syntax |
|------|--------|
| Function | `func="FunctionName"` |
| Struct | `struct="StructName"` |
| Const | `const="ConstName"` |
| Var | `var="VarName"` |

## CI/CD Integration

### GitHub Actions (PR Bot)

```yaml
name: DocuGuard

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

      - uses: actions/setup-go@v5
        with:
          go-version: "1.21"

      - name: Install DocuGuard
        run: go install github.com/blueberrycongee/docuguard/cmd/docuguard@latest

      - name: Check Documentation
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
        run: |
          docuguard pr \
            --github \
            --pr ${{ github.event.pull_request.number }} \
            --two-stage \
            --comment
```

### GitHub Actions (Annotation Mode)

```yaml
name: DocuGuard Check

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

      - uses: actions/setup-go@v5
        with:
          go-version: "1.21"

      - name: Install DocuGuard
        run: go install github.com/blueberrycongee/docuguard/cmd/docuguard@latest

      - name: Check Documentation
        env:
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
        run: docuguard check --all --format github-actions
```

## Contributing

Contributions are welcome! Please read our [Contributing Guide](CONTRIBUTING.md) before submitting a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [spf13/cobra](https://github.com/spf13/cobra) - CLI framework
- [spf13/viper](https://github.com/spf13/viper) - Configuration management
- OpenAI - LLM API
