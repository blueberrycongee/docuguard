package engine

import (
	"context"
	"fmt"
	"time"

	"github.com/blueberrycongee/docuguard/internal/config"
	"github.com/blueberrycongee/docuguard/internal/llm"
	"github.com/blueberrycongee/docuguard/internal/parser"
	"github.com/blueberrycongee/docuguard/pkg/types"
)

// Engine is the core consistency checking engine.
type Engine struct {
	cfg       *config.Config
	llmClient llm.Client
	goParser  *parser.GoParser
}

// New creates a new Engine instance.
func New(cfg *config.Config) (*Engine, error) {
	client, err := llm.NewClient(
		cfg.LLM.Provider,
		cfg.LLM.Model,
		cfg.LLM.APIKey,
		cfg.LLM.BaseURL,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create LLM client: %w", err)
	}

	return &Engine{
		cfg:       cfg,
		llmClient: client,
		goParser:  parser.NewGoParser(),
	}, nil
}

// CheckFile checks a single documentation file for consistency.
func (e *Engine) CheckFile(ctx context.Context, docPath string) (*types.Report, error) {
	startTime := time.Now()

	bindings, err := parser.ExtractBindings(docPath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract bindings: %w", err)
	}

	report := &types.Report{
		TotalBindings: len(bindings),
		Results:       make([]types.CheckResult, 0, len(bindings)),
	}

	for _, binding := range bindings {
		result, err := e.checkBinding(ctx, binding)
		if err != nil {
			report.Errors++
			continue
		}

		if result.Consistent {
			report.Consistent++
		} else {
			report.Inconsistent++
		}
		report.Results = append(report.Results, *result)
	}

	report.ExecutionTimeMs = time.Since(startTime).Milliseconds()
	return report, nil
}

func (e *Engine) checkBinding(ctx context.Context, binding types.Binding) (*types.CheckResult, error) {
	codeContent, codeLine, err := e.goParser.ExtractSymbol(
		binding.CodeFile,
		binding.CodeSymbol,
		binding.CodeType,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to extract code: %w", err)
	}

	if codeContent == "" {
		return &types.CheckResult{
			Consistent: false,
			Confidence: 1.0,
			Reason:     fmt.Sprintf("symbol %s not found in file %s", binding.CodeSymbol, binding.CodeFile),
			DocLoc:     types.Location{File: binding.DocFile, Line: binding.DocLine},
			CodeLoc:    types.Location{File: binding.CodeFile, Line: 0},
		}, nil
	}

	result, err := e.llmClient.Analyze(ctx, llm.AnalyzeRequest{
		DocContent:  binding.DocContent,
		CodeContent: codeContent,
		CodeSymbol:  binding.CodeSymbol,
		CodeFile:    binding.CodeFile,
	})
	if err != nil {
		return nil, fmt.Errorf("LLM analysis failed: %w", err)
	}

	result.DocLoc = types.Location{File: binding.DocFile, Line: binding.DocLine}
	result.CodeLoc = types.Location{File: binding.CodeFile, Line: codeLine, Symbol: binding.CodeSymbol}

	return result, nil
}
