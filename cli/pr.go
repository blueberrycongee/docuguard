package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourname/docuguard/internal/config"
	"github.com/yourname/docuguard/internal/engine"
	"github.com/yourname/docuguard/internal/git"
	"github.com/yourname/docuguard/internal/github"
	"github.com/yourname/docuguard/internal/matcher"
	"github.com/yourname/docuguard/internal/reporter"
	"github.com/yourname/docuguard/internal/scanner"
	"github.com/yourname/docuguard/pkg/types"
)

var (
	prBaseBranch string
	prDryRun     bool
	prFormat     string
	prDocs       []string
	prSkipLLM    bool
	// GitHub 模式参数
	prGitHub  bool
	prNumber  int
	prToken   string
	prRepo    string
	prComment bool
)

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "Check documentation consistency for PR changes",
	Long: `Analyze code changes in a PR and check if related documentation needs updates.

Local mode (development):
  docuguard pr                    # Compare current branch vs main
  docuguard pr --base main        # Specify base branch
  docuguard pr --base HEAD~3      # Compare last 3 commits
  docuguard pr --dry-run          # Only show detected changes

GitHub mode (CI):
  docuguard pr --github --pr 123  # Check specific PR
  docuguard pr --github --comment # Post comment on PR`,
	RunE: runPR,
}

func init() {
	// 本地模式参数
	prCmd.Flags().StringVar(&prBaseBranch, "base", "main", "base branch for comparison")
	prCmd.Flags().BoolVar(&prDryRun, "dry-run", false, "only show detected changes, skip consistency check")
	prCmd.Flags().StringVar(&prFormat, "format", "text", "output format (text|json)")
	prCmd.Flags().StringSliceVar(&prDocs, "docs", []string{"README.md", "docs/**/*.md"}, "documentation patterns to scan")
	prCmd.Flags().BoolVar(&prSkipLLM, "skip-llm", false, "skip LLM check, use keyword matching only")

	// GitHub 模式参数
	prCmd.Flags().BoolVar(&prGitHub, "github", false, "enable GitHub mode")
	prCmd.Flags().IntVar(&prNumber, "pr", 0, "PR number (required in GitHub mode)")
	prCmd.Flags().StringVar(&prToken, "token", "", "GitHub token (or use GITHUB_TOKEN env)")
	prCmd.Flags().StringVar(&prRepo, "repo", "", "repository owner/repo (auto-detected from git remote)")
	prCmd.Flags().BoolVar(&prComment, "comment", false, "post comment on PR")

	rootCmd.AddCommand(prCmd)
}

func runPR(cmd *cobra.Command, args []string) error {
	// 检查是否在 git 仓库中
	if !git.IsInGitRepo() {
		return fmt.Errorf("not in a git repository")
	}

	if prGitHub {
		return runPRGitHub()
	}
	return runPRLocal()
}

// runPRLocal 本地模式：比较当前分支与基准分支
func runPRLocal() error {
	fmt.Printf("🔍 Analyzing changes from %s...\n\n", prBaseBranch)

	// 获取 diff
	diff, err := git.GetDiff(prBaseBranch)
	if err != nil {
		// 尝试获取未提交的变更
		diff, err = git.GetDiffUncommitted()
		if err != nil {
			return fmt.Errorf("failed to get diff: %w", err)
		}
		if diff != "" {
			fmt.Println("📋 Checking uncommitted changes...\n")
		}
	}

	if diff == "" {
		fmt.Println("✅ No changes detected")
		return nil
	}

	// 提取变更符号
	extractor := git.NewSymbolExtractor()
	symbols, err := extractor.ExtractChangedSymbols(diff)
	if err != nil {
		return fmt.Errorf("failed to extract symbols: %w", err)
	}

	if len(symbols) == 0 {
		fmt.Println("✅ No Go symbol changes detected")
		return nil
	}

	// dry-run 模式：只输出变更符号
	if prDryRun {
		if prFormat == "json" {
			return outputSymbolsJSON(symbols)
		}
		outputSymbolsText(symbols)
		fmt.Println("\n💡 Use without --dry-run to check documentation consistency")
		return nil
	}

	// 扫描文档
	fmt.Printf("📄 Scanning documentation...\n")
	segments, err := scanner.ScanMarkdownDir(".", prDocs)
	if err != nil {
		return fmt.Errorf("failed to scan documents: %w", err)
	}

	if len(segments) == 0 {
		fmt.Println("⚠️  No documentation found matching patterns")
		return nil
	}
	fmt.Printf("   Found %d document segments\n\n", len(segments))

	// 查找相关文档
	fmt.Printf("🔗 Finding relevant documentation...\n")
	relevantPairs := matcher.QuickMatch(symbols, segments)
	fmt.Printf("   Found %d potential matches\n\n", len(relevantPairs))

	if len(relevantPairs) == 0 {
		fmt.Println("✅ No documentation appears to be affected by these changes")
		return nil
	}

	// 如果不跳过 LLM，尝试加载配置并进行一致性检查
	if !prSkipLLM {
		cfg, err := config.Load(cfgFile)
		if err == nil && cfg.LLM.APIKey != "" {
			return runPRWithLLM(cfg, diff)
		}
		fmt.Println("⚠️  No LLM configured, using keyword matching only")
	}

	// 输出结果
	if prFormat == "json" {
		return outputRelevanceJSON(relevantPairs)
	}
	outputRelevanceText(relevantPairs)

	return nil
}

