#!/bin/sh
# PreToolUse(Bash): heredoc×コマンド置換（例: git commit -m "$(cat <<EOF ... EOF)"）をブロックし、
# $TMPDIR 配下の一時ファイル + git commit -F / gh pr create --body-file に誘導する。

CMD=$(jq -r '.tool_input.command // empty')

# heredocは改行なしでは書けないため、1行のgrep/rg/ag検索はパターン文字列を含んでいても通す。
# 2行目以降や && で連結された後続コマンドを除外対象にしないよう、単一行のコマンドに限定する。
if [ "$(printf '%s' "$CMD" | wc -l)" -eq 0 ] && printf '%s' "$CMD" | grep -qE '^[[:space:]]*(grep|rg|ag)[[:space:]]'; then
  exit 0
fi

if printf '%s' "$CMD" | grep -qE '\$\(\s*cat\s*<<'; then
  jq -n '{hookSpecificOutput:{hookEventName:"PreToolUse",permissionDecision:"deny",permissionDecisionReason:"heredocとcommand substitutionの組み合わせ（例: git commit -m \"$(cat <<EOF ... EOF)\" / gh pr create --body \"$(cat <<EOF ...)\"）はsandbox環境によって不安定になることがあります。$TMPDIR 配下に一時ファイルを書いて、git commit -F <file> や gh pr create --body-file <file> / gh pr comment --body-file <file> を使ってください。"}}'
fi
