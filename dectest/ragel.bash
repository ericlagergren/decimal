#!/usr/bin/env bash

set -xeuo pipefail

ragel -e -p -Z -G2 -o scanner.go scanner.rl
gofmt -s -w scanner.go
