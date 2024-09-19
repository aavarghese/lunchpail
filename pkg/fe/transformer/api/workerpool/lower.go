package workerpool

import (
	"fmt"
	"strconv"

	"lunchpail.io/pkg/compilation"
	"lunchpail.io/pkg/fe/transformer/api"
	"lunchpail.io/pkg/fe/transformer/api/shell"
	"lunchpail.io/pkg/ir/hlir"
	"lunchpail.io/pkg/ir/llir"
	"lunchpail.io/pkg/lunchpail"
)

func Lower(compilationName, runname string, app hlir.Application, pool hlir.WorkerPool, ir llir.LLIR, opts compilation.Options) (llir.Component, error) {
	spec := llir.ShellComponent{Component: lunchpail.WorkersComponent}

	spec.RunAsJob = true
	spec.Sizing = api.WorkerpoolSizing(pool, app, opts)
	spec.GroupName = pool.Metadata.Name
	spec.InstanceName = fmt.Sprintf("%s-%s", pool.Metadata.Name, runname)
	spec.QueuePrefixPath = api.QueuePrefixPathForWorker(ir.Queue, runname, pool.Metadata.Name)

	startupDelay, err := parseHumanTime(pool.Spec.StartupDelay)
	if err != nil {
		return nil, err
	}
	if app.Spec.Env == nil {
		app.Spec.Env = make(map[string]string)
	}
	app.Spec.Env["LUNCHPAIL_STARTUP_DELAY"] = strconv.Itoa(startupDelay)

	app.Spec.Command = fmt.Sprintf(`trap "$LUNCHPAIL_EXE worker prestop" EXIT
$LUNCHPAIL_EXE worker run --debug=%v -- %s`, opts.Log.Debug, app.Spec.Command)

	return shell.LowerAsComponent(
		compilationName,
		runname,
		app,
		ir,
		spec,
		opts,
	)
}
