#!/bin/sh
set -e

echo "Activating feature 'git'"
echo "============================"
echo "pull-rebase: ${PULL_REBASE}"
echo "rebase-autostash: ${REBASE_AUTOSTASH}"

git config --global pull.rebase ${PULL_REBASE}
git config --global rebase.autoStash ${REBASE_AUTOSTASH}
