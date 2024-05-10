package shrinkwrap

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"golang.org/x/sync/errgroup"
	"lunchpail.io/pkg/lunchpail"
)

type LogsOptions struct {
	Namespace  string
	Verbose    bool
	Components []lunchpail.Component
}

func streamLogs(appname, namespace string, component lunchpail.Component, verbose bool) error {
	containers := "app"
	appSelector := ",app.kubernetes.io/part-of=" + appname
	if component == lunchpail.DispatcherComponent {
		containers = "main"
		// FIXME: the workdispatcher has an invalid part-of
		appSelector = ""
	} else if component == lunchpail.WorkStealerComponent {
		containers = "workstealer"
	} else if component == lunchpail.RuntimeComponent {
		containers = "controller"
		appSelector = ""
	} else if component == lunchpail.WorkersComponent {
		appSelector = ""
	}

	selector := "app.kubernetes.io/component=" + string(component) + appSelector
	cmdline := "kubectl logs -n " + namespace + " -l " + selector + " --tail=-1 -f -c " + containers + " --max-log-requests=99 | grep -v 'workerpool worker'"

	if verbose {
		fmt.Fprintf(os.Stderr, "Tracking logs of component=%s\n", component)
		fmt.Fprintf(os.Stderr, "Tracking logs via cmdline=%s\n", cmdline)
	}

	cmd := exec.Command("/bin/sh", "-c", cmdline)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}

func Logs(opts LogsOptions) error {
	appname := lunchpail.AssembledAppName()
	namespace := appname
	if opts.Namespace != "" {
		namespace = opts.Namespace
	}

	group, _ := errgroup.WithContext(context.Background())

	for _, component := range opts.Components {
		group.Go(func() error {
			return streamLogs(appname, namespace, component, opts.Verbose)
		})
	}

	return group.Wait()
}
