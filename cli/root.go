package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	verbose bool
)

var rootCmd = &cobra.Command{
	Use:   "docuguard",
	Short: "Documentation-code consistency checker",
	Long: `DocuGuard is a lightweight CLI tool for checking documentation-code consistency.

It uses LLM semantic analysis to detect conflicts between documentation and code.

Examples:
  docuguard check docs/api.md
  docuguard check --all
  docuguard init`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file path (default: .docuguard.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}
