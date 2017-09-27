// +build ignore

package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ericlagergren/decimal/suite"
)

var indent = flag.Bool("indent", false, "output indented JSON")

func main() {
	flag.Parse()

	const dir = "_testdata"
	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatalln(err)
	}

	for _, info := range infos {
		name := info.Name()
		if !strings.HasSuffix(name, ".fptest") {
			continue
		}
		func() {
			file, err := os.Open(filepath.Join(dir, name))
			if err != nil {
				log.Fatalln(err)
			}
			defer file.Close()

			cases, err := suite.ParseCases(file)
			if err != nil {
				log.Fatalln(err)
			}

			name = strings.TrimSuffix(name, ".fptest")
			out, err := os.Create(filepath.Join(dir, name+".json"))
			if err != nil {
				log.Fatalln(err)
			}
			defer out.Close()

			enc := json.NewEncoder(out)
			if *indent {
				enc.SetIndent("  ", "    ")
			}
			err = enc.Encode(cases)
			if err != nil {
				log.Fatalln(err)
			}
		}()
	}
}
