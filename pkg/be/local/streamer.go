package local

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/nxadm/tail"
	"github.com/shirou/gopsutil/v4/process"
	"golang.org/x/sync/errgroup"

	"lunchpail.io/pkg/be/events/qstat"
	"lunchpail.io/pkg/be/events/utilization"
	"lunchpail.io/pkg/be/local/files"
	"lunchpail.io/pkg/be/streamer"
	"lunchpail.io/pkg/ir/queue"
	"lunchpail.io/pkg/lunchpail"
)

type localStreamer struct {
	context.Context
	run     queue.RunContext
	backend Backend
}

// Return a streamer
func (backend Backend) Streamer(ctx context.Context, run queue.RunContext) streamer.Streamer {
	return localStreamer{ctx, run, backend}
}

// Stream cpu and memory statistics
func (s localStreamer) Utilization(c chan utilization.Model, intervalSeconds int) error {
	for {
		ps, err := process.ProcessesWithContext(s.Context)
		if err != nil {
			return err
		}

		var m utilization.Model

		parts, err := partsOfRun(s.run.RunName)
		if err != nil {
			return err
		}

		for _, p := range ps {
			part, ok := parts[p.Pid]
			if !ok {
				continue
			}

			worker := utilization.Worker{Name: part.InstanceName, Component: part.Component}
			cpu, err := p.CPUPercentWithContext(s.Context)
			if err != nil {
				return err
			}
			worker.CpuUtil = cpu

			mem, err := p.MemoryInfoWithContext(s.Context)
			if err != nil {
				return err
			}
			worker.MemoryBytes = mem.RSS

			m.Workers = append(m.Workers, worker)
		}

		if len(m.Workers) > 0 {
			c <- m
		}

		select {
		case <-s.Context.Done():
			return nil
		default:
			time.Sleep(time.Duration(intervalSeconds) * time.Second)
		}
	}
}

// Stream queue statistics
func (s localStreamer) QueueStats(c chan qstat.Model, opts qstat.Options) error {
	f, err := files.LogsForComponent(s.run, lunchpail.WorkStealerComponent)
	if err != nil {
		return err
	}

	tail, err := tailfChan(f, streamer.LogOptions{Follow: opts.Follow, Verbose: opts.Verbose})
	if err != nil {
		return err
	}

	lines := make(chan string)
	errs, _ := errgroup.WithContext(s.Context)
	errs.Go(func() error {
		defer close(lines)
		for line := range tail.Lines {
			if line.Err != nil {
				return line.Err
			}
			lines <- line.Text
		}
		return nil
	})

	return streamer.QstatFromChan(s.Context, lines, c)
}

func (s localStreamer) watchForWorkerPools(logdir string, opts streamer.LogOptions) error {
	watching := make(map[string]bool)
	group, _ := errgroup.WithContext(s.Context)

	// TODO fsnotify/fsnotify doesn't seem to work on macos
	for {
		fs, err := os.ReadDir(logdir)
		if err != nil {
			return err
		}

		for _, f := range fs {
			file := f.Name()
			if strings.HasPrefix(file, lunchpail.ComponentShortName(lunchpail.WorkersComponent)) {
				alreadyWatching, exists := watching[file]
				if !alreadyWatching || !exists {
					watching[file] = true
					group.Go(func() error {
						return tailf(filepath.Join(logdir, file), opts)
					})
				}
			}
		}

		runStillGoing, err := isRunning(s.run.RunName)
		if err != nil {
			return err
		} else if !runStillGoing || !opts.Follow {
			break
		}

		select {
		case <-s.Context.Done():
			return nil
		default:
			time.Sleep(2 * time.Second)
		}
	}

	return group.Wait()
}

// Stream logs from a given Component to os.Stdout
func (s localStreamer) ComponentLogs(c lunchpail.Component, opts streamer.LogOptions) error {
	logdir, err := files.LogDir(s.run, true)
	if err != nil {
		return err
	}

	switch c {
	case lunchpail.WorkersComponent:
		return s.watchForWorkerPools(logdir, opts)

	default:
		// TODO allow caller to select stderr versus stdout
		group, _ := errgroup.WithContext(s.Context)
		group.Go(func() error { return tailf(filepath.Join(logdir, files.LogFileForComponent(c)+".out"), opts) })
		group.Go(func() error { return tailf(filepath.Join(logdir, files.LogFileForComponent(c)+".err"), opts) })
		return group.Wait()
	}
}

func tailfChan(outfile string, opts streamer.LogOptions) (*tail.Tail, error) {
	Logger := tail.DiscardingLogger
	if opts.Verbose {
		Logger = log.New(os.Stderr, "", log.LstdFlags)
	}

	return tail.TailFile(outfile, tail.Config{Follow: opts.Follow, ReOpen: opts.Follow, Logger: Logger})
}

func tailf(outfile string, opts streamer.LogOptions) error {
	c, err := tailfChan(outfile, opts)
	if err != nil {
		return err
	}

	for line := range c.Lines {
		prefix := ""
		if opts.LinePrefix != nil {
			prefix = opts.LinePrefix(strings.TrimSuffix(filepath.Base(outfile), filepath.Ext(outfile)))
		}
		fmt.Printf("%s%s\n", prefix, line.Text)
	}

	return nil
}
