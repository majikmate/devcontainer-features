#!/bin/bash
set -e

source dev-container-features-test-lib

check "prettier runs" prettier --version
check "no global configuration" bash -c "test ! -e /.prettierrc.json"

reportResults
