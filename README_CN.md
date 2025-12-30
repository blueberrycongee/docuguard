# DocuGuard

[![Build Status](https://github.com/blueberrycongee/docuguard/workflows/CI/badge.svg)](https://github.com/blueberrycongee/docuguard/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/blueberrycongee/docuguard)](https://goreportcard.com/report/github.com/blueberrycongee/docuguard)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

[English](README.md) | 中文

轻量级文档-代码一致性检查工具，基于 LLM 语义分析自动检测文档与代码实现之间的冲突。

## 特性

- 自动检测文档描述与代码实现的不一致
- **PR Bot 模式**：自动检查 PR 中可能需要更新的文档
- 支持 Markdown 文件中的绑定注解
- 多种输出格式：文本、JSON、GitHub Actions
- YAML 配置文件
- 支持 CI/CD 集成

## 环境要求

- Go 1.21+

## 安装

### 从源码安装

```bash
go install github.com/blueberrycongee/docuguard/cmd/docuguard@latest
```

### 下载二进制

从 [Releases](https://github.com/blueberrycongee/docuguard/releases) 页面下载。

## 快速开始

DocuGuard 支持两种模式：

### 模式一：注解绑定检查

通过注解手动绑定文档与代码。

#### 1. 初始化配置

```bash
docuguard init
```

#### 2. 添加绑定注解

在 Markdown 文件中添加：

```markdown
<!-- docuguard:start -->
<!-- docuguard:bindCode path="src/payment.go" func="CalculateShipping" -->

订单金额超过 100 元免运费。

<!-- docuguard:end -->
```

#### 3. 运行检查

```bash
docuguard check docs/api.md
```

### 模式二：PR Bot 模式（推荐）

自动检测代码变更并查找相关文档。

#### 本地开发

```bash
# 比较当前分支与 main
docuguard pr

# 指定基准分支
docuguard pr --base main

# 比较最近 3 个提交
docuguard pr --base HEAD~3

# 仅显示检测到的变更（dry run）
docuguard pr --dry-run

# 跳过 LLM，仅使用关键词匹配
docuguard pr --skip-llm
```

#### GitHub CI

```bash
# 检查指定 PR
docuguard pr --github --pr 123

# 在 PR 上发表评论
docuguard pr --github --pr 123 --comment
```

## 配置说明

创建 `.docuguard.yaml`：

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

设置 API Key：

```bash
export OPENAI_API_KEY=your-api-key
```

## 命令说明

### docuguard check

使用注解检查文档-代码一致性。

```bash
docuguard check [files...]
docuguard check --all
docuguard check --format json docs/api.md
```

### docuguard pr

检查 PR 变更的文档一致性。

```bash
# 本地模式
docuguard pr [flags]

参数:
  --base string     基准分支 (默认 "main")
  --docs strings    文档匹配模式 (默认 [README.md,docs/**/*.md])
  --dry-run         仅显示检测到的变更
  --skip-llm        跳过 LLM 检查，仅使用关键词匹配
  --format string   输出格式: text, json (默认 "text")

# GitHub 模式
docuguard pr --github [flags]

参数:
  --pr int          PR 编号 (必需)
  --token string    GitHub Token (或使用 GITHUB_TOKEN 环境变量)
  --repo string     仓库 owner/repo (自动检测)
  --comment         在 PR 上发表评论
```

### docuguard init

初始化配置文件。

```bash
docuguard init
```

## 支持的绑定类型

| 类型 | 语法 |
|------|------|
| 函数 | `func="FunctionName"` |
| 结构体 | `struct="StructName"` |
| 常量 | `const="ConstName"` |
| 变量 | `var="VarName"` |

## 输出格式

- `text` - 可读文本格式（默认）
- `json` - JSON 格式
- `github-actions` - GitHub Actions 注解格式

## CI/CD 集成

### GitHub Actions (PR Bot 模式)

添加到你的工作流 (`.github/workflows/docuguard.yml`)：

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
        run: go install github.com/blueberrycongee/docuguard/cmd/docuguard@latest

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

### GitHub Actions (注解模式)

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
        run: go install github.com/blueberrycongee/docuguard/cmd/docuguard@latest

      - name: Check Documentation
        run: docuguard check --all --format github-actions
        env:
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
```

## PR Bot 工作原理

1. 从 git diff 检测代码变更
2. 提取变更的 Go 符号（函数、结构体等）
3. 扫描文档文件（README.md、docs/*.md）
4. 使用关键词匹配查找相关文档
5. 可选使用 LLM 验证一致性
6. 输出报告或发表 PR 评论

## 参与贡献

请阅读 [CONTRIBUTING.md](CONTRIBUTING.md)。

## 开源协议

[MIT](LICENSE)
