package workstealer

import (
	"bytes"
	"fmt"
	"os"
	"text/tabwriter"
	"time"
)

// Record the current state of Model for observability
func (model Model) report(c client) error {
	now := time.Now()

	var b bytes.Buffer
	writer := tabwriter.NewWriter(&b, 0, 8, 1, '\t', tabwriter.AlignRight)

	fmt.Fprintf(writer, "lunchpail.io\tunassigned\t%d\t\t\t\t\t%s\t%s\n", len(model.UnassignedTasks), c.Spec.RunName, now.Format(time.UnixDate))
	fmt.Fprintf(writer, "lunchpail.io\tdispatcherDone\t%v\t\t\t\t\t%s\n", model.DispatcherDone, c.Spec.RunName)
	fmt.Fprintf(writer, "lunchpail.io\tassigned\t%d\t\t\t\t\t%s\n", len(model.AssignedTasks), c.Spec.RunName)
	fmt.Fprintf(writer, "lunchpail.io\tprocessing\t\t%d\t\t\t\t%s\n", len(model.ProcessingTasks), c.Spec.RunName)
	fmt.Fprintf(writer, "lunchpail.io\tdone\t\t\t%d\t%d\t\t%s\n", len(model.SuccessfulTasks), len(model.FailedTasks), c.Spec.RunName)
	fmt.Fprintf(writer, "lunchpail.io\tliveworkers\t%d\t\t\t\t\t%s\n", len(model.LiveWorkers), c.Spec.RunName)
	fmt.Fprintf(writer, "lunchpail.io\tdeadworkers\t%d\t\t\t\t\t%s\n", len(model.DeadWorkers), c.Spec.RunName)

	for _, worker := range model.LiveWorkers {
		fmt.Fprintf(
			writer, "lunchpail.io\tliveworker\t%d\t%d\t%d\t%d\t%s\t%s\t%v\n",
			len(worker.assignedTasks), len(worker.processingTasks), worker.nSuccess, worker.nFail, worker.name, c.Spec.RunName, worker.killfilePresent,
		)
	}
	for _, worker := range model.DeadWorkers {
		fmt.Fprintf(
			writer, "lunchpail.io\tdeadworker\t%d\t%d\t%d\t%d\t%s\t%s\n",
			len(worker.assignedTasks), len(worker.processingTasks), worker.nSuccess, worker.nFail, worker.name, c.Spec.RunName,
		)
	}
	fmt.Fprintf(writer, "lunchpail.io\t---\n")

	writer.Flush()

	// for now, also log to stdout
	fmt.Fprintf(os.Stdout, b.String())

	// and write to the log file
	/*if err := os.MkdirAll(logDir, 0700); err != nil {
		return err
	}
	logFile := filepath.Join(logDir, "qstat.txt")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	if _, err := b.WriteTo(f); err != nil {
		return err
	}

	return c.reportChangedFile(logFile)*/
	return nil
}
