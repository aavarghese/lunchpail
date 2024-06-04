package api

import (
	_ "embed"
	"fmt"
	"lunchpail.io/pkg/fe/linker/queue"
	"lunchpail.io/pkg/ir/hlir"
	"lunchpail.io/pkg/lunchpail"
	"strconv"
)

//go:embed "parametersweep.sh"
var parameterSweepMain string

func appForSweep(sweep hlir.ParameterSweep) (hlir.Application, error) {
	app := hlir.Application{}
	app.ApiVersion = sweep.ApiVersion
	app.Kind = "Application"
	app.Metadata.Name = sweep.Metadata.Name
	app.Spec.Image = "ghcr.io/lunchpail/jaas-workerpool-worker-alpine-component:" + lunchpail.Version()
	app.Spec.Api = "shell"
	app.Spec.Role = "dispatcher"
	app.Spec.Command = "./main.sh"
	app.Spec.Code = []hlir.Code{
		hlir.Code{
			Name:   "main.sh",
			Source: parameterSweepMain,
		},
	}

	if sweep.Spec.Min < 0 {
		return app, fmt.Errorf("parameter sweep %s min should be >= 0, got %d", sweep.Metadata.Name, sweep.Spec.Min)
	}
	if sweep.Spec.Max <= 0 {
		return app, fmt.Errorf("parameter sweep %s max should be >= 0, got %d", sweep.Metadata.Name, sweep.Spec.Max)
	}
	if sweep.Spec.Max <= sweep.Spec.Min {
		return app, fmt.Errorf("parameter sweep %s max should be >= min, got %d", sweep.Metadata.Name, sweep.Spec.Max)
	}
	if sweep.Spec.Step <= 0 {
		return app, fmt.Errorf("parameter sweep %s step should be >= 0, got %d", sweep.Metadata.Name, sweep.Spec.Step)
	}
	if sweep.Spec.Interval < 0 {
		return app, fmt.Errorf("parameter sweep %s interval should be > 0, got %d", sweep.Metadata.Name, sweep.Spec.Interval)
	}

	app.Spec.Env = hlir.Env{}
	for key, value := range sweep.Spec.Env {
		app.Spec.Env[key] = value
	}

	app.Spec.Env["__LUNCHPAIL_METHOD"] = "parametersweep"
	app.Spec.Env["__LUNCHPAIL_SWEEP_MIN"] = strconv.Itoa(sweep.Spec.Min)
	app.Spec.Env["__LUNCHPAIL_SWEEP_MAX"] = strconv.Itoa(sweep.Spec.Max)

	if sweep.Spec.Step > 0 {
		app.Spec.Env["__LUNCHPAIL_SWEEP_STEP"] = strconv.Itoa(sweep.Spec.Step)
	}

	if sweep.Spec.Interval > 0 {
		app.Spec.Env["__LUNCHPAIL_INTERVAL"] = strconv.Itoa(sweep.Spec.Interval)
	}

	return app, nil
}

func LowerParameterSweep(assemblyName, runname, namespace string, sweep hlir.ParameterSweep, queueSpec queue.Spec, repoSecrets []hlir.RepoSecret, verbose bool) ([]string, error) {
	app, err := appForSweep(sweep)
	if err != nil {
		return []string{}, err
	}

	return LowerShell(
		assemblyName,
		runname,
		namespace,
		app,
		queueSpec,
		repoSecrets,
		verbose,
	)
}
