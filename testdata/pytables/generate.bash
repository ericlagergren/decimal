#!/usr/bin/env bash

set -xeuo pipefail

if [ "${TRAVIS_EVENT_TYPE}" == "cron" ] || ! ls testdata/pytables/*.gz 1> /dev/null 2>&1; then
	pushd testdata/pytables
	time ./tables.py 500
	popd
fi