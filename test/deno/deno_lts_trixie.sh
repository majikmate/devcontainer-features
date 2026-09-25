#!/bin/bash
set -e

source dev-container-features-test-lib

LTS="$(curl -fsSL https://dl.deno.land/release-lts-latest.txt || curl -fsSL https://dl.deno.land/release-latest.txt)"

check "deno is on the PATH" bash -c "command -v deno"
check "deno runs" deno --version
check "deno is the newest LTS release (${LTS})" bash -c "deno --version | head -n 1 | grep -F \"deno ${LTS#v} \""
check "deno executes TypeScript" bash -c "echo 'const n: number = 1 + 1; console.log(n)' > /tmp/t.ts && deno run /tmp/t.ts | grep -x 2"

reportResults
