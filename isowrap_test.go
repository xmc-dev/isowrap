package isowrap

import (
	"io/ioutil"
	"testing"
)

func initBox(id uint, t *testing.T) *Box {
	b := NewBox()
	b.ID = id
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

func TestInitAndCleanupBox(t *testing.T) {
	b := initBox(42, t)
	cleanupBox(b, t)
}

func TestSuccessfulRunNoLimits(t *testing.T) {
	b := initBox(0, t)
	testProgram := []byte("#!/bin/sh\necho test")
	err := ioutil.WriteFile(b.Path+"/testProgram", testProgram, 0777)

	if err != nil {
		t.Error("Couldn't create test program: ", err)
	}

	result, err := b.Run("testProgram")
	t.Logf("Result: %+v", result)

	if err != nil {
		t.Fatal("Couldn't run test program: ", err)
	}

	cleanupBox(b, t)
}

func TestFailRunNoLimits(t *testing.T) {
	b := initBox(0, t)
	testProgram := []byte("#!/bin/sh\na=0\necho $((2/a))")
	err := ioutil.WriteFile(b.Path+"/testProgram", testProgram, 0777)

	if err != nil {
		t.Error("Couldn't create test program: ", err)
	}

	result, err := b.Run("testProgram")
	t.Logf("Result: %+v", result)

	if err != nil {
		t.Fatal("Couldn't run test program: ", err)
	}

	if result.ExitCode == 0 {
		t.Fatal("Program probably ran successfully")
	}

	cleanupBox(b, t)
}
