#!/bin/bash
# PRの既存レビュー・レビュースレッド・通常コメントと、PRの作成者が返信していないコメントの判定をまとめてJSONで出力する。
# Usage: bash scripts/fetch-pr-review-threads.sh <PR番号>
set -euo pipefail

if [[ $# -ne 1 || ! "$1" =~ ^[0-9]+$ ]]; then
  echo "Usage: $0 <PR number>" >&2
  exit 1
fi

PR="$1"
SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)

# レスポンスが大きいとコマンドライン引数の長さの上限を超えるため、ファイル経由でjqに渡す
# macOSの mktemp -d はテンプレートを省略すると $TMPDIR を使わないため、明示する
WORK_DIR=$(mktemp -d "${TMPDIR:-/tmp}/fetch-pr-review-threads.XXXXXX")
trap 'rm -rf "$WORK_DIR"' EXIT

# {owner}/{repo} はgh apiがカレントディレクトリのリポジトリで置き換える
AUTHOR=$(gh api "repos/{owner}/{repo}/pulls/$PR" --jq .user.login)
# --paginate と --slurp でページをまたいで全件を1つの配列にまとめる
gh api "repos/{owner}/{repo}/pulls/$PR/reviews" --paginate --slurp > "$WORK_DIR/reviews.json"
gh api "repos/{owner}/{repo}/issues/$PR/comments" --paginate --slurp > "$WORK_DIR/issue_comments.json"

# スレッドの解決状態はRESTでは取得できないため、GraphQLの reviewThreads を使う
# プレースホルダは -f では置き換えられないため -F で渡す（数値変換より後に置き換えるので文字列のまま渡る）
# shellcheck disable=SC2016 # GraphQLの変数をシェルで展開しない
gh api graphql --paginate --slurp \
  -F owner='{owner}' -F name='{repo}' -F number="$PR" \
  -f query='
    query($owner: String!, $name: String!, $number: Int!, $endCursor: String) {
      repository(owner: $owner, name: $name) {
        pullRequest(number: $number) {
          reviewThreads(first: 100, after: $endCursor) {
            pageInfo { hasNextPage endCursor }
            nodes {
              isResolved
              isOutdated
              path
              line
              startLine
              originalLine
              originalStartLine
              diffSide
              comments(first: 100) {
                nodes { databaseId author { login } body createdAt url originalCommit { oid } }
              }
            }
          }
        }
      }
    }' > "$WORK_DIR/threads.json"

jq -n \
  --argjson pr "$PR" \
  --arg author "$AUTHOR" \
  --slurpfile reviews "$WORK_DIR/reviews.json" \
  --slurpfile issue_comments "$WORK_DIR/issue_comments.json" \
  --slurpfile threads "$WORK_DIR/threads.json" \
  -f "$SCRIPT_DIR/fetch-pr-review-threads.jq"
