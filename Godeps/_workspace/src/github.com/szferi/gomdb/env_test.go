package mdb

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestEnvOpen(t *testing.T) {
	env, err := NewEnv()
	if err != nil {
		t.Errorf("Cannot create enviroment: %s", err)
	}
	err = env.Open("adsjgfadsfjg", 0, 0664)
	if err == nil {
		t.Errorf("should not be able to open")
	}
	path, err := ioutil.TempDir("/tmp", "mdb_test")
	if err != nil {
		t.Errorf("Cannot create temporary directory")
	}
	err = os.MkdirAll(path, 0770)
	if err != nil {
		t.Errorf("Cannot create directory: %s", path)
	}
	err = env.Open(path, 0, 0664)
	if err != nil {
		t.Errorf("Cannot open environment: %s", err)
	}
	err = env.Close()
	if err != nil {
		t.Errorf("Error during close of environment: %s", err)
	}
	// clean up
	os.RemoveAll(path)
}

func setup(t *testing.T) *Env {
	env, err := NewEnv()
	if err != nil {
		t.Errorf("Cannot create enviroment: %s", err)
	}
	path, err := ioutil.TempDir("/tmp", "mdb_test")
	if err != nil {
		t.Errorf("Cannot create temporary directory")
	}
	err = os.MkdirAll(path, 0770)
	if err != nil {
		t.Errorf("Cannot create directory: %s", path)
	}
	err = env.Open(path, 0, 0664)
	if err != nil {
		t.Errorf("Cannot open environment: %s", err)
	}

	return env
}

func clean(env *Env, t *testing.T) {
	path, err := env.Path()
	if err != nil {
		t.Errorf("Cannot get path")
	}
	if path == "" {
		t.Errorf("Invalid path")
	}
	t.Logf("Env path: %s", path)
	err = env.Close()
	if err != nil {
		t.Errorf("Error during close of environment: %s", err)
	}
	// clean up
	os.RemoveAll(path)
}

func TestEnvCopy(t *testing.T) {
	env := setup(t)
	clean(env, t)
}
