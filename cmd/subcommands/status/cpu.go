//go:build full || observe

package runs

import (
	"context"

	"github.com/spf13/cobra"

	"lunchpail.io/cmd/options"
	"lunchpail.io/pkg/be"
	"lunchpail.io/pkg/observe/cpu"
)

func Cpu() *cobra.Command {
	var intervalSecondsFlag int

	var cmd = &cobra.Command{
		Use:   "cpu",
		Short: "Displays CPU utilization",
		Long:  "Displays CPU utilization",
	}

	cmd.Flags().IntVarP(&intervalSecondsFlag, "interval", "i", 2, "Sampling interval")
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

		return cpu.UI(ctx, maybeRun, backend, cpu.CpuOptions{Verbose: opts.Log.Verbose, IntervalSeconds: intervalSecondsFlag})
	}

	return cmd
}
