//go:build full || manage

package boot

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"

	"lunchpail.io/pkg/be"
	"lunchpail.io/pkg/build"
	"lunchpail.io/pkg/fe"
	"lunchpail.io/pkg/ir/hlir"
	"lunchpail.io/pkg/ir/llir"
	q "lunchpail.io/pkg/ir/queue"
	s3 "lunchpail.io/pkg/runtime/queue"
	"lunchpail.io/pkg/runtime/queue/upload"
	"lunchpail.io/pkg/util"
)

type UpOptions struct {
	Inputs       []string
	DryRun       bool
	Watch        bool
	BuildOptions build.Options
	Executable   string
	NoRedirect   bool
}

func Up(ctx context.Context, backend be.Backend, opts UpOptions) error {
	pipelineContext, err := handlePipelineStdin()
	if err != nil {
		return err
	}

	ir, err := fe.PrepareForRun(pipelineContext, fe.PrepareOptions{NoDispatchers: pipelineContext.Run.Step > 0 || len(opts.Inputs) > 0}, opts.BuildOptions)
	if err != nil {
		return err
	}

	return upLLIR(ctx, backend, ir, opts)
}

func UpHLIR(ctx context.Context, backend be.Backend, ir hlir.HLIR, opts UpOptions) error {
	pipelineContext, err := handlePipelineStdin()
	if err != nil {
		return err
	}

	llir, err := fe.PrepareHLIRForRun(ir, pipelineContext, fe.PrepareOptions{NoDispatchers: pipelineContext.Run.Step > 0 || len(opts.Inputs) > 0}, opts.BuildOptions)
	if err != nil {
		return err
	}

	return upLLIR(ctx, backend, llir, opts)
}