// runPRWithLLM 使用 LLM 进行完整检查
func runPRWithLLM(cfg *config.Config, diff string) error {
	ctx := context.Background()

	prEngine, err := engine.NewPREngine(cfg)
	if err != nil {
		return fmt.Errorf("failed to create PR engine: %w", err)
	}

	opts := engine.PRCheckOptions{
		BaseBranch:  prBaseBranch,
		DocPatterns: prDocs,
		SkipLLM:     prSkipLLM,
	}

	report, err := prEngine.CheckFromDiff(ctx, diff, opts)
	if err != nil {
		return fmt.Errorf("failed to check: %w", err)
	}

	if prFormat == "json" {
		return outputReportJSON(report)
	}
	outputReportText(report)

	return nil
}

// runPRGitHub GitHub 模式：检查指定 PR
func runPRGitHub() error {
	if prNumber == 0 {
		return fmt.Errorf("--pr flag is required in GitHub mode")
	}

	// 获取 token
	token := prToken
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token == "" {
		return fmt.Errorf("GitHub token required: use --token or set GITHUB_TOKEN env")
	}

	fmt.Printf("🔍 Checking PR #%d...\n\n", prNumber)

	// 创建 GitHub 客户端
	ghClient, err := github.NewClient(token, prRepo)
	if err != nil {
		return fmt.Errorf("failed to create GitHub client: %w", err)
	}

	// 获取 PR 信息
	prInfo, err := ghClient.GetPRInfo(prNumber)
	if err != nil {
		return fmt.Errorf("failed to get PR info: %w", err)
	}
	fmt.Printf("📋 PR: %s\n", prInfo.Title)
	fmt.Printf("   Base: %s ← Head: %s\n\n", prInfo.BaseBranch, prInfo.HeadBranch)

	// 获取 PR 文件变更
	files, err := ghClient.GetPRFiles(prNumber)
	if err != nil {
		return fmt.Errorf("failed to get PR files: %w", err)
	}

	// 构建 diff
	diff := github.BuildDiffFromFiles(files)
	if diff == "" {
		fmt.Println("✅ No changes detected")
		return nil
	}

	// 提取变更符号
	extractor := git.NewSymbolExtractor()
	symbols, err := extractor.ExtractChangedSymbols(diff)
	if err != nil {
		return fmt.Errorf("failed to extract symbols: %w", err)
	}

	if len(symbols) == 0 {
		fmt.Println("✅ No Go symbol changes detected")
		return nil
	}

	fmt.Printf("📝 Found %d changed symbol(s)\n", len(symbols))

	// 扫描文档
	segments, err := scanner.ScanMarkdownDir(".", prDocs)
	if err != nil {
		return fmt.Errorf("failed to scan documents: %w", err)
	}
	fmt.Printf("📄 Found %d document segments\n", len(segments))

	// 查找相关文档
	relevantPairs := matcher.QuickMatch(symbols, segments)
	fmt.Printf("🔗 Found %d potential matches\n\n", len(relevantPairs))

	// 构建报告
	report := &types.PRReport{
		TotalSymbols:  len(symbols),
		TotalSegments: len(segments),
		RelevantPairs: len(relevantPairs),
	}

	for _, pair := range relevantPairs {
		report.Results = append(report.Results, types.PRCheckResult{
			Segment:    pair.Segment,
			Symbol:     pair.Symbol,
			Consistent: true, // 默认一致（没有 LLM 检查）
			Confidence: pair.Confidence,
			Reason:     pair.Reason,
		})
	}

	// 发表评论
	if prComment {
		repoURL := fmt.Sprintf("https://github.com/%s/%s", ghClient.GetOwner(), ghClient.GetRepo())
		commentBody := reporter.FormatPRComment(report, repoURL)

		// 查找已存在的评论
		existingID, _ := ghClient.FindExistingComment(prNumber)
		if existingID > 0 {
			fmt.Printf("📝 Updating existing comment...\n")
			if err := ghClient.UpdateComment(existingID, commentBody); err != nil {
				return fmt.Errorf("failed to update comment: %w", err)
			}
		} else {
			fmt.Printf("📝 Creating comment...\n")
			if err := ghClient.CreateComment(prNumber, commentBody); err != nil {
				return fmt.Errorf("failed to create comment: %w", err)
			}
		}
		fmt.Println("✅ Comment posted successfully")
	}

	// 输出结果
	if prFormat == "json" {
		return outputReportJSON(report)
	}
	outputReportText(report)

	return nil
}

