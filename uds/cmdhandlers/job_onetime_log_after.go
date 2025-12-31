package cmdhandlers

import (
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/logitools/gw/framework"
	"github.com/logitools/gw/schedjobs"
)

type JobOnetimeLogAfter struct {
	AppProvider framework.AppProviderFunc
}

func (*JobOnetimeLogAfter) GroupName() string {
	return "job"
}

func (h *JobOnetimeLogAfter) Command() string {
	return "job-onetime-log-after"
}

func (h *JobOnetimeLogAfter) Desc() string {
	return "[TEST] Add a onetime scheduled job to log a message after given delay in minutes"
}

func (h *JobOnetimeLogAfter) Usage() string {
	return h.Command() + " delay message"
}

func (h *JobOnetimeLogAfter) HandleCommand(args []string, w io.Writer) error {
	argLen := len(args)
	if argLen < 2 {
		return fmt.Errorf("usage: %s", h.Usage())
	}
	delayStr := args[0]
	delayInMinutes, err := strconv.Atoi(delayStr)
	if err != nil {
		return err
	}
	msg := strings.Join(args[1:], " ")
	jobID := "log-msg-once"
	job := &schedjobs.OneTimeJob{
		ID:       jobID,
		ExecTime: time.Now().Add(time.Duration(delayInMinutes) * time.Minute),
		Task: func() error {
			log.Printf("[JOB] message: %s", msg)
			return nil
		},
	}
	appCore := h.AppProvider().AppCore()
	err = appCore.JobScheduler.AddOneTimeJob(job)
	if err != nil {
		return err
	}
	_, _ = fmt.Fprintf(w, "cron job %q scheduled", jobID)
	return nil
}
