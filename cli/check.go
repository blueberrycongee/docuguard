package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yourname/docuguard/internal/config"
	"github.com/yourname/docuguard/internal/engine"
	"github.com/yourname/docuguard/internal/reporter"
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
		files = cfg.Scan.Include
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

		rep.Report(os.Stdout, report)

		if report.Inconsistent > 0 {
			hasInconsistent = true
		}
	}

	if hasInconsistent && cfg.Rules.FailOnInconsistent {
		os.Exit(1)
	}

	return nil
}
