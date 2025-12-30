# DocuGuard

[![Build Status](https://github.com/yourname/docuguard/workflows/CI/badge.svg)](https://github.com/yourname/docuguard/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/yourname/docuguard)](https://goreportcard.com/report/github.com/yourname/docuguard)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

[English](README.md) | 中文

轻量级文档-代码一致性检查工具，基于 LLM 语义分析自动检测文档与代码实现之间的冲突。

## 特性

- 自动检测文档描述与代码实现的不一致
- 支持 Markdown 文件中的绑定注解
- 多种输出格式：文本、JSON、GitHub Actions
- YAML 配置文件
- 支持 CI/CD 集成

## 环境要求

- Go 1.21+

## 安装

### 从源码安装

```bash
go install github.com/yourname/docuguard/cmd/docuguard@latest
```

### 下载二进制

从 [Releases](https://github.com/yourname/docuguard/releases) 页面下载。

## 快速开始

### 1. 初始化配置

```bash
docuguard init
```

### 2. 添加绑定注解

在 Markdown 文件中添加：

```markdown
<!-- docuguard:start -->
<!-- docuguard:bindCode path="src/payment.go" func="CalculateShipping" -->

订单金额超过 100 元免运费。

<!-- docuguard:end -->
```

### 3. 运行检查

```bash
docuguard check docs/api.md
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

### GitHub Actions

添加到你的工作流 (`.github/workflows/docs.yml`)：

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

或直接下载二进制：

```yaml
- name: Install DocuGuard
  run: |
    curl -sL https://github.com/yourname/docuguard/releases/latest/download/docuguard-linux-amd64 -o docuguard
    chmod +x docuguard
    sudo mv docuguard /usr/local/bin/
```

## 参与贡献

请阅读 [CONTRIBUTING.md](CONTRIBUTING.md)。

## 开源协议

[MIT](LICENSE)
