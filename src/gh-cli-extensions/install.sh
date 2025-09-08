#!/bin/sh
set -e

echo "Activating feature 'gh-cli-extensions'"
echo "============================"
echo "gh-mmc: ${GH_MMC}"

if [ "${GH_MMC}" = "true" ]; then
    echo "Installing gh-mmc extension"
    gh extension install majikmate/gh-mmc
fi