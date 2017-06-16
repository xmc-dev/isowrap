// +build freebsd

package isowrap

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// BoxRunner is a Runner based on jail
type BoxRunner struct {
	B *Box
}

func (br *BoxRunner) rctl(flag, rule string) error {
	params := []string{}
	params = append(
		params,
		flag,
		fmt.Sprintf("jail:isowrap%d:%s/jail", br.B.ID, rule),
	)
	_, _, _, err := Exec("rctl", params...)
	return err
}

// Init creates the jail and sets the rctl's
func (br *BoxRunner) Init() error {
	p := filepath.Join(os.TempDir(), fmt.Sprintf("isowrap%d", br.B.ID))
	err := os.Mkdir(p, os.ModePerm)
	if err != nil {
		return err
	}

	// Create jail
	params := []string{}
	params = append(
		params,
		"-c",
		fmt.Sprintf("name=isowrap%d", br.B.ID),
		"path="+p,
		"persist",
	)
	_, _, _, err = Exec("jail", params...)
	if err != nil {
		return err
	}

	cl := func(rule string) {
		if err != nil {
			return
		}
		err = br.rctl("-a", rule)
	}
	// Set resource limits
	if br.B.Config.MemoryLimit > 0 {
		cl(fmt.Sprintf("memoryuse:sigsegv=%dK", br.B.Config.MemoryLimit))
	}
	if err != nil {
		return err
	}

	br.B.Path = p
	return nil
}

// Run executes jexec to execute the given command.
func (br *BoxRunner) Run(command string) (result RunResult, err error) {
	params := []string{}
	params = append(
		params,
		fmt.Sprintf("%fs", br.B.Config.WallTime),
		"jexec",
		fmt.Sprintf("isowrap%d", br.B.ID),
		"/"+command,
	)

	result.ErrorType = NoError
	// Using timeout to limit wall time is a pretty dirty hack but such is life...
	stdout, stderr, r, err := Exec("timeout", params...)
	state := r.State
	result.Stdout = stdout
	result.Stderr = stderr

	ws, ok := state.Sys().(syscall.WaitStatus)
	if !ok {
		result.ErrorType = InternalError
		return
	}
	us, ok := state.SysUsage().(*syscall.Rusage)
	if !ok {
		result.ErrorType = InternalError
		return
	}
	result.ExitCode = ws.ExitStatus()
	if result.ExitCode == 124 {
		result.ErrorType = Timeout
	} else if result.ExitCode != 0 {
		result.ErrorType = RunTimeError
	}
	result.CPUTime = float64(state.SystemTime()+state.UserTime()) / float64(time.Second)
	result.MemUsed = uint(us.Maxrss)
	result.WallTime = float64(r.WallTime) / float64(time.Second)
	if err != nil {
		return
	}

	return
}

// Cleanup deletes the jail and its directory.
func (br *BoxRunner) Cleanup() error {
	// stop jail
	params := []string{}
	params = append(
		params,
		"-r",
		fmt.Sprintf("isowrap%d", br.B.ID),
	)
	_, _, _, err := Exec("rctl", "-r", fmt.Sprintf("jail:isowrap%d", br.B.ID))
	if err != nil {
		return err
	}
	_, _, _, err = Exec("jail", params...)
	if err != nil {
		return err
	}
	err = os.RemoveAll(br.B.Path)
	if err != nil {
		return err
	}
	return nil
}
