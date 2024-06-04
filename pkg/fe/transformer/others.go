package transformer

import (
	"gopkg.in/yaml.v3"
	"lunchpail.io/pkg/ir/hlir"
)

// HLIR -> LLIR for non-lunchpail resources
func lowerOthers(assemblyName, runname string, model hlir.AppModel) ([]string, error) {
	yamls := []string{}

	for _, r := range model.Others {
		maybemetadata, ok := r["metadata"]
		if ok {
			if metadata, ok := maybemetadata.(hlir.UnknownResource); ok {
				var labels hlir.UnknownResource
				maybelabels, ok := metadata["labels"]
				if !ok || maybelabels == nil {
					labels = hlir.UnknownResource{}
				} else if yeslabels, ok := maybelabels.(hlir.UnknownResource); ok {
					labels = yeslabels
				}

				if labels != nil {
					labels["app.kubernetes.io/part-of"] = assemblyName
					labels["app.kubernetes.io/instance"] = runname
					labels["app.kubernetes.io/managed-by"] = "lunchpail.io"

					metadata["labels"] = labels
					r["metadata"] = metadata
				}
			}
		}

		yaml, err := yaml.Marshal(r)
		if err != nil {
			return []string{}, err
		}
		yamls = append(yamls, string(yaml))
	}

	return yamls, nil
}
