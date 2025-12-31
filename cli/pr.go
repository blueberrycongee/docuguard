package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/blueberrycongee/docuguard/internal/config"
	"github.com/blueberrycongee/docuguard/internal/engine"
	"github.com/blueberrycongee/docuguard/internal/git"
	"github.com/blueberrycongee/docuguard/internal/github"
	"github.com/blueberrycongee/docuguard/internal/matcher"
	"github.com/blueberrycongee/docuguard/internal/reporter"
	"github.com/blueberrycongee/docuguard/internal/scanner"
	"github.com/blueberrycongee/docuguard/internal/ui"
	"github.com/blueberrycongee/docuguard/pkg/types"
)

var (
	prBaseBranch string
	prDryRun     bool
	prFormat     string
	prDocs       []string
	prSkipLLM    bool
	prTwoStage   bool
	prGitHub     bool
	prNumber     int
	prToken      string
	prRepo       string
	prComment    bool
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
	prCmd.Flags().StringVar(&prBaseBranch, "base", "main", "base branch for comparison")
	prCmd.Flags().BoolVar(&prDryRun, "dry-run", false, "only show detected changes, skip consistency check")
	prCmd.Flags().StringVar(&prFormat, "format", "text", "output format (text|json)")
	prCmd.Flags().StringSliceVar(&prDocs, "docs", []string{"README.md", "docs/**/*.md"}, "documentation patterns to scan")
	prCmd.Flags().BoolVar(&prSkipLLM, "skip-llm", false, "skip LLM check, use keyword matching only")
	prCmd.Flags().BoolVar(&prTwoStage, "two-stage", false, "use two-stage matching (broad match + LLM relevance filter)")

	prCmd.Flags().BoolVar(&prGitHub, "github", false, "enable GitHub mode")
	prCmd.Flags().IntVar(&prNumber, "pr", 0, "PR number (required in GitHub mode)")
	prCmd.Flags().StringVar(&prToken, "token", "", "GitHub token (or use GITHUB_TOKEN env)")
	prCmd.Flags().StringVar(&prRepo, "repo", "", "repository owner/repo (auto-detected from git remote)")
	prCmd.Flags().BoolVar(&prComment, "comment", false, "post comment on PR")

	rootCmd.AddCommand(prCmd)
}

func runPR(cmd *cobra.Command, args []string) error {
	if !git.IsInGitRepo() {
		return fmt.Errorf("not in a git repository")
	}

	if prGitHub {
		return runPRGitHub()
	}
	return runPRLocal()
}

func runPRLocal() error {
	printer := ui.NewPrinter(os.Stdout, false)

	printer.Info("Analyzing changes from %s...", prBaseBranch)
	fmt.Println()

	diff, err := git.GetDiffUncommitted()
	if err != nil {
		return fmt.Errorf("failed to get uncommitted diff: %w", err)
	}

	if diff != "" {
		printer.Info("Checking uncommitted changes...")
	} else {
		diff, err = git.GetDiff(prBaseBranch)
		if err != nil {
			return fmt.Errorf("failed to get diff: %w", err)
		}
	}

	if diff == "" {
		printer.Warning("No changes detected")
		return nil
	}

	extractor := git.NewSymbolExtractor()
	symbols, err := extractor.ExtractChangedSymbols(diff)
	if err != nil {
		return fmt.Errorf("failed to extract symbols: %w", err)
	}

	if len(symbols) == 0 {
		printer.Warning("No Go symbol changes detected")
		return nil
	}

	if prDryRun {
		if prFormat == "json" {
			return outputSymbolsJSON(symbols)
		}
		outputSymbolsText(symbols, printer)
		fmt.Println()
		printer.Info("Use without --dry-run to check documentation consistency")
		return nil
	}

	printer.Info("Scanning documentation...")
	segments, err := scanner.ScanMarkdownDir(".", prDocs)
	if err != nil {
		return fmt.Errorf("failed to scan documents: %w", err)
	}

	if len(segments) == 0 {
		printer.Warning("No documentation found matching patterns")
		return nil
	}
	printer.Success("Found %d document segments", len(segments))
	fmt.Println()

	printer.Info("Finding relevant documentation...")
	relevantPairs := matcher.QuickMatch(symbols, segments)
	printer.Success("Found %d potential matches", len(relevantPairs))
	fmt.Println()

	if len(relevantPairs) == 0 {
		printer.Success("No documentation appears to be affected by these changes")
		return nil
	}

	if !prSkipLLM {
		cfg, err := config.Load(cfgFile)
		if err == nil && cfg.LLM.APIKey != "" {
			return runPRWithLLM(cfg, diff)
		}
		printer.Warning("No LLM configured, using keyword matching only")
	}

	if prFormat == "json" {
		return outputRelevanceJSON(relevantPairs)
	}
	outputRelevanceText(relevantPairs, printer)

	return nil
}

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
		UseTwoStage: prTwoStage,
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

