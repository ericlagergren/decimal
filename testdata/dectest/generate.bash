#!/usr/bin/env bash

set -xeuo pipefail

curl -sSOL http://speleotrove.com/decimal/dectest.zip
unzip -o dectest.zip
rm dectest.zip
