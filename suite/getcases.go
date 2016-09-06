// +build ignore

package main

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

const prefix = "https://www.research.ibm.com/haifa/projects/verification/fpgen/download/Decimal-"

var urls = [...]string{
	prefix + "Basic-Types-Inputs.fptest",
	prefix + "Basic-Types-Intermediate.fptest",
	prefix + "Rounding.fptest",
	prefix + "Overflow.fptest",
	prefix + "Underflow.fptest",
	prefix + "Trailing-And-Leading-Zeros-Input.fptest",
	prefix + "Trailing-And-Leading-Zeros-Result.fptest",
	prefix + "Clamping.fptest",
	prefix + "Mul-Trailing-Zeros.fptest",
}

func main() {
	const dir = "tests"
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		log.Fatalln(err)
	}

	for _, url := range urls {
		resp, err := http.Get(url)
		defer resp.Body.Close()
		if err != nil {
			log.Fatalln(err)
		}
		name := strings.TrimPrefix(url, prefix)
		name = strings.Replace(name, "-", "", -1)
		out, err := os.Create(filepath.Join(dir, name))
		if err != nil {
			log.Fatalln(err)
		}
		defer out.Close()
		err = copyLines(out, resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
	}
}

// copyLines copies r to w but omits invalid lines.
func copyLines(w io.Writer, r io.Reader) (err error) {
	s := bufio.NewScanner(r)
	s.Split(scanLines)
	// wc -L *.fptest | sort shows max line length is 142. Alloc a few extra
	// bytes in case it grows in the future.
	s.Buffer(make([]byte, 150), bufio.MaxScanTokenSize)
	var p []byte
	for s.Scan() {
		p = s.Bytes()
		if badLine(p) {
			continue
		}
		_, err = w.Write(p)
		if err != nil {
			return err
		}
	}
	return s.Err()
}

// badLine returns true if the line is one of the non-test lines. (E.g., is
// a <pre> tags, a copyright date, etc.)
func badLine(p []byte) bool {
	if len(p) == 0 {
		return true
	}
	switch p[0] {
	case '\n', '\r':
		return true
	case '<':
		return bytes.Contains(p, []byte("pre>"))
	case 'D':
		return bytes.HasPrefix(p, []byte("Decimal floating point tests: "))
	case 'C':
		return bytes.HasPrefix(p, []byte("Copyright of IBM Corp. 20"))
	default:
		return false
	}
}

// scanLines is copied from the stdlib but includes newlines.
func scanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil

	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, convCR(data[0:i]), nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), convCR(data), nil
	}
	// Request more data.
	return 0, nil, nil
}

// convCR converts a trailing carriage return to a line feed.
func convCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		data[len(data)-1] = '\n'
	}
	return data
}
