package decimal_test

import (
	"path/filepath"
	"testing"

	"github.com/ericlagergren/decimal/dectest"
)

// TestDecTests runs the dectest test suite.
func TestDecTests(t *testing.T) {
	path := filepath.Join("testdata", "dectest")
	files, err := filepath.Glob(filepath.Join(path, "*.decTest"))
	if err != nil {
		t.Fatal(err)
	}

	if len(files) == 0 {
		t.Fatalf("no .decTest files found in %q; run %q",
			path, filepath.Join(path, "generate.bash"))
	}

	for _, file := range files {
		t.Run(filepath.Base(file), func(t *testing.T) {
			dectest.Test(t, file)
		})
	}
}
