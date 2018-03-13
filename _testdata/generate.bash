#!/usr/bin/env bash

set -euo pipefail

if [ "${TRAVIS_EVENT_TYPE}" == "cron" ]; then
	pushd _testdata
	time ./tables.py 500
	popd
fi