func outputSymbolsText(symbols []types.ChangedSymbol) {
	fmt.Printf("📝 Found %d changed symbol(s):\n\n", len(symbols))

	for i, sym := range symbols {
		icon := getChangeIcon(sym.ChangeType)
		fmt.Printf("%d. %s %s %s (%s)\n", i+1, icon, sym.Type, sym.Name, sym.File)
		fmt.Printf("   Lines: %d-%d\n", sym.StartLine, sym.EndLine)

		if sym.NewCode != "" && len(sym.NewCode) < 200 {
			fmt.Printf("   Code:\n")
			for _, line := range splitLines(sym.NewCode, 3) {
				fmt.Printf("      %s\n", line)
			}
		}
		fmt.Println()
	}
}

func outputSymbolsJSON(symbols []types.ChangedSymbol) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(map[string]interface{}{
		"symbols": symbols,
		"count":   len(symbols),
	})
}

func outputRelevanceText(pairs []types.RelevanceResult) {
	fmt.Printf("📋 Documentation that may need review:\n\n")

	for i, pair := range pairs {
		fmt.Printf("%d. %s ↔ %s\n", i+1, pair.Segment.Heading, pair.Symbol.Name)
		fmt.Printf("   Doc: %s (L%d-%d)\n", pair.Segment.File, pair.Segment.StartLine, pair.Segment.EndLine)
		fmt.Printf("   Code: %s (L%d-%d)\n", pair.Symbol.File, pair.Symbol.StartLine, pair.Symbol.EndLine)
		fmt.Printf("   Confidence: %.0f%%\n", pair.Confidence*100)
		fmt.Println()
	}
}

func outputRelevanceJSON(pairs []types.RelevanceResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(map[string]interface{}{
		"matches": pairs,
		"count":   len(pairs),
	})
}

func outputReportText(report *types.PRReport) {
	fmt.Printf("\n📊 Summary:\n")
	fmt.Printf("   Symbols changed: %d\n", report.TotalSymbols)
	fmt.Printf("   Documents scanned: %d\n", report.TotalSegments)
	fmt.Printf("   Relevant pairs: %d\n", report.RelevantPairs)
	fmt.Printf("   Inconsistent: %d\n", report.Inconsistent)
	fmt.Printf("   Time: %dms\n", report.ExecutionTimeMs)

	if report.Inconsistent > 0 {
		fmt.Printf("\n❌ Inconsistencies found:\n\n")
		for _, r := range report.Results {
			if !r.Consistent {
				fmt.Printf("   • %s ↔ %s\n", r.Segment.Heading, r.Symbol.Name)
				fmt.Printf("     Reason: %s\n", r.Reason)
				if r.Suggestion != "" {
					fmt.Printf("     Suggestion: %s\n", r.Suggestion)
				}
				fmt.Println()
			}
		}
	}
}

func outputReportJSON(report *types.PRReport) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

func getChangeIcon(ct types.ChangeType) string {
	switch ct {
	case types.ChangeAdded:
		return "➕"
	case types.ChangeModified:
		return "📝"
	case types.ChangeDeleted:
		return "➖"
	default:
		return "•"
	}
}

func splitLines(s string, maxLines int) []string {
	var lines []string
	start := 0
	count := 0
	for i, c := range s {
		if c == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
			count++
			if count >= maxLines {
				lines = append(lines, "...")
				break
			}
		}
	}
	if count < maxLines && start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}