func runPRGitHub() error {
	if prNumber == 0 {
		return fmt.Errorf("--pr flag is required in GitHub mode")
	}

	token := prToken
	if token == "" {
		token = os.Getenv("GITHUB_TOKEN")
	}
	if token == "" {
		return fmt.Errorf("GitHub token required: use --token or set GITHUB_TOKEN env")
	}

	fmt.Printf("Checking PR #%d...\n\n", prNumber)

	ghClient, err := github.NewClient(token, prRepo)
	if err != nil {
		return fmt.Errorf("failed to create GitHub client: %w", err)
	}

	prInfo, err := ghClient.GetPRInfo(prNumber)
	if err != nil {
		return fmt.Errorf("failed to get PR info: %w", err)
	}
	fmt.Printf("PR: %s\n", prInfo.Title)
	fmt.Printf("Base: %s <- Head: %s\n\n", prInfo.BaseBranch, prInfo.HeadBranch)

	files, err := ghClient.GetPRFiles(prNumber)
	if err != nil {
		return fmt.Errorf("failed to get PR files: %w", err)
	}

	// Try to get full diff first (more accurate), fallback to patch from files
	diff, err := ghClient.GetPRDiff(prNumber)
	if err != nil || diff == "" {
		diff = github.BuildDiffFromFiles(files)
	}
	if diff == "" {
		fmt.Println("No changes detected")
		return nil
	}

	extractor := git.NewSymbolExtractor()
	symbols, err := extractor.ExtractChangedSymbols(diff)
	if err != nil {
		return fmt.Errorf("failed to extract symbols: %w", err)
	}

	if len(symbols) == 0 {
		fmt.Println("No Go symbol changes detected")
		return nil
	}

	fmt.Printf("Found %d changed symbol(s)\n", len(symbols))

	segments, err := scanner.ScanMarkdownDir(".", prDocs)
	if err != nil {
		return fmt.Errorf("failed to scan documents: %w", err)
	}
	fmt.Printf("Found %d document segments\n", len(segments))

	relevantPairs := matcher.QuickMatch(symbols, segments)
	fmt.Printf("Found %d potential matches\n\n", len(relevantPairs))

	// Use LLM for consistency check if configured
	var report *types.PRReport
	if !prSkipLLM {
		cfg, err := config.Load(cfgFile)
		if err == nil && cfg.LLM.APIKey != "" {
			ctx := context.Background()
			prEngine, err := engine.NewPREngine(cfg)
			if err != nil {
				return fmt.Errorf("failed to create PR engine: %w", err)
			}

			opts := engine.PRCheckOptions{
				BaseBranch:  prInfo.BaseBranch,
				DocPatterns: prDocs,
				SkipLLM:     prSkipLLM,
				UseTwoStage: prTwoStage,
			}

			report, err = prEngine.CheckFromDiff(ctx, diff, opts)
			if err != nil {
				return fmt.Errorf("failed to check: %w", err)
			}
		} else {
			// Fallback to keyword matching only
			report = &types.PRReport{
				TotalSymbols:  len(symbols),
				TotalSegments: len(segments),
				RelevantPairs: len(relevantPairs),
			}

			for _, pair := range relevantPairs {
				report.Results = append(report.Results, types.PRCheckResult{
					Segment:    pair.Segment,
					Symbol:     pair.Symbol,
					Consistent: true,
					Confidence: pair.Confidence,
					Reason:     pair.Reason,
				})
			}
		}
	} else {
		// Skip LLM, use keyword matching only
		report = &types.PRReport{
			TotalSymbols:  len(symbols),
			TotalSegments: len(segments),
			RelevantPairs: len(relevantPairs),
		}

		for _, pair := range relevantPairs {
			report.Results = append(report.Results, types.PRCheckResult{
				Segment:    pair.Segment,
				Symbol:     pair.Symbol,
				Consistent: true,
				Confidence: pair.Confidence,
				Reason:     pair.Reason,
			})
		}
	}

	if prComment {
		repoURL := fmt.Sprintf("https://github.com/%s/%s", ghClient.GetOwner(), ghClient.GetRepo())
		commentBody := reporter.FormatPRComment(report, repoURL)

		existingID, _ := ghClient.FindExistingComment(prNumber)
		if existingID > 0 {
			fmt.Println("Updating existing comment...")
			if err := ghClient.UpdateComment(existingID, commentBody); err != nil {
				return fmt.Errorf("failed to update comment: %w", err)
			}
		} else {
			fmt.Println("Creating comment...")
			if err := ghClient.CreateComment(prNumber, commentBody); err != nil {
				return fmt.Errorf("failed to create comment: %w", err)
			}
		}
		fmt.Println("Comment posted successfully")
	}

	if prFormat == "json" {
		return outputReportJSON(report)
	}
	outputReportText(report)

	return nil
}

