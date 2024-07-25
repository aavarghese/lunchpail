package be

import (
	"errors"
	"fmt"
	"math"

	"github.com/IBM/vpc-go-sdk/vpcv1"
	"lunchpail.io/pkg/be/kubernetes"
	"lunchpail.io/pkg/be/platform"
	"lunchpail.io/pkg/observe/events"
)

func ChangeWorkers(poolName, poolNamespace string, poolPlatform platform.Platform, context string, delta int) error {
	if poolPlatform == platform.Kubernetes {
		return kubernetes.ChangeWorkers(poolName, poolNamespace, context, delta)
	}
	return nil
}

func ComputeParallelismAndInstanceCount(vpcService *vpcv1.VpcV1, profile string, workers int32) (parallelism int64, instanceCount int, err error) {
	prof, response, err := vpcService.GetInstanceProfile(
		&vpcv1.GetInstanceProfileOptions{
			Name: &profile,
		})
	if err != nil {
		return parallelism, instanceCount, fmt.Errorf("Failed to retrieve instance profile: %v and the response is: %s", err, response)
	}
	if prof != nil {
		numaCount, ok := prof.NumaCount.(*vpcv1.InstanceProfileNumaCount)
		if !ok {
			return parallelism, instanceCount, errors.New("Failed to get NumaCount from instance profile")
		}
		vcpuCount, ok := prof.VcpuCount.(*vpcv1.InstanceProfileVcpu)
		if !ok {
			return parallelism, instanceCount, errors.New("Failed to get VcpuCount from instance profile")
		}

		parallelism = (*vcpuCount.Value) * (*numaCount.Value)
		if workers < int32(parallelism) {
			parallelism = int64(workers)
		}
		instanceCount = max(1, int(math.Ceil(float64(workers)/float64(parallelism))))
	}

	return parallelism, instanceCount, nil
}

func StreamRunEvents(appname, runname, namespace string) (chan events.Message, error) {
	return kubernetes.StreamRunEvents(appname, runname, namespace)
}

func StreamRunComponentUpdates(appname, runname, namespace string) (chan events.ComponentUpdate, chan events.Message, error) {
	return kubernetes.StreamRunComponentUpdates(appname, runname, namespace)
}

func Ok(target platform.Platform) error {
	switch target {
	case platform.Kubernetes:
		return kubernetes.Ok()
	}

	return nil
}
