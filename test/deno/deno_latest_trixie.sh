#!/bin/bash
set -e

source dev-container-features-test-lib

LATEST="$(curl -fsSL https://dl.deno.land/release-latest.txt)"

check "deno is the newest release (${LATEST})" bash -c "deno --version | head -n 1 | grep -F \"deno ${LATEST#v} \""

reportResults
