package isowrap

import (
	"bytes"
	"os/exec"
)

// Exec executes a command and returns its stdout, stderr and exit status
func Exec(program string, args ...string) (stdout string, stderr string, exitOk bool, err error) {
	var bout, berr bytes.Buffer
	cmd := exec.Command(program, args...)
	cmd.Stdout = &bout
	cmd.Stderr = &berr

	err = cmd.Run()
	if err != nil {
		_, ok := err.(*exec.ExitError)
		if !ok {
			return "", "", false, err
		}
	}

	return bout.String(), berr.String(), true, nil
}
