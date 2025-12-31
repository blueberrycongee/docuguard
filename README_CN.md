# DocuGuard

<p align="center">
  <img src="https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go" alt="Go Version">
  <a href="https://github.com/blueberrycongee/docuguard/actions"><img src="https://github.com/blueberrycongee/docuguard/actions/workflows/ci.yml/badge.svg?branch=main" alt="Build Status"></a>
  <a href="https://goreportcard.com/report/github.com/blueberrycongee/docuguard"><img src="https://goreportcard.com/badge/github.com/blueberrycongee/docuguard" alt="Go Report Card"></a>
  <a href="https://pkg.go.dev/github.com/blueberrycongee/docuguard"><img src="https://pkg.go.dev/badge/github.com/blueberrycongee/docuguard.svg" alt="Go Reference"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="License"></a>
</p>

<p align="center">
  <a href="README.md">English</a> | <b>中文</b>
</p>

轻量级文档-代码一致性检查工具，基于 LLM 语义分析自动检测文档与代码实现之间的冲突。

## 特性

- **智能检测** - 自动检测文档描述与代码实现的不一致
- **PR Bot 模式** - 自动检查 PR 中可能需要更新的文档
- **两阶段匹配** - 宽松关键词匹配 + LLM 相关性过滤，提高准确率
- **注解支持** - 支持 Markdown 文件中的绑定注解
- **多种格式** - 支持文本、JSON、GitHub Actions 输出格式
- **灵活配置** - YAML 配置文件
- **CI/CD 就绪** - 轻松集成 GitHub Actions

## 安装

### 从源码安装

```bash
go install github.com/blueberrycongee/docuguard/cmd/docuguard@latest
```

### 下载二进制

从 [Releases](https://github.com/blueberrycongee/docuguard/releases) 页面下载。

### 环境要求

- Go 1.21+
- OpenAI API Key（或兼容的 API）

## 快速开始

### PR Bot 模式（推荐）

自动检测代码变更并查找相关文档。

```bash
# 比较当前分支与 main
docuguard pr

# 使用两阶段匹配提高准确率
docuguard pr --two-stage

# 指定基准分支
docuguard pr --base develop

# 仅显示检测到的变更（dry run）
docuguard pr --dry-run
```

### 注解绑定检查

通过注解手动绑定文档与代码。

```markdown
<!-- docuguard:start -->
<!-- docuguard:bindCode path="src/payment.go" func="CalculateShipping" -->

订单金额超过 100 元免运费。

<!-- docuguard:end -->
```

```bash
docuguard check docs/api.md
```

## 配置说明

在项目根目录创建 `.docuguard.yaml`：

```yaml
version: "1.0"

llm:
  provider: "openai"        # openai, anthropic, ollama
  model: "gpt-4"
  base_url: ""              # 可选：自定义 API 端点
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

设置 API Key：

```bash
export OPENAI_API_KEY=your-api-key

# 或使用自定义端点（如 Azure OpenAI、SiliconFlow）
export OPENAI_API_BASE=https://your-api-endpoint
```

## 命令说明

### `docuguard pr`

检查 PR/代码变更的文档一致性。

```bash
docuguard pr [flags]

参数:
  --base string       基准分支 (默认 "main")
  --docs strings      文档匹配模式 (默认 [README.md,docs/**/*.md])
  --dry-run           仅显示检测到的变更，跳过 LLM 检查
  --skip-llm          跳过 LLM，仅使用关键词匹配
  --two-stage         使用两阶段匹配（宽松匹配 + LLM 过滤）
  --format string     输出格式: text, json (默认 "text")

GitHub 模式:
  --github            启用 GitHub 模式
  --pr int            PR 编号 (GitHub 模式必需)
  --token string      GitHub Token (或使用 GITHUB_TOKEN 环境变量)
  --repo string       仓库 owner/repo (自动检测)
  --comment           在 PR 上发表评论
```

### `docuguard check`

使用绑定注解检查文档。

```bash
docuguard check [files...]
docuguard check --all
docuguard check --format json docs/api.md
```

### `docuguard init`

初始化配置文件。

```bash
docuguard init
```

## 工作原理

### 两阶段匹配（推荐）

```
代码变更 → 提取符号 → [阶段1: 宽松匹配] → 候选文档 → [阶段2: LLM 过滤] → 相关文档 → [阶段3: 一致性检查]
```

1. **阶段 1 - 宽松匹配**：使用多种策略（精确名称、代码块、关键词、部分匹配）查找候选文档
2. **阶段 2 - LLM 过滤**：批量调用 LLM 过滤出真正相关的文档
3. **阶段 3 - 一致性检查**：验证文档是否与代码实现一致

### 支持的绑定类型

| 类型 | 语法 |
|------|------|
| 函数 | `func="FunctionName"` |
| 结构体 | `struct="StructName"` |
| 常量 | `const="ConstName"` |
| 变量 | `var="VarName"` |

## CI/CD 集成

### GitHub Actions (PR Bot 模式)

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

### GitHub Actions (注解模式)

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

## 参与贡献

欢迎贡献代码！请在提交 Pull Request 之前阅读我们的 [贡献指南](CONTRIBUTING.md)。

1. Fork 本仓库
2. 创建特性分支 (`git checkout -b feature/amazing-feature`)
3. 提交更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 提交 Pull Request

## 开源协议

本项目采用 MIT 协议 - 详见 [LICENSE](LICENSE) 文件。

## 致谢

- [spf13/cobra](https://github.com/spf13/cobra) - CLI 框架
- [spf13/viper](https://github.com/spf13/viper) - 配置管理
- OpenAI - LLM API
