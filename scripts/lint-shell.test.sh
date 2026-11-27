#!/bin/bash
# lint-shell.sh の対象ファイル判定と終了コードを一時リポジトリで検証する。
# Usage: bash scripts/lint-shell.test.sh
set -euo pipefail

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
LINT_SHELL="$SCRIPT_DIR/lint-shell.sh"
WORK_DIR=$(mktemp -d "${TMPDIR:-/tmp}/lint-shell-test.XXXXXX")
trap 'rm -rf "$WORK_DIR"' EXIT

failures=0
assert_eq() {
  # $1: テスト名, $2: 期待値, $3: 実際の値
  if [[ "$2" == "$3" ]]; then
    echo "ok: $1"
  else
    echo "NG: $1"
    echo "  expected: $(printf '%q' "$2")"
    echo "  actual:   $(printf '%q' "$3")"
    failures=$((failures + 1))
  fi
}

cd "$WORK_DIR"
git init -q
mkdir -p scripts hooks docs
printf '#!/bin/bash\necho ok\n' > scripts/a.sh
printf 'echo no-shebang\n' > scripts/b.bash
printf '#!/usr/bin/env bash\necho ok\n' > hooks/env-bash
printf '#!/bin/sh\necho ok\n' > hooks/posix-sh
printf '#!/usr/bin/env python3\nprint("ok")\n' > scripts/tool
printf '# doc\n' > docs/README.md
printf '#!/bin/bash\necho untracked\n' > untracked.sh
git add scripts hooks docs

expected_list=$(printf '%s\n' hooks/env-bash hooks/posix-sh scripts/a.sh scripts/b.bash)
assert_eq "--list は git管理下の .sh/.bash と sh/bash の shebang を持つファイルだけを出力する" \
  "$expected_list" "$(bash "$LINT_SHELL" --list)"

mkdir -p sub
assert_eq "--list はサブディレクトリから実行してもリポジトリルートからのパスを出力する" \
  "$expected_list" "$(cd sub && bash "$LINT_SHELL" --list)"

status=0
bash "$LINT_SHELL" > /dev/null 2>&1 || status=$?
assert_eq "指摘がなければ終了コード0" "0" "$status"

# SC2086（info レベル）も検出する
# shellcheck disable=SC2016 # $1 は展開せずにそのまま書き込む
printf '#!/bin/bash\nrm $1\n' > hooks/env-bash
status=0
output=$(bash "$LINT_SHELL" 2>&1) || status=$?
assert_eq "info レベルの指摘があれば終了コード1" "1" "$status"
contains=no
grep -q 'SC2086' <<< "$output" && contains=yes
assert_eq "指摘内容に SC2086 が含まれる" "yes" "$contains"

if (( failures > 0 )); then
  echo "$failures test(s) failed"
  exit 1
fi
echo "all tests passed"
