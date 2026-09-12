#!/bin/bash
set -euo pipefail

DRY_RUN=false
if [[ "${1:-}" == "--dry-run" ]]; then
  DRY_RUN=true
fi

echo "Fetching and pruning remote refs..."
git fetch --prune

BRANCHES=$(git branch -vv | grep ': gone]' | grep -v '^\*' | awk '{print $1}')

if [[ -z "$BRANCHES" ]]; then
  echo "No branches to delete."
  exit 0
fi

echo ""
echo "Branches to delete:"
echo "$BRANCHES" | sed 's/^/  /'
echo ""

if $DRY_RUN; then
  echo "[dry-run] No branches deleted."
  exit 0
fi

read -r -p "Delete these branches? [y/N] " REPLY
if [[ ! "$REPLY" =~ ^[Yy]$ ]]; then
  echo "Aborted."
  exit 0
fi

echo "$BRANCHES" | xargs git branch -D
echo "Done."
