package config

import (
	"os"
	"testing"
)

func TestLoad_Defaults(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.LLM.Provider != "openai" {
		t.Errorf("expected provider 'openai', got '%s'", cfg.LLM.Provider)
	}

	if cfg.LLM.Model != "gpt-4" {
		t.Errorf("expected model 'gpt-4', got '%s'", cfg.LLM.Model)
	}

	if cfg.Output.Format != "text" {
		t.Errorf("expected format 'text', got '%s'", cfg.Output.Format)
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "test-key")
	defer os.Unsetenv("OPENAI_API_KEY")

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.LLM.APIKey != "test-key" {
		t.Errorf("expected API key 'test-key', got '%s'", cfg.LLM.APIKey)
	}
}
