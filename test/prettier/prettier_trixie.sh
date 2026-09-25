#!/bin/bash
set -e

source dev-container-features-test-lib

WORK="$(mktemp -d)"
printf '<div class="text-center p-4 flex card"></div>\n' > "${WORK}/index.html"
printf 'const a = {b:1}\n' > "${WORK}/a.js"

check "prettier is installed in /usr/local" test -x /usr/local/bin/prettier
check "prettier runs" prettier --version
check "global configuration exists" test -f /.prettierrc.json
check "global configuration contains only the plugin" bash -c "grep -q prettier-plugin-tailwindcss /.prettierrc.json && ! grep -q -E 'tabWidth|semi|singleQuote' /.prettierrc.json"
check "tailwind classes are sorted" bash -c "prettier '${WORK}/index.html' | grep -F 'class=\"card flex p-4 text-center\"'"
check "standard prettier style is used" bash -c "prettier '${WORK}/a.js' | grep -Fx 'const a = { b: 1 };'"

reportResults
