package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourname/docuguard/internal/git"
	"github.com/yourname/docuguard/pkg/types"
)

var (
	prBaseBranch string
	prDryRun     bool
	prFormat     string
	prDocs       []string
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
		return fmt.Errorf("failed to get diff: %w", err)
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

	// 输出变更符号
	if prFormat == "json" {
		return outputJSON(symbols)
	}

	outputText(symbols)

	if prDryRun {
		fmt.Println("\n💡 Use without --dry-run to check documentation consistency")
		return nil
	}

	// TODO: Phase 2-4 实现后，这里将调用文档扫描和一致性检查
	fmt.Println("\n⚠️  Documentation consistency check not yet implemented")
	fmt.Println("   Coming in Phase 2-4...")

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

	// TODO: Phase 5 实现 GitHub API 集成
	fmt.Printf("🔍 Checking PR #%d...\n", prNumber)
	fmt.Println("\n⚠️  GitHub mode not yet implemented")
	fmt.Println("   Coming in Phase 5...")

	return nil
}

func outputText(symbols []types.ChangedSymbol) {
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

func outputJSON(symbols []types.ChangedSymbol) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(map[string]interface{}{
		"symbols": symbols,
		"count":   len(symbols),
	})
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
