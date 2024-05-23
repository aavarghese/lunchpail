package status

import (
	"container/ring"
	"slices"

	"lunchpail.io/pkg/observe/cpu"
	"lunchpail.io/pkg/observe/qstat"

	"context"
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"path/filepath"
)

type WorkerStatus string

const (
	Pending     WorkerStatus = "Pending"
	Booting                  = "Booting"
	Running                  = "Running"
	Succeeded                = "Succeeded"
	Failed                   = "Failed"
	Terminating              = "Terminating"
)

type Worker struct {
	Name   string
	Status WorkerStatus
	Qstat  qstat.Worker
}

type Pool struct {
	Name        string
	Namespace   string
	Parallelism int
	Workers     []Worker
}

type Model struct {
	AppName       string
	RunName       string
	Namespace     string
	Pools         []Pool
	Runtime       WorkerStatus
	Dispatcher    WorkerStatus
	InternalS3    WorkerStatus
	WorkStealer   WorkerStatus
	LastNMessages *ring.Ring // ring of type Message
	Qstat         qstat.Model
	Cpu           cpu.Model
}

func (model *Model) numPools() int {
	return len(model.Pools)
}

func (model *Model) workers() []Worker {
	workers := []Worker{}
	for _, pool := range model.Pools {
		workers = slices.Concat(workers, pool.Workers)
	}
	return workers
}

// return the pair (numRunning, numTotal) of Runtime pods
func (model *Model) split(ws WorkerStatus) (int, int) {
	if ws == Running {
		return 1, 1
	} else {
		return 0, 1
	}
}

// return the pair (numRunning, numTotal) of Workers across all Pools
func (model *Model) workersSplit() (int, int) {
	running := 0
	total := 0

	for _, pool := range model.Pools {
		r, t := pool.workersSplit()
		running += r
		total += t
	}

	return running, total
}

// return the pair (numRunning, numTotal) of Workers for the given Pool
func (pool *Pool) workersSplit() (int, int) {
	running := 0
	total := len(pool.Workers)

	for _, worker := range pool.Workers {
		if worker.Status == Running {
			running++
		}
	}

	return running, total
}

// return the maximum size of any task box
func (model *Model) maxbox() int {
	N := max(model.Qstat.Unassigned, model.Qstat.Assigned, model.Qstat.Processing, model.Qstat.Success, model.Qstat.Failure)

	for _, pool := range model.Pools {
		for _, worker := range pool.Workers {
			N = max(N, worker.Qstat.Inbox, worker.Qstat.Processing, worker.Qstat.Outbox, worker.Qstat.Errorbox)
		}
	}

	return N
}

// return total across pools and workers of Inbox count
func (model *Model) allInbox() int {
	N := 0

	for _, pool := range model.Pools {
		for _, worker := range pool.Workers {
			N += worker.Qstat.Inbox
		}
	}

	return N
}

// return (inbox, processing, success, failure) across all Workers
func (pool *Pool) qsummary() (int, int, int, int) {
	inbox := 0
	processing := 0
	success := 0
	failure := 0

	for _, worker := range pool.Workers {
		inbox += worker.Qstat.Inbox
		processing += worker.Qstat.Processing
		success += worker.Qstat.Outbox
		failure += worker.Qstat.Errorbox
	}

	return inbox, processing, success, failure
}

func client() (*kubernetes.Clientset, error) {
	kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

func (pool *Pool) changeWorkers(delta int) error {
	clientset, err := client()
	if err != nil {
		return err
	}

	jobsClient := clientset.BatchV1().Jobs(pool.Namespace)
	patch := []byte(fmt.Sprintf(`{"spec": {"parallelism": %d}}`, pool.Parallelism+delta))

	if _, err := jobsClient.Patch(context.Background(), pool.Name, types.StrategicMergePatchType, patch, metav1.PatchOptions{}); err != nil {
		return err
	}

	return nil
}

func (pool *Pool) removeWorker() error {
	return nil
}