func outputSymbolsText(symbols []types.ChangedSymbol, printer *ui.Printer) {
	printer.Success("Found %d changed symbol(s):", len(symbols))
	fmt.Println()

	for i, sym := range symbols {
		icon := getChangeIcon(sym.ChangeType)
		changeColor := getChangeColor(sym.ChangeType)
		fmt.Printf("%d. [%s] %s %s (%s)\n", i+1, changeColor(icon), sym.Type, ui.Highlight(sym.Name), ui.Dim(sym.File))
		fmt.Printf("   Lines: %s\n", ui.Dim(fmt.Sprintf("%d-%d", sym.StartLine, sym.EndLine)))

		if sym.NewCode != "" && len(sym.NewCode) < 200 {
			fmt.Printf("   Code:\n")
			for _, line := range splitLines(sym.NewCode, 3) {
				fmt.Printf("      %s\n", ui.Dim(line))
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

func outputRelevanceText(pairs []types.RelevanceResult, printer *ui.Printer) {
	printer.Info("Documentation that may need review:")
	fmt.Println()

	for i, pair := range pairs {
		fmt.Printf("%d. %s <-> %s\n", i+1, ui.Highlight(pair.Segment.Heading), ui.Highlight(pair.Symbol.Name))
		fmt.Printf("   Doc: %s (L%d-%d)\n", ui.Dim(pair.Segment.File), pair.Segment.StartLine, pair.Segment.EndLine)
		fmt.Printf("   Code: %s (L%d-%d)\n", ui.Dim(pair.Symbol.File), pair.Symbol.StartLine, pair.Symbol.EndLine)

		confidence := pair.Confidence * 100
		confidenceStr := fmt.Sprintf("%.0f%%", confidence)
		if confidence >= 70 {
			confidenceStr = ui.Success(confidenceStr)
		} else if confidence >= 40 {
			confidenceStr = ui.Warning(confidenceStr)
		} else {
			confidenceStr = ui.Dim(confidenceStr)
		}
		fmt.Printf("   Confidence: %s\n", confidenceStr)
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
	printer := ui.NewPrinter(os.Stdout, false)

	fmt.Println()
	printer.Info("Summary:")
	fmt.Printf("  Symbols changed: %s\n", ui.Highlight(fmt.Sprintf("%d", report.TotalSymbols)))
	fmt.Printf("  Documents scanned: %s\n", ui.Highlight(fmt.Sprintf("%d", report.TotalSegments)))
	fmt.Printf("  Relevant pairs: %s\n", ui.Highlight(fmt.Sprintf("%d", report.RelevantPairs)))

	if report.Inconsistent > 0 {
		fmt.Printf("  Inconsistent: %s\n", ui.Error(fmt.Sprintf("%d", report.Inconsistent)))
	} else {
		fmt.Printf("  Inconsistent: %s\n", ui.Success("0"))
	}
	fmt.Printf("  Time: %s\n", ui.Dim(fmt.Sprintf("%dms", report.ExecutionTimeMs)))

	if report.Inconsistent > 0 {
		fmt.Println()
		printer.Warning("Inconsistencies found:")
		fmt.Println()
		for _, r := range report.Results {
			if !r.Consistent {
				fmt.Printf("  - %s <-> %s\n", ui.Highlight(r.Segment.Heading), ui.Highlight(r.Symbol.Name))
				fmt.Printf("    %s: %s\n", ui.Error("Reason"), r.Reason)
				if r.Suggestion != "" {
					fmt.Printf("    %s: %s\n", ui.Info("Suggestion"), r.Suggestion)
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
		return "+"
	case types.ChangeModified:
		return "M"
	case types.ChangeDeleted:
		return "-"
	default:
		return "?"
	}
}

func getChangeColor(ct types.ChangeType) func(a ...interface{}) string {
	switch ct {
	case types.ChangeAdded:
		return ui.Success
	case types.ChangeModified:
		return ui.Warning
	case types.ChangeDeleted:
		return ui.Error
	default:
		return ui.Dim
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
