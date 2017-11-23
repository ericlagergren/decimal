// +build !go1.10

package compat

import "bytes"

type Builder struct {
	bytes.Buffer
}
