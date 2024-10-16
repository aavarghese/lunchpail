package worker

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"lunchpail.io/pkg/runtime/queue"
	"lunchpail.io/pkg/util"
)

func killfileExists(client queue.S3Client, bucket, prefix string) bool {
	return client.Exists(bucket, prefix, "kill")
}

func startWatch(ctx context.Context, handler []string, client queue.S3Client, debug bool) error {
	bucket := client.Paths.Bucket
	prefix := client.Paths.Prefix
	inbox := client.Paths.Inbox
	processing := client.Paths.Processing
	outbox := client.Paths.Outbox
	alive := client.Paths.Alive
	// dead := client.Paths.dead

	if debug {
		fmt.Fprintf(os.Stderr, "Mkdirp bucket=%s\n", bucket)
	}
	if err := client.Mkdirp(bucket); err != nil {
		return err
	}

	if debug {
		fmt.Fprintf(os.Stderr, "Lunchpail worker touching alive file bucket=%s path=%s\n", bucket, alive)
	}
	err := client.Touch(bucket, alive)
	if err != nil {
		return err
	}

	s, err := util.SleepyTime("QUEUE_POLL_INTERVAL_SECONDS", 3)
	if err != nil {
		return err
	}

	localdir, err := os.MkdirTemp("", "lunchpail_local_queue_")
	if err != nil {
		return err
	}

	tasks, errs := client.Listen(client.Paths.Bucket, prefix, "", false)
	for {
		if killfileExists(client, bucket, prefix) {
			break
		}

		select {
		case err := <-errs:
			if debug {
				fmt.Fprintln(os.Stderr, err)
			}

			// sleep for a bit
			time.Sleep(s)

		case <-tasks:
		}

		tasks, err := client.Lsf(bucket, filepath.Join(prefix, inbox))
		if err != nil {
			return err
		}

		for _, task := range tasks {
			if task != "" {
				// TODO: re-check if task still exists in our inbox before starting on it
				in := filepath.Join(prefix, inbox, task)
				inprogress := filepath.Join(prefix, processing, task)
				out := filepath.Join(prefix, outbox, task)

				// capture exit code, stdout and stderr of the handler
				ec := filepath.Join(prefix, outbox, task+".code")
				succeeded := filepath.Join(prefix, outbox, task+".succeeded")
				failed := filepath.Join(prefix, outbox, task+".failed")
				stdout := filepath.Join(prefix, outbox, task+".stdout")
				stderr := filepath.Join(prefix, outbox, task+".stderr")

				localinbox := filepath.Join(localdir, inbox)
				localprocessing := filepath.Join(localdir, processing, task)
				localoutbox := filepath.Join(localdir, outbox, task)
				localec := localoutbox + ".code"
				localstdout := localoutbox + ".stdout"
				localstderr := localoutbox + ".stderr"

				err := os.MkdirAll(localinbox, os.ModePerm)
				if err != nil {
					fmt.Println("Internal Error creating local inbox:", err)
					continue
				}
				err = os.MkdirAll(filepath.Dir(localprocessing), os.ModePerm)
				if err != nil {
					fmt.Println("Internal Error creating local processing:", err)
					continue
				}
				err = os.MkdirAll(filepath.Dir(localoutbox), os.ModePerm)
				if err != nil {
					fmt.Println("Internal Error creating local outbox:", err)
					continue
				}

				err = client.Download(bucket, in, localprocessing)
				if err != nil {
					if !strings.Contains(err.Error(), "key does not exist") {
						// we ignore "key does not exist" errors, as these result from the work
						// we thought we were assigned having been stolen by the workstealer
						fmt.Printf("Internal Error copying task to worker processing %s %s->%s: %v\n", bucket, in, localprocessing, err)
					}
					continue
				}

				// fmt.Println("sending file to handler: " + in)
				err = os.Remove(localoutbox)
				if err != nil && !os.IsNotExist(err) {
					fmt.Println("Internal Error removing task from local outbox:", err)
					continue
				}

				err = client.Moveto(bucket, in, inprogress)
				if err != nil {
					fmt.Printf("Internal Error moving task to global processing %s->%s: %v\n", in, inprogress, err)
					continue
				}

				// signify that the process is still going... or prematurely terminated
				os.WriteFile(localec, []byte("-1"), os.ModePerm)

				handlercmd := exec.CommandContext(ctx, handler[0], slices.Concat(handler[1:], []string{localprocessing, localoutbox})...)

				// open stdout/err files for writing
				stdoutfile, err := os.Create(localstdout)
				if err != nil {
					fmt.Println("Internal Error creating stdout file:", err)
					continue
				}
				defer stdoutfile.Close()

				stderrfile, err := os.Create(localstderr)
				if err != nil {
					fmt.Println("Internal Error creating stderr file:", err)
					continue
				}
				defer stderrfile.Close()

				multiout := io.MultiWriter(os.Stdout, stdoutfile)
				multierr := io.MultiWriter(os.Stderr, stderrfile)
				handlercmd.Stdout = multiout
				handlercmd.Stderr = multierr
				err = handlercmd.Run()
				if err != nil {
					fmt.Println("Internal Error running the handler:", err)
					multierr.Write([]byte(err.Error()))
				}
				EC := handlercmd.ProcessState.ExitCode()

				os.WriteFile(localec, []byte(fmt.Sprintf("%d", EC)), os.ModePerm)

				err = client.Upload(bucket, localec, ec)
				if err != nil {
					fmt.Println("Internal Error moving exitcode to remote:", err)
				}

				err = client.Upload(bucket, localstdout, stdout)
				if err != nil {
					fmt.Println("Internal Error moving stdout to remote:", err)
				}

				err = client.Upload(bucket, localstderr, stderr)
				if err != nil {
					fmt.Println("Internal Error moving stderr to remote:", err)
				}

				if EC == 0 {
					if debug {
						fmt.Printf("Worker succeeded on task %s\n", localprocessing)
					}
					err = client.Touch(bucket, succeeded)
					if err != nil {
						fmt.Println("Internal Error creating succeeded marker:", err)
					}
				} else {
					err = client.Touch(bucket, failed)
					if err != nil {
						fmt.Println("Internal Error creating failed marker:", err)
					}
					fmt.Println("Worker error with exit code " + strconv.Itoa(EC) + " while processing " + filepath.Base(in))
				}

				if _, err := os.Stat(localoutbox); err == nil {
					if debug {
						fmt.Printf("Uploading worker-produced outbox file %s->%s\n", localoutbox, out)
					}
					if err := client.Upload(bucket, localoutbox, out); err != nil {
						fmt.Printf("Internal Error uploading task to global outbox %s->%s: %v\n", localoutbox, out, err)
					}
				} else if err := client.Moveto(bucket, inprogress, out); err != nil {
					fmt.Printf("Internal Error moving task to global outbox %s->%s: %v\n", inprogress, out, err)
				}
			}
		}
	}

	if debug {
		fmt.Println("Worker exiting normally")
	}

	return nil
}
