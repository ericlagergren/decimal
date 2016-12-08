// +build ignore

package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ericlagergren/decimal/suite"
)

func main() {
	const dir = "tests"
	infos, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatalln(err)
	}

	for _, info := range infos {
		if strings.HasSuffix(info.Name(), ".json") {
			continue
		}
		func() {
			name := info.Name()
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
			enc.SetIndent("  ", "    ")
			err = enc.Encode(cases)
			if err != nil {
				log.Fatalln(err)
			}
		}()
	}
}
