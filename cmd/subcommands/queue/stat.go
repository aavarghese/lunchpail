//go:build full || observe

package queue

import (
	"context"

	"github.com/spf13/cobra"

	"lunchpail.io/cmd/options"
	"lunchpail.io/pkg/be"
	"lunchpail.io/pkg/observe/qstat"
)

func Stat() *cobra.Command {
	var tailFlag int64
	var followFlag bool
	var quietFlag bool

	var cmd = &cobra.Command{
		Use:   "stat",
		Short: "Stream queue statistics to console",
	}

	cmd.Flags().BoolVarP(&followFlag, "follow", "f", false, "Track updates (rather than printing once)")
	cmd.Flags().Int64VarP(&tailFlag, "tail", "T", -1, "Number of lines to tail")
	cmd.Flags().BoolVarP(&quietFlag, "quiet", "q", false, "Silence extraneous output")

	opts, err := options.RestoreBuildOptions()
	if err != nil {
		panic(err)
	}

	options.AddTargetOptionsTo(cmd, &opts)
	options.AddLogOptionsTo(cmd, &opts)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		maybeRun := ""
		if len(args) > 0 {
			maybeRun = args[0]
		}

		ctx := context.Background()
		backend, err := be.New(ctx, opts)
		if err != nil {
			return err
		}

		return qstat.UI(ctx, maybeRun, backend, qstat.Options{Follow: followFlag, Tail: tailFlag, Verbose: opts.Log.Verbose, Quiet: quietFlag})
	}

	return cmd
}
