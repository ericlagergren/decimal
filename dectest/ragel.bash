#!/usr/bin/env bash

set -euo pipefail

ragel -e -p -Z -G2 -o scanner.go scanner.rl
