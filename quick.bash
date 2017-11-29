#!/usr/bin/env bash

set -euo pipefail

r=${1:-.}

# Simple battery of tests for sanity checking changes.
# Timeout is set in case N in _testdata/tables.py is set high.
go test -timeout=12h -tags=ddebug -short -v -run="${r}"
