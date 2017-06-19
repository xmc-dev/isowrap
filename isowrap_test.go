package isowrap

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

func copyTest(from, to string) error {
	data, err := ioutil.ReadFile(from)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(to, data, 0777)
	return err
}

func compileTestData() error {
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
	b := initBox(42, BoxConfig{}, t)
	cleanupBox(b, t)
}

func TestSuccessfulRunNoLimits(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("Couldn't get workding directory: ", err)
	}
	b := initBox(0, BoxConfig{}, t)
	copyTest(
		filepath.Join(wd, testDataDir, "success_no_limits"),
		filepath.Join(b.Path, "testProgram"),
	)

	if err != nil {
		t.Error("Couldn't create test program: ", err)
	}

	result, err := b.Run("testProgram")
	t.Logf("Result: %+v", result)
	cleanupBox(b, t)

	if err != nil {
		t.Fatal("Couldn't run test program: ", err)
	}
}

func TestFailRunNoLimits(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("Couldn't get working directory: ", err)
	}
	b := initBox(0, BoxConfig{}, t)
	copyTest(
		filepath.Join(wd, testDataDir, "fail_no_limits"),
		filepath.Join(b.Path, "testProgram"),
	)

	result, err := b.Run("testProgram")
	t.Logf("Result: %+v", result)
	cleanupBox(b, t)

	if err != nil {
		t.Fatal("Couldn't run test program: ", err)
	}

	if result.ExitCode == 0 {
		t.Fatal("Program probably ran successfully")
	}
}

func TestFailTimeLimit(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("Couldn't get working directory: ", err)
	}
	cfg := BoxConfig{}
	cfg.WallTime = 0.3
	b := initBox(0, cfg, t)
	copyTest(
		filepath.Join(wd, testDataDir, "fail_time_limit"),
		filepath.Join(b.Path, "testProgram"),
	)

	result, err := b.Run("testProgram")
	t.Logf("Result: %+v", result)
	cleanupBox(b, t)

	if err != nil {
		t.Fatal("Couldn't run test program: ", err)
	}
	if result.ErrorType != Timeout {
		t.Error("Program didn't exit because of timeout")
	}
}

func TestFailSigsegv(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal("Couldn't get working directory: ", err)
	}
	cfg := BoxConfig{}
	b := initBox(0, cfg, t)
	copyTest(
		filepath.Join(wd, testDataDir, "fail_sigsegv"),
		filepath.Join(b.Path, "testProgram"),
	)

	result, err := b.Run("testProgram")
	t.Logf("Result: %+v", result)
	cleanupBox(b, t)

	if err != nil {
		t.Fatal("Couldn't run test program: ", err)
	}
	if result.ErrorType != KilledBySignal {
		t.Error("Program didn't exit because of runtime error")
	}
}
