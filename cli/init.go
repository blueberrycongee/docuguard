package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize configuration file",
	Long:  "Generate a .docuguard.yaml configuration file in the current directory",
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	configTemplate := `# DocuGuard Configuration
version: "1.0"

llm:
  provider: "openai"
  model: "gpt-4"
  api_key: "${OPENAI_API_KEY}"
  # base_url: ""

scan:
  include:
    - "docs/**/*.md"
    - "README.md"
  exclude:
    - "docs/archive/**"

rules:
  fail_on_inconsistent: true
  confidence_threshold: 0.8

output:
  format: "text"
  color: true
`

	if _, err := os.Stat(".docuguard.yaml"); err == nil {
		return fmt.Errorf(".docuguard.yaml already exists, delete it first to regenerate")
	}

	if err := os.WriteFile(".docuguard.yaml", []byte(configTemplate), 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	fmt.Println("Created .docuguard.yaml")
	fmt.Println("Edit the config file to set your API key.")
	return nil
}
