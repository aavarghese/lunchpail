package subcommands

import (
	"os"

	"github.com/spf13/cobra"

	"lunchpail.io/pkg/build"
)

// RootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use: "lunchpail",
}

func Execute() error {
	if build.IsBuilt() {
		rootCmd.Use = build.Name()
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}

	return nil
}

func init() {
	initGroups(rootCmd)

	// To tell Cobra to mark the default completion command as
	// hidden (see
	// https://github.com/spf13/cobra/blob/main/site/content/completions/_index.md#adapting-the-default-completion-command)
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	// We still want usage errors for legitimate usage errors
	// (e.g. passing an unsupported dash option). We don't want it
	// for random errors emitted by RunE handlers. This trick
	// seems to accomplish that: register `SilenceUsage` only just
	// before (PreRun) the RunE is about to be invoked.
	// https://github.com/spf13/cobra/issues/340#issuecomment-378726225
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		rootCmd.SilenceUsage = true
	}
}
