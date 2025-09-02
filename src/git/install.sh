#!/bin/sh
set -e

echo "Activating feature 'git'"
echo "============================"
echo "pull-rebase: ${PULL_REBASE}"
echo "rebase-autostash: ${REBASE_AUTOSTASH}"

git config --system pull.rebase ${PULL_REBASE}
git config --system rebase.autoStash ${REBASE_AUTOSTASH}
git config --system --list