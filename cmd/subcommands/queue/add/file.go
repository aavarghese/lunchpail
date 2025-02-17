package add

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"lunchpail.io/cmd/options"
	q "lunchpail.io/pkg/ir/queue"
	"lunchpail.io/pkg/runtime/queue"
)

func File() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "file <file>",
		Short: "Enqueue a single file as a work task",
		Long:  "Enqueue a single file as a work task",
		Args:  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	}

	var opts queue.AddOptions
	var ignoreWorkerErrors bool
	cmd.Flags().BoolVarP(&opts.Wait, "wait", "w", false, "Wait for the task to be completed, and exit with the exit code of that task")
	cmd.Flags().BoolVar(&ignoreWorkerErrors, "ignore-worker-errors", false, "When --wait, ignore any errors from the workers processing the tasks")

	runOpts := options.AddRunOptions(cmd)
	logOpts := options.AddLogOptions(cmd)

	var next bool
	cmd.Flags().BoolVar(&next, "next", false, "Inject tasks into the next step")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if !opts.Wait && ignoreWorkerErrors {
			return fmt.Errorf("Invalid combination of options, not --wait and --ignore-worker-errors")
		}

		opts.LogOptions = *logOpts

		run, err := q.LoadRunContextInsideComponent(runOpts.Run)
		if err != nil {
			return err
		}

		step := run.Step
		if next {
			step++
		}
		exitcode, err := queue.Add(context.Background(), run.ForStep(step), args[0], opts)

		switch {
		case err != nil:
			return err
		case exitcode != 0 && !ignoreWorkerErrors:
			os.Exit(exitcode)
		}

		return nil
	}

	return cmd
}
