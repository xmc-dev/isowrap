package isowrap

import (
	"bytes"
	"os"
	"os/exec"
	"time"
)

// ExecResult holds information about the program after execution.
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
		_, ok := err.(*exec.ExitError)
		if ok {
			err = nil
		} else {
			return
		}
	}
	result.State = cmd.ProcessState
	result.WallTime = elapsed

	stdout = string(bout.Bytes())
	stderr = string(berr.Bytes())
	return
}
