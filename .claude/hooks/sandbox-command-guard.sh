#!/bin/sh
# PreToolUse(Bash): sandbox環境で失敗しやすいコマンドの書き方をブロックし、代替手段に誘導する。

CMD=$(jq -r '.tool_input.command // empty')

deny() {
  jq -n --arg reason "$1" '{hookSpecificOutput:{hookEventName:"PreToolUse",permissionDecision:"deny",permissionDecisionReason:$reason}}'
  exit 0
}

# 1. heredoc×コマンド置換（例: git commit -m "$(cat <<EOF ... EOF)"）
# heredocは改行なしでは書けないため、1行のgrep/rg/ag検索はパターン文字列を含んでいても通す。
# 2行目以降や && で連結された後続コマンドを除外対象にしないよう、単一行のコマンドに限定する。
if ! { [ "$(printf '%s' "$CMD" | wc -l)" -eq 0 ] && printf '%s' "$CMD" | grep -qE '^[[:space:]]*(grep|rg|ag)[[:space:]]'; } &&
  printf '%s' "$CMD" | grep -qE '\$\(\s*cat\s*<<'; then
  # shellcheck disable=SC2016 # 例示のため $(...) を展開せずに表示する
  deny 'heredocとcommand substitutionの組み合わせ（例: git commit -m "$(cat <<EOF ... EOF)" / gh pr create --body "$(cat <<EOF ...)"）はsandbox環境によって不安定になることがあります。$TMPDIR 配下に一時ファイルを書いて、git commit -F <file> や gh pr create --body-file <file> / gh pr comment --body-file <file> を使ってください。'
fi

# 2. excludedCommandsの対象コマンドを他のコマンドと連結する
# sandbox.excludedCommands はコマンド単独で実行したときだけ効き、パイプや && で連結すると
# コマンド全体がsandbox内で実行される。ネットワークを使うgit操作とghは認証・TLSで失敗するため単独実行させる。
# 引用符内の | や ; 、2>&1・&> 等のリダイレクトは連結とみなさない。
# 引用符は左から1パスで除去する（"it's" のように別種の引用符を含む文字列で除去範囲がずれないように）。
# git/gh はコマンド位置（先頭・区切り文字の直後、環境変数代入の後）にあるときだけ対象とし、
# grep gh のような引数としての出現は対象外とする。
if printf '%s' "$CMD" | perl -0ne '
  s/'\''[^'\'']*'\''|"(?:\\.|[^"\\])*"//g;
  s/[0-9]*[<>]&[0-9-]+|&>>?//g;
  my $target = qr/(?:^|[|;&(\n])\s*(?:\w+=\S*\s+)*(?:git\s+(?:push|fetch|pull|ls-remote)|gh)(?:\s|$)/;
  exit((/$target/ && /[|;&\n]/) ? 0 : 1);
'; then
  deny 'git push/fetch/pull/ls-remote と gh は sandbox.excludedCommands の対象ですが、パイプ・&&・;・改行で他のコマンドと連結するとコマンド全体がsandbox内で実行され、認証やTLSで失敗します。他のコマンドと連結せず単独で実行してください（ghの出力整形は --jq を使ってください）。'
fi
