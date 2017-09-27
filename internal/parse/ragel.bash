#!/usr/bin/env bash

set -euo pipefail

ragel -e -p -Z -G2 -o number.go number.rl
#ragel -e -p -V -Z -G2 -o number.dot number.rl
#dot -Tsvg -o number.svg number.dot