func upLLIR(ctx context.Context, backend be.Backend, ir llir.LLIR, opts UpOptions) error {
	if opts.DryRun {
		out, err := backend.DryRun(ir, opts.BuildOptions)
		if err != nil {
			return err
		}
		fmt.Printf(out)
		return nil
	}

	if ir.Context.Run.Step == 0 && !ir.HasDispatcher() && len(opts.Inputs) == 0 {
		return fmt.Errorf("please provide input files on the command line")
	}

	submissionComplete := make(chan struct{}) // is the job submission complete?
	cancellable, cancel := context.WithCancel(ctx)

	// Respond to SIGINT by cancelling our context. This will help
	// with cleaning up any loitering subprocesses, as Golang on
	// its own only kills the top level of a process tree. See
	// be/local/shell/spawn.go and its handling of context
	// cancellation by killing the process group it has created.
	var gotSigInt bool
	go func() {
		sigint := make(chan os.Signal)
		signal.Notify(sigint, os.Interrupt)

		// Wait for a SIGINT
		for {
			select {
			case <-cancellable.Done():
			case <-sigint:
				gotSigInt = true

				// Now cancel the context
				cancel()

				if err := backend.Down(ctx, ir, opts.BuildOptions); err != nil {
					fmt.Fprintln(os.Stderr, "Error bringing down run", err)
				}

				// And wait for all of the subprocesses to clean
				// themselves up. Because as soon as we exit from this
				// handler, the process will die. We need to wait for
				// the process group reaping to finish up. Sigh, why
				// is this so complicated in Golang?
				<-submissionComplete

				return
			}
		}
	}()

	if opts.Watch && !util.StdoutIsTty() {
		// if stdout is not a tty, then we can't support
		// watch, no matter what the user asked for
		fmt.Fprintf(os.Stderr, "Warning: disabling watch mode because stdout is not a tty\n")
		opts.Watch = false
	}

	// We need to chain the isRunning channel to our 0-2 consumers
	// below. This is because golang channels are not multicast.
	isRunning := make(chan llir.Context) // is the job ready for business?
	isRunning6 := make(chan llir.Context)
	needsCatAndRedirect := len(opts.Inputs) > 0 || ir.Context.Run.Step > 0 || ir.HasDispatcher()
	go func() {
		select {
		case <-cancellable.Done():
		case ctx := <-isRunning:
			isRunning6 <- ctx
			isRunning6 <- ctx
			isRunning6 <- ctx
			if opts.Executable != "" {
				isRunning6 <- ctx
			}
			if needsCatAndRedirect {
				isRunning6 <- ctx
			}
			if opts.Watch {
				isRunning6 <- ctx
			}
		}
	}()

	alldone := make(chan struct{})
	var errorFromAllDone error
	go func() {
		select {
		case <-cancellable.Done():
		case ctx := <-isRunning6:
			if ctx.Run.Step == 0 || isFinalStep(ctx) {
				errorFromAllDone = waitForAllDone(cancellable, backend, ctx.Run, *opts.BuildOptions.Log)
				if errorFromAllDone != nil && strings.Contains(errorFromAllDone.Error(), "connection refused") {
					// Then Minio went away on its own. That's probably ok.
					errorFromAllDone = nil
				}
				alldone <- struct{}{}
				cancel()
			} else {
				alldone <- struct{}{}
			}
		}
	}()

	var errorFromIo error
	redirectDone := make(chan struct{})
	if needsCatAndRedirect {
		// Behave like `cat inputs | ... > outputs`
		go func() {
			// wait for the run to be ready for us to enqueue
			select {
			case <-cancellable.Done():
			case <-isRunning6:
			}

			defer func() { redirectDone <- struct{}{} }()
			if err := catAndRedirect(cancellable, opts.Inputs, backend, ir, opts.NoRedirect, *opts.BuildOptions.Log); err != nil {
				errorFromIo = err
				cancel()
			}
		}()
	}

	logsDone := make(chan error)
	if opts.Watch {
		verbose := opts.BuildOptions.Log.Verbose
		go func() {
			select {
			case <-cancellable.Done():
			case <-isRunning6:
			}
			go watchLogs(cancellable, backend, ir, logsDone, WatchOptions{Verbose: verbose})
			go watchUtilization(cancellable, backend, ir, WatchOptions{Verbose: verbose})
		}()
	}

	go func() {
		select {
		case <-cancellable.Done():
		case ctx := <-isRunning6:
			if err := handlePipelineStdout(ctx); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}
	}()

	var errorFromTask error
	go func() {
		select {
		case <-cancellable.Done():
		case <-isRunning6:
		}
		if err := lookForTaskFailures(cancellable, backend, ir.Context.Run, *opts.BuildOptions.Log); err != nil {
			errorFromTask = err
			// fail fast? cancel()
		}
	}()

	//inject executable into s3
	if opts.Executable != "" {
		go func() {
			// wait for the run to be ready for us to enqueue
			select {
			case <-cancellable.Done():
			case <-isRunning6:
			}

			if err := s3.UploadFiles(cancellable, backend, ir.Context.Run, []upload.Upload{upload.Upload{LocalPath: opts.Executable, TargetDir: ir.Context.Run.AsFile(q.Blobs)}}, *opts.BuildOptions.Log); err != nil {
				fmt.Fprintln(os.Stderr, err)
			}
		}()
	}

	defer cancel()
	errorFromUp := backend.Up(cancellable, ir, opts.BuildOptions, isRunning)

	/* TODO defer func() {
		err := backend.Down(cancellable, ir, opts.BuildOptions)
	}()*/

	if errorFromUp != nil {
		// Oops, something failed in the up portion, i.e. before the job even started
		// Note the use of `select` to implement a non-blocking send
		select {
		case submissionComplete <- struct{}{}:
		default:
		}
		return errorFromUp
	}

	select {
	case <-cancellable.Done():
	case <-alldone:
	}

	if needsCatAndRedirect {
		select {
		case <-cancellable.Done():
		case <-redirectDone:
		}
	}

	// Note the use of `select` to implement a non-blocking send
	select {
	case submissionComplete <- struct{}{}:
	default:
	}

	if opts.Watch {
		cancel() // causes log streamer to stop
		<-logsDone
	}

	switch {
	case gotSigInt:
		// then squash any other errors as they are likely
		// side-effects of the user-initiated cancellation
	case errorFromTask != nil:
		return errorFromTask
	case errorFromIo != nil:
		return errorFromIo
	case errorFromUp != nil:
		return errorFromUp
	case errorFromAllDone != nil:
		return errorFromAllDone
	}

	return nil
}
