package transformer

import (
	"lunchpail.io/pkg/build"
	"lunchpail.io/pkg/fe/transformer/api/shell"
	"lunchpail.io/pkg/fe/transformer/api/workstealer"
	"lunchpail.io/pkg/ir/hlir"
	"lunchpail.io/pkg/ir/llir"
)

// HLIR -> LLIR for []hlir.Application
func lowerApplications(buildName, runname string, model hlir.HLIR, ir llir.LLIR, opts build.Options) ([]llir.Component, error) {
	components := []llir.Component{}

	if workstealer.IsNeeded(model) {
		// Note, the actual worker resources will be dealt
		// with when a WorkerPool is created. Here, we only
		// need to specify a WorkStealer.
		c, err := workstealer.Lower(buildName, runname, ir, opts)
		if err != nil {
			return nil, err
		}
		components = append(components, c)
	}

	// Then, for every non-Worker, we lower it as a "shell"
	for _, app := range model.Applications {
		if app.Spec.Role != hlir.WorkerRole {
			c, err := shell.Lower(buildName, runname, app, ir, opts)
			if err != nil {
				return nil, err
			}
			components = append(components, c)
		}
	}

	return components, nil
}
