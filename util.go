package isowrap

import (
	"bytes"
	"os"
	"os/exec"
	"time"
)

type ExecResult struct {
	State    *os.ProcessState
	WallTime time.Duration
}

// Exec executes a command and returns its stdout, stderr and exit status
func Exec(program string, args ...string) (stdout string, stderr string, result ExecResult, err error) {
	var bout, berr bytes.Buffer
	cmd := exec.Command(program, args...)
	cmd.Stdout = &bout
	cmd.Stderr = &berr

	start := time.Now()

	err = cmd.Run()
	elapsed := time.Since(start)
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return
		} else {
			err = nil
		}
	}
	result.State = cmd.ProcessState
	result.WallTime = elapsed

	return
}
