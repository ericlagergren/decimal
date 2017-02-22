#!/usr/bin/env bash

set -euo pipefail

ragel -e -p -Z -G2 -o parser.go parser.rl
#ragel -e -p -V -Z -G2 -o parser.dot parser.rl
#dot -Tsvg -o parser.svg parser.dot
