#!/bin/bash
# git管理下の手書きシェルスクリプトに shellcheck を実行する。
# lefthook・CIの両方から呼び出し、対象ファイルの定義はこのスクリプトだけに置く。
# 対象: 拡張子が .sh/.bash のファイルと、shebang が sh/bash のファイル
# Usage: bash scripts/lint-shell.sh [--list]
#   --list  shellcheck を実行せず、対象ファイルを出力する
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

list_shell_scripts() {
  local file first_line
  while IFS= read -r -d '' file; do
    [[ -f "$file" ]] || continue
    case "$file" in
      *.sh | *.bash)
        echo "$file"
        continue
        ;;
    esac
    IFS= read -r first_line < "$file" || true
    if [[ "$first_line" =~ ^#!.*[/[:space:]](ba)?sh([[:space:]]|$) ]]; then
      echo "$file"
    fi
  done < <(git ls-files -z)
}

files=()
while IFS= read -r file; do
  files+=("$file")
done < <(list_shell_scripts | LC_ALL=C sort)

if [[ "${1:-}" == "--list" ]]; then
  if (( ${#files[@]} > 0 )); then
    printf '%s\n' "${files[@]}"
  fi
  exit 0
fi

if (( ${#files[@]} == 0 )); then
  exit 0
fi

shellcheck --severity=info "${files[@]}"
