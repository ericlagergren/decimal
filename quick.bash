#!/usr/bin/env bash

set -euo pipefail

dtests='Abs|Add|Cmp|Mul|Quantize|Quo|Rat|String|Sign|Signbit|Sub'

# Simple basttery of tests for sanity checking changes.
go test -tags=ddebug -v -run=$dtests
