package transformer

import (
	"fmt"
	"lunchpail.io/pkg/fe/linker/queue"
	"lunchpail.io/pkg/fe/transformer/api"
	"lunchpail.io/pkg/ir/hlir"
	"lunchpail.io/pkg/ir/llir"
	"slices"
)

// HLIR -> LLIR for []hlir.WorkerPool
func lowerWorkerPools(assemblyName, runname, namespace string, model hlir.AppModel, queueSpec queue.Spec, verbose bool) ([]llir.Yaml, error) {
	yamls := []llir.Yaml{}

	app, found := model.GetApplicationByRole(hlir.WorkerRole)
	if !found {
		return yamls, fmt.Errorf("No Application with role Worker found")
	}

	for _, pool := range model.WorkerPools {
		if tyamls, err := api.LowerWorkerPool(assemblyName, runname, namespace, app, pool, queueSpec, model.RepoSecrets, verbose); err != nil {
			return yamls, err
		} else {
			yamls = slices.Concat(yamls, tyamls)
		}
	}

	return yamls, nil
}
