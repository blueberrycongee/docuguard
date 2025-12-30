package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/blueberrycongee/docuguard/internal/config"
	"github.com/blueberrycongee/docuguard/internal/engine"
	"github.com/blueberrycongee/docuguard/internal/reporter"
	"github.com/spf13/cobra"
)

var (
	checkAll     bool
	outputFormat string
)

var checkCmd = &cobra.Command{
	Use:   "check [files...]",
	Short: "Check documentation-code consistency",
	Long: `Check the consistency between Markdown documentation and associated code.

Examples:
  docuguard check docs/api.md
  docuguard check docs/api.md docs/payment.md
  docuguard check --all
  docuguard check --format json docs/api.md`,
	RunE: runCheck,
}

func init() {
	checkCmd.Flags().BoolVar(&checkAll, "all", false, "check all configured documents")
	checkCmd.Flags().StringVar(&outputFormat, "format", "text", "output format (text|json|github-actions)")
	rootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if outputFormat != "" {
		cfg.Output.Format = outputFormat
	}

	var files []string
	if checkAll {
		// Expand glob patterns to actual file paths
		files, err = expandGlobPatterns(".", cfg.Scan.Include)
		if err != nil {
			return fmt.Errorf("failed to expand patterns: %w", err)
		}
		if len(files) == 0 {
			return fmt.Errorf("no files found matching patterns: %v", cfg.Scan.Include)
		}
	} else if len(args) > 0 {
		files = args
	} else {
		return fmt.Errorf("please specify files to check or use --all")
	}

	eng, err := engine.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to initialize engine: %w", err)
	}

	rep := reporter.New(cfg.Output.Format, cfg.Output.Color)

	ctx := context.Background()
	hasInconsistent := false

	for _, file := range files {
		report, err := eng.CheckFile(ctx, file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to check %s: %v\n", file, err)
			continue
		}

		_ = rep.Report(os.Stdout, report)

		if report.Inconsistent > 0 {
			hasInconsistent = true
		}
	}

	if hasInconsistent && cfg.Rules.FailOnInconsistent {
		os.Exit(1)
	}

	return nil
}


// expandGlobPatterns expands glob patterns to actual file paths.
func expandGlobPatterns(rootDir string, patterns []string) ([]string, error) {
	seen := make(map[string]bool)
	var files []string

	for _, pattern := range patterns {
		fullPattern := filepath.Join(rootDir, pattern)
		matches, err := filepath.Glob(fullPattern)
		if err != nil {
			continue
		}

		for _, match := range matches {
			info, err := os.Stat(match)
			if err != nil || info.IsDir() {
				continue
			}
			if strings.HasSuffix(strings.ToLower(match), ".md") {
				if !seen[match] {
					seen[match] = true
					files = append(files, match)
				}
			}
		}

		// Handle non-glob patterns (direct file paths)
		if !strings.Contains(pattern, "*") {
			filePath := filepath.Join(rootDir, pattern)
			if info, err := os.Stat(filePath); err == nil && !info.IsDir() {
				if !seen[filePath] {
					seen[filePath] = true
					files = append(files, filePath)
				}
			}
		}
	}

	return files, nil
}
