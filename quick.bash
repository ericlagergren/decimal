#!/usr/bin/env bash

set -euo pipefail

dtests='Abs|Add|Cmp|Example|FMA|Mul|Quantize|Quo|Rat|String|Sign|Signbit|Sub'

# Simple battery of tests for sanity checking changes.
# Timeout is set in case N in _testdata/tables.py is set high.
go test -timeout=12h -race -tags=ddebug -run=$dtests -v
