package isowrap

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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

func runTest(b *Box, t *testing.T, args ...string) RunResult {

	result, err := b.Run("testProgram", args...)
	t.Logf("Result: %+v", result)
	if err != nil {
		t.Fatal("Couldn't run test program: ", err)
	}
	return result
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
		t.Fatal("Couldn't write test program '" + testProgram + "' to its final destination")
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
			_, _, _, err := Exec(
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

	runTest(b, t, "0")
	cleanupBox(b, t)
}

func TestFailRunNoLimits(t *testing.T) {
	b := newBox(0)
	initBox(b, t)
	copyTest("no_limits", b, t)

	result := runTest(b, t, "1")
	cleanupBox(b, t)

	if result.ExitCode != 1 {
		t.Fatal("Exit code not 1")
	}
}

func TestFailTimeLimit(t *testing.T) {
	b := newBox(0)
	b.Config.WallTime = time.Duration(0.3 * float64(time.Second))
	initBox(b, t)
	copyTest("fail_time_limit", b, t)

	result := runTest(b, t)
	cleanupBox(b, t)

	if result.ErrorType != Timeout {
		t.Error("Program didn't exit because of timeout")
	}
}

func TestFailSigsegv(t *testing.T) {
	b := newBox(0)
	initBox(b, t)
	copyTest("fail_sigsegv", b, t)

	result := runTest(b, t)
	cleanupBox(b, t)

	if result.ErrorType != KilledBySignal {
		t.Error("Program didn't exit because of runtime error")
	}
}

func TestEnvInherit(t *testing.T) {
	os.Setenv("ISOWRAP_SPECIAL_VAL", "test432")
	b := newBox(0)
	b.Config.Env = append(b.Config.Env, EnvPair{"ISOWRAP_SPECIAL_VAL", ""})
	initBox(b, t)
	copyTest("env_val", b, t)

	result := runTest(b, t)
	cleanupBox(b, t)

	if result.ErrorType != NoError {
		t.Error("Program failed")
	}
	if strings.TrimSpace(result.Stdout) != os.Getenv("ISOWRAP_SPECIAL_VAL") {
		t.Error("Program returned the wrong value for the given environment variable")
	}
}

func TestEnvValue(t *testing.T) {
	b := newBox(0)
	b.Config.Env = append(b.Config.Env, EnvPair{"ISOWRAP_SPECIAL_VAL", "test321"})
	initBox(b, t)
	copyTest("env_val", b, t)

	result := runTest(b, t)
	cleanupBox(b, t)

	if result.ErrorType != NoError {
		t.Error("Program failed")
	}
	if strings.TrimSpace(result.Stdout) != "test321" {
		t.Error("Program returned the wrong value for the given environment variable")
	}
}

func TestFailProcLimit(t *testing.T) {
	b := newBox(0)
	b.Config.MaxProc = 3
	initBox(b, t)
	copyTest("proc_limit", b, t)

	result := runTest(b, t, strconv.FormatUint(uint64(b.Config.MaxProc+1), 10))
	cleanupBox(b, t)

	if result.ErrorType != RunTimeError {
		t.Error("Program didn't get runtime error")
	}
}
