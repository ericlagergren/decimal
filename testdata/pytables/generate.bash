#!/usr/bin/env bash

set -xeuo pipefail

if [ "${TRAVIS_EVENT_TYPE}" == "cron" ]; then
	pushd testdata/pytables
	time ./tables.py 500
	popd
fi
