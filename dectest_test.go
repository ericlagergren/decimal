package decimal_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ericlagergren/decimal/v4/dectest"
)

// TestDecTests runs the dectest test suite.
func TestDecTests(t *testing.T) {
	path := filepath.Join("testdata", "dectest")
	files, err := filepath.Glob(filepath.Join(path, "*.decTest"))
	if err != nil {
		t.Fatal(err)
	}

	if len(files) == 0 {
		t.Fatalf("no .detect files found inside %[1]q, re-run %[1]s%cgenerate.bash",
			path, os.PathSeparator)
	}

	for _, file := range files {
		file := file
		t.Run(filepath.Base(file), func(t *testing.T) {
			dectest.Test(t, file)
		})
	}
}
