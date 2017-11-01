#!/usr/bin/env bash

set -euo pipefail

# Simple basttery of tests for sanity checking changes.
go test -tags=ddebug -v -run='Abs|Add|Cmp|Mul|Quantize|Quo|String|Sub'
