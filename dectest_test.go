package decimal_test

import (
	"path/filepath"
	"testing"

	"github.com/ericlagergren/decimal/dectest"
)

func TestDecTests(t *testing.T) {
	files, err := filepath.Glob("_dectest/*.decTest")
	if err != nil {
		t.Fatal(err)
	}

	if len(files) == 0 {
		t.Skip("No .detect files found, please run _dectests/generate.bash")
	}

	for _, file := range files {
		file := file // shadow range variable
		t.Run(filepath.Base(file), func(t *testing.T) {
			dectest.Test(t, file)
		})
	}
}
