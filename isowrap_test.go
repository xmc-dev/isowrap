package isowrap

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

const testSrcDir = "test_src"
const testDataDir = "test_data"

func initBox(id uint, cfg BoxConfig, t *testing.T) *Box {
	b := NewBox()
	b.ID = id
	b.Config = cfg
	err := b.Init()

	if err != nil {
		t.Fatal("Couldn't initialize box: ", err)
	}

	return b
}

func cleanupBox(b *Box, t *testing.T) {
	err := b.Cleanup()

	if err != nil {
		t.Fatal("Couldn't cleanup box: ", err)
	}
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
		t.Fatal("Couldn't copy test program '" + testProgram + "'")
	}

	err = ioutil.WriteFile(to, data, 0777)
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

func runTest(b *Box, t *testing.T) RunResult {
	result, err := b.Run("testProgram")
	t.Logf("Result: %+v", result)
	if err != nil {
		t.Fatal("Couldn't run test program: ", err)
	}
	return result
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
	b := initBox(42, BoxConfig{}, t)
	cleanupBox(b, t)
}

func TestSuccessfulRunNoLimits(t *testing.T) {
	b := initBox(0, BoxConfig{}, t)
	copyTest("success_no_limits", b, t)

	runTest(b, t)
	cleanupBox(b, t)
}

func TestFailRunNoLimits(t *testing.T) {
	b := initBox(0, BoxConfig{}, t)
	copyTest("fail_no_limits", b, t)

	result := runTest(b, t)
	cleanupBox(b, t)

	if result.ExitCode == 0 {
		t.Fatal("Program probably ran successfully")
	}
}

func TestFailTimeLimit(t *testing.T) {
	cfg := BoxConfig{}
	cfg.WallTime = time.Duration(0.3 * float64(time.Second))
	b := initBox(0, cfg, t)
	copyTest("fail_time_limit", b, t)

	result := runTest(b, t)
	cleanupBox(b, t)

	if result.ErrorType != Timeout {
		t.Error("Program didn't exit because of timeout")
	}
}

func TestFailSigsegv(t *testing.T) {
	cfg := BoxConfig{}
	b := initBox(0, cfg, t)
	copyTest("fail_sigsegv", b, t)

	result := runTest(b, t)
	cleanupBox(b, t)

	if result.ErrorType != KilledBySignal {
		t.Error("Program didn't exit because of runtime error")
	}
}

func TestEnvFull(t *testing.T) {
	cfg := BoxConfig{}
	cfg.FullEnv = true
	b := initBox(0, cfg, t)
	copyTest("env_test", b, t)

	result := runTest(b, t)
	if result.ErrorType != NoError {
		t.Error("Program failed")
	}
	if strings.TrimSpace(result.Stdout) != os.Getenv("HOME") {
		t.Error("Program returned the wrong value for the given environment variable")
	}
}

func TestEnvInherit(t *testing.T) {
	cfg := BoxConfig{}
	os.Setenv("ISOWRAP_SPECIAL_VAL", "test432")
	cfg.Env = append(cfg.Env, EnvPair{"ISOWRAP_SPECIAL_VAL", ""})
	b := initBox(0, cfg, t)
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
	cfg := BoxConfig{}
	cfg.Env = append(cfg.Env, EnvPair{"ISOWRAP_SPECIAL_VAL", "test321"})
	b := initBox(0, cfg, t)
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
