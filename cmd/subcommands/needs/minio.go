package needs

import (
	"context"

	"github.com/spf13/cobra"

	"lunchpail.io/cmd/options"
	"lunchpail.io/pkg/runtime/needs"
)

func Minio() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "minio <version>",
		Short: "Install minio",
		Long:  "Install minio",
		Args:  cobra.MatchAll(cobra.MaximumNArgs(1), cobra.OnlyValidArgs),
	}

	logOpts := options.AddLogOptions(cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		version := "latest"
		if len(args) > 0 {
			version = args[0]
		}

		return needs.InstallMinio(context.Background(), version, needs.Options{LogOptions: *logOpts})
	}

	return cmd
}