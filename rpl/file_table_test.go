package rpl

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestFileTable(t *testing.T) {
	base, err := ioutil.TempDir("./", "test_table")
	if err != nil {
		t.Fatal(err)
	}

	defer os.RemoveAll(base)

}
