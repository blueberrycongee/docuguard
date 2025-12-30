package config

import "time"

// Config 应用配置
type Config struct {
	Version string     `mapstructure:"version"`
	LLM     LLMConfig  `mapstructure:"llm"`
	Scan    ScanConfig `mapstructure:"scan"`
	Rules   RuleConfig `mapstructure:"rules"`
	Output  OutConfig  `mapstructure:"output"`
}

// LLMConfig LLM 配置
type LLMConfig struct {
	Provider string        `mapstructure:"provider"` // openai, anthropic, ollama
	Model    string        `mapstructure:"model"`
	APIKey   string        `mapstructure:"api_key"`
	BaseURL  string        `mapstructure:"base_url"`
	Timeout  time.Duration `mapstructure:"timeout"`
}

// ScanConfig 扫描配置
type ScanConfig struct {
	Include []string `mapstructure:"include"`
	Exclude []string `mapstructure:"exclude"`
}

// RuleConfig 规则配置
type RuleConfig struct {
	FailOnInconsistent  bool    `mapstructure:"fail_on_inconsistent"`
	SeverityThreshold   string  `mapstructure:"severity_threshold"`
	ConfidenceThreshold float64 `mapstructure:"confidence_threshold"`
}

// OutConfig 输出配置
type OutConfig struct {
	Format string `mapstructure:"format"` // text, json, github-actions
	Color  bool   `mapstructure:"color"`
}
