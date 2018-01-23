package isowrap

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

const testSrcDir = "test_src"
const testDataDir = "test_data"

func newBox(id uint) *Box {
	b := NewBox()
	b.ID = id

	return b
}

func initBox(b *Box, t *testing.T) {
	err := b.Init()
	if err != nil {
		t.Fatal("Couldn't initialize box: ", err)
	}
}

func cleanupBox(b *Box, t *testing.T) {
	err := b.Cleanup()

	if err != nil {
		t.Fatal("Couldn't cleanup box: ", err)
	}
}

func runTest(b *Box, t *testing.T, args ...string) (string, string, RunResult) {
	stdout, stderr, result, err := b.RunOutput("testProgram", args...)
	t.Logf("Result: %+v\nStdout: %s\nStderr: %s", result, stdout, stderr)
	if err != nil {
		t.Fatal("Couldn't run test program: ", err)
	}
	return stdout, stderr, result
}

func copyTest(testProgram string, b *Box, t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("Couldn't get working directory: ", err)
	}
	from := filepath.Join(wd, testDataDir, testProgram)
	to := filepath.Join(b.Path, "testProgram")
	data, err := ioutil.ReadFile(from)
	if err != nil {
		t.Fatal("Couldn't read test program '" + testProgram + "'")
	}

	err = ioutil.WriteFile(to, data, 0777)
	if err != nil {
		t.Fatal("Couldn't write test program '"+testProgram+"' to its final destination: ", err)
	}
}

func compileTestData() error {
	_ = os.Mkdir("test_data", os.ModePerm)
	wd, err := os.Getwd()
	if err != nil {
		return err
	}
	walk := func(path string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if f.IsDir() {
			return nil
		}

		if filepath.Ext(path) == ".c" {
			bn := strings.TrimSuffix(filepath.Base(path), ".c")
			log.Println("Compiling " + filepath.Base(path))
			_, err := Exec(nil, nil, nil,
				"cc",
				"-o",
				filepath.Join(wd, testDataDir, bn),
				"-static",
				path,
			)
			if err != nil {
				return err
			}
		}
		return nil
	}
	err = filepath.Walk(filepath.Join(wd, testSrcDir), walk)
	return err
}

func TestMain(m *testing.M) {
	log.Println("Compiling test programs")
	err := compileTestData()
	if err != nil {
		log.Fatal("Couldn't compile test data: ", err)
	}
	os.Exit(m.Run())
}

func TestInitAndCleanupBox(t *testing.T) {
	b := newBox(42)
	initBox(b, t)
	cleanupBox(b, t)
}

func TestSuccessfulRunNoLimits(t *testing.T) {
	b := newBox(0)
	initBox(b, t)
	copyTest("no_limits", b, t)

	_, _, result := runTest(b, t, "0")
	cleanupBox(b, t)

	if result.ErrorType != NoError {
		t.Fatal("Unexpected error: ", result.ErrorType)
	}
}

func TestFailRunNoLimits(t *testing.T) {
	b := newBox(0)
	initBox(b, t)
	copyTest("no_limits", b, t)

	_, _, result := runTest(b, t, "1")
	cleanupBox(b, t)

	if result.ErrorType != RunTimeError {
		t.Fatal("Unexpected error: ", result.ErrorType)
	}

	if result.ExitCode != 1 {
		t.Fatal("Exit code not 1")
	}
}

func TestFailTimeLimit(t *testing.T) {
	b := newBox(0)
	b.Config.WallTime = time.Duration(0.3 * float64(time.Second))
	initBox(b, t)
	copyTest("fail_time_limit", b, t)

	_, _, result := runTest(b, t)
	cleanupBox(b, t)

	if result.ErrorType != Timeout {
		t.Error("Program didn't exit because of timeout")
	}
}

func TestFailSigsegv(t *testing.T) {
	b := newBox(0)
	initBox(b, t)
	copyTest("fail_sigsegv", b, t)

	_, _, result := runTest(b, t)
	cleanupBox(b, t)

	if result.ErrorType != KilledBySignal && result.Signal != syscall.SIGSEGV {
		t.Error("Program didn't exit because of runtime error")
	}
}

func TestEnvInherit(t *testing.T) {
	os.Setenv("ISOWRAP_SPECIAL_VAL", "test432")
	b := newBox(0)
	b.Config.Env = append(b.Config.Env, EnvPair{"ISOWRAP_SPECIAL_VAL", ""})
	initBox(b, t)
	copyTest("env_val", b, t)

	stdout, _, result := runTest(b, t)
	cleanupBox(b, t)

	if result.ErrorType != NoError {
		t.Error("Program failed")
	}
	if strings.TrimSpace(stdout) != os.Getenv("ISOWRAP_SPECIAL_VAL") {
		t.Error("Program returned the wrong value for the given environment variable")
	}
}

func TestEnvValue(t *testing.T) {
	b := newBox(0)
	b.Config.Env = append(b.Config.Env, EnvPair{"ISOWRAP_SPECIAL_VAL", "test321"})
	initBox(b, t)
	copyTest("env_val", b, t)

	stdout, _, result := runTest(b, t)
	cleanupBox(b, t)

	if result.ErrorType != NoError {
		t.Error("Program failed")
	}
	if strings.TrimSpace(stdout) != "test321" {
		t.Error("Program returned the wrong value for the given environment variable")
	}
}

func TestFailProcLimit(t *testing.T) {
	b := newBox(0)
	b.Config.MaxProc = 5
	initBox(b, t)
	copyTest("proc_limit", b, t)

	_, _, result := runTest(b, t, strconv.FormatUint(uint64(b.Config.MaxProc), 10))
	cleanupBox(b, t)

	if result.ErrorType != RunTimeError {
		t.Error("Program didn't get runtime error")
	}
}

func TestSuccessProcLimit(t *testing.T) {
	b := newBox(0)
	b.Config.MaxProc = 5
	initBox(b, t)
	copyTest("proc_limit", b, t)

	_, _, result := runTest(b, t, strconv.FormatUint(uint64(b.Config.MaxProc-1), 10))
	cleanupBox(b, t)

	if result.ErrorType != NoError {
		t.Error("Program got an error")
	}
}

func TestFailMemoryLimit(t *testing.T) {
	b := newBox(0)
	b.Config.MemoryLimit = 2 * 1024
	b.Config.WallTime = 2 * time.Second
	initBox(b, t)
	copyTest("memory_limit", b, t)

	_, _, result := runTest(b, t)
	cleanupBox(b, t)

	if result.ErrorType != MemoryExceeded {
		t.Error("Program did not exceed memory")
	}
}
