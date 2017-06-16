package isowrap

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

type JailsRunner struct {
	B *Box
}

func (jb *JailsRunner) rctl(flag, rule string) error {
	params := []string{}
	params = append(
		params,
		flag,
		fmt.Sprintf("jail:isowrap%d:%s/jail", jb.B.ID, rule),
	)
	_, _, _, err := Exec("rctl", params...)
	return err
}

func (jb *JailsRunner) Init() error {
	p := filepath.Join(os.TempDir(), fmt.Sprintf("isowrap%d", jb.B.ID))
	err := os.Mkdir(p, os.ModePerm)
	if err != nil {
		return err
	}

	// Create jail
	params := []string{}
	params = append(
		params,
		"-c",
		fmt.Sprintf("name=isowrap%d", jb.B.ID),
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
		err = jb.rctl("-a", rule)
	}
	// Set resource limits
	if jb.B.Config.MemoryLimit > 0 {
		cl(fmt.Sprintf("memoryuse:sigsegv=%dK", jb.B.Config.MemoryLimit))
	}
	//if jb.B.Config.CPUTime > 0 {
	//	cl(fmt.Sprintf("cputime:sigsegv=%d", int(jb.B.Config.CPUTime)))
	//}
	//if jb.B.Config.WallTime > 0 {
	//	cl(fmt.Sprintf("wallclock:sigsegv=%d", int(jb.B.Config.WallTime)))
	//}
	if err != nil {
		return err
	}

	jb.B.Path = p
	return nil
}

func (jb *JailsRunner) Run(command string) (result RunResult, err error) {
	params := []string{}
	params = append(
		params,
		fmt.Sprintf("isowrap%d", jb.B.ID),
		"/"+command,
	)

	result.ErrorType = NoError
	stdout, stderr, r, err := Exec("jexec", params...)
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
	result.CPUTime = float64(state.SystemTime()+state.UserTime()) / float64(time.Second)
	result.MemUsed = uint(us.Maxrss)
	result.WallTime = float64(r.WallTime) / float64(time.Second)
	if err != nil {
		return
	}

	return
}

func (jb *JailsRunner) Cleanup() error {
	// stop jail
	params := []string{}
	params = append(
		params,
		"-r",
		fmt.Sprintf("isowrap%d", jb.B.ID),
	)
	_, _, _, err := Exec("rctl", "-r", fmt.Sprintf("jail:isowrap%d", jb.B.ID))
	if err != nil {
		return err
	}
	_, _, _, err = Exec("jail", params...)
	if err != nil {
		return err
	}
	err = os.RemoveAll(jb.B.Path)
	if err != nil {
		return err
	}
	return nil
}
