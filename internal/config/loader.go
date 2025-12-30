package config

import (
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Load reads configuration from file and environment variables.
// If configPath is empty, it searches for .docuguard.yaml in current
// directory and home directory.
func Load(configPath string) (*Config, error) {
	v := viper.New()

	setDefaults(v)

	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName(".docuguard")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("$HOME")
	}

	v.SetEnvPrefix("DOCUGUARD")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if strings.HasPrefix(cfg.LLM.APIKey, "${") && strings.HasSuffix(cfg.LLM.APIKey, "}") {
		envVar := strings.TrimSuffix(strings.TrimPrefix(cfg.LLM.APIKey, "${"), "}")
		cfg.LLM.APIKey = os.Getenv(envVar)
	}

	if cfg.LLM.APIKey == "" {
		cfg.LLM.APIKey = os.Getenv("OPENAI_API_KEY")
	}

	if cfg.LLM.BaseURL == "" {
		if baseURL := os.Getenv("OPENAI_API_BASE"); baseURL != "" {
			cfg.LLM.BaseURL = baseURL
		}
	}

	if envModel := os.Getenv("OPENAI_MODEL"); envModel != "" {
		cfg.LLM.Model = envModel
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("version", "1.0")
	v.SetDefault("llm.provider", "openai")
	v.SetDefault("llm.model", "gpt-4")
	v.SetDefault("llm.timeout", "30s")
	v.SetDefault("scan.include", []string{"**/*.md"})
	v.SetDefault("scan.exclude", []string{})
	v.SetDefault("rules.fail_on_inconsistent", true)
	v.SetDefault("rules.severity_threshold", "warning")
	v.SetDefault("rules.confidence_threshold", 0.8)
	v.SetDefault("output.format", "text")
	v.SetDefault("output.color", true)
}
