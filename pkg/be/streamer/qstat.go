package streamer

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"lunchpail.io/pkg/be/events/qstat"
)

func QstatFromStream(ctx context.Context, stream io.ReadCloser, c chan qstat.Model) error {
	lines := make(chan string)
	go func() {
		sc := bufio.NewScanner(stream)
		for sc.Scan() {
			lines <- sc.Text()
		}
		close(lines)
	}()

	return QstatFromChan(ctx, lines, c)
}

func QstatFromChan(ctx context.Context, lines chan string, c chan qstat.Model) error {
	model := qstat.Model{}
	for line := range lines {
		if !strings.HasPrefix(line, "lunchpail.io") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}

		marker := fields[1]
		if marker == "---" && model.Valid {
			select {
			case <-ctx.Done():
				break
			default:
				c <- model
			}

			model = qstat.Model{Valid: true}
			continue
		} else if len(fields) >= 3 {
			count, err := strconv.Atoi(fields[2])
			if err != nil {
				continue
			}

			if marker == "unassigned" {
				model.Valid = true
				model.Unassigned = count
				model.Timestamp = strings.Join(fields[4:], " ")
			} else if marker == "assigned" {
				model.Assigned = count
			} else if marker == "processing" {
				model.Processing = count
			} else if marker == "done" {
				count2, err := strconv.Atoi(fields[3])
				if err != nil {
					continue
				}

				model.Success = count
				model.Failure = count2
			} else if marker == "liveworker" || marker == "deadworker" {
				count2, err := strconv.Atoi(fields[3])
				if err != nil {
					continue
				}
				count3, err := strconv.Atoi(fields[4])
				if err != nil {
					continue
				}
				count4, err := strconv.Atoi(fields[5])
				if err != nil {
					continue
				}

				// The workstealer labels workers with
				// "poolName/workerName". TODO figure
				// out how to avoid this abstraction
				// violation?
				poolName, workerName, err := extractNamesFromSubPathForWorker(fields[6])
				if err != nil {
					continue
				}

				worker := qstat.Worker{
					Name: workerName, Inbox: count, Processing: count2, Outbox: count3, Errorbox: count4,
				}

				pidx := slices.IndexFunc(model.Pools, func(pool qstat.Pool) bool { return pool.Name == poolName })
				var pool qstat.Pool
				if pidx < 0 {
					// new pool
					pool = qstat.Pool{Name: poolName}
				} else {
					pool = model.Pools[pidx]
				}

				if marker == "liveworker" {
					pool.LiveWorkers = append(pool.LiveWorkers, worker)
				} else {
					pool.DeadWorkers = append(pool.DeadWorkers, worker)
				}

				if pidx < 0 {
					model.Pools = append(model.Pools, pool)
				} else {
					model.Pools = slices.Concat(model.Pools[:pidx], []qstat.Pool{pool}, model.Pools[pidx+1:])
				}
			}
		}
	}

	return nil
}

// The workstealer labels workers with "poolName/workerName". TODO
// figure out how to avoid this abstraction violation?
// e.g. test7f-pool1/w96bh -> (test7f-pool1,w96bh)
func extractNamesFromSubPathForWorker(combo string) (poolName string, workerName string, err error) {
	if idx := strings.Index(combo, "/"); idx < 0 {
		// TODO error handling here. what do we want to do?
		err = fmt.Errorf("Invalid subpath %s", combo)
	} else {
		poolName = combo[:idx]
		workerName = combo[idx+1:]
	}
	return
}
