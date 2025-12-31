package cli

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
	
	"github.com/blueberrycongee/docuguard/internal/ui"
)

// Version information, injected at build time.
var (
	Version   = "0.1.0"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		printer := ui.NewPrinter(os.Stdout, false)
		
		fmt.Printf("%s %s\n", ui.Highlight("DocuGuard"), ui.Success(Version))
		fmt.Printf("  Git Commit: %s\n", ui.Dim(GitCommit))
		fmt.Printf("  Build Date: %s\n", ui.Dim(BuildDate))
		fmt.Printf("  Go Version: %s\n", ui.Info(runtime.Version()))
		fmt.Printf("  OS/Arch:    %s\n", ui.Dim(fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)))
		
		_ = printer // avoid unused warning
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
