#!/bin/bash
# fetch-pr-review-threads.jq の返信判定をfixtureで検証する。
# Usage: bash scripts/fetch-pr-review-threads.test.sh
set -euo pipefail

SCRIPT_DIR=$(cd "$(dirname "$0")" && pwd)
WORK_DIR=$(mktemp -d "${TMPDIR:-/tmp}/fetch-pr-review-threads-test.XXXXXX")
trap 'rm -rf "$WORK_DIR"' EXIT

# gh api --paginate --slurp の出力と同じく、ページの配列で渡す
cat > "$WORK_DIR/reviews.json" <<'EOF'
[[
  {"id": 1, "user": {"login": "author"}, "state": "COMMENTED", "body": "<!-- pr-review -->\n指摘2件", "commit_id": "c1", "submitted_at": "2026-01-01T00:00:00Z"},
  {"id": 2, "user": {"login": "human"}, "state": "COMMENTED", "body": "全体的に命名を見直してください", "commit_id": "c1", "submitted_at": "2026-01-01T01:00:00Z"},
  {"id": 3, "user": {"login": "human"}, "state": "COMMENTED", "body": "", "commit_id": "c1", "submitted_at": "2026-01-01T02:00:00Z"},
  {"id": 4, "user": {"login": "human"}, "state": "COMMENTED", "body": "テストも追加してください", "commit_id": "c2", "submitted_at": "2026-01-01T04:00:00Z"}
]]
EOF

cat > "$WORK_DIR/issue_comments.json" <<'EOF'
[[
  {"id": 10, "user": {"login": "human"}, "body": "CIが落ちています", "created_at": "2026-01-01T00:30:00Z", "html_url": "u10"},
  {"id": 11, "user": {"login": "author"}, "body": "対応しました", "created_at": "2026-01-01T03:00:00Z", "html_url": "u11"},
  {"id": 12, "user": {"login": "human"}, "body": "まだ落ちています", "created_at": "2026-01-01T05:00:00Z", "html_url": "u12"}
]]
EOF

thread() {
  # $1: スレッドのコメントの配列（author と body だけ）
  jq -c '{
    isResolved: false, isOutdated: false, path: "a.md", line: 1, startLine: null,
    originalLine: 1, originalStartLine: null, diffSide: "RIGHT",
    comments: {nodes: [to_entries[] | {
      databaseId: (.key + 100), author: {login: .value[0]}, body: .value[1],
      createdAt: "2026-01-01T00:00:00Z", url: "t", originalCommit: {oid: "c1"}
    }]}
  }' <<< "$1"
}

{
  echo '[{"data": {"repository": {"pullRequest": {"reviewThreads": {"nodes": ['
  # 0: 作成者の返信がない
  thread '[["author", "<!-- pr-review -->\n指摘A"]]'
  echo ','
  # 1: 作成者が最後に返信した
  thread '[["author", "<!-- pr-review -->\n指摘B"], ["author", "対応しました"]]'
  echo ','
  # 2: 作成者の返信の後に、pr-reviewが再び返信した
  thread '[["author", "<!-- pr-review -->\n指摘C"], ["author", "対応しました"], ["author", "<!-- pr-review -->\nまだ残っています"]]'
  echo ','
  # 3: 作成者の返信の後に、人が返信した
  thread '[["human", "指摘D"], ["author", "対応しました"], ["human", "了解です"], ["human", "もう1点あります"]]'
  echo ','
  # 4: 印が本文の途中にある作成者のコメントは返信とみなす
  thread '[["human", "指摘E"], ["author", "前回の <!-- pr-review --> の件、対応しました"]]'
  echo ','
  # 5: pr-reviewのインラインの指摘は印を付けないため、作成者が投稿した印のない1件目も指摘とみなす
  thread '[["author", "指摘F"]]'
  echo ','
  # 6: 印のない1件目に作成者が返信した
  thread '[["author", "指摘G"], ["author", "対応しました"]]'
  echo ','
  # 7: 退会したユーザーのコメントは、GraphQLで author が null になる
  thread '[["human", "指摘H"], [null, "削除されたユーザーのコメント"]]'
  echo ']}}}}}]'
} > "$WORK_DIR/threads.json"

OUTPUT=$(jq -n \
  --argjson pr 1 \
  --arg author author \
  --slurpfile reviews "$WORK_DIR/reviews.json" \
  --slurpfile issue_comments "$WORK_DIR/issue_comments.json" \
  --slurpfile threads "$WORK_DIR/threads.json" \
  -f "$SCRIPT_DIR/fetch-pr-review-threads.jq")

FAILED=0
assert() {
  # $1: 説明, $2: 出力に対するjqの式, $3: 期待する値（JSON）
  local actual
  actual=$(jq -c "$2" <<< "$OUTPUT")
  if [[ "$actual" != "$3" ]]; then
    echo "FAIL: $1: expected $3, got $actual" >&2
    FAILED=1
  fi
}

assert "作成者を出力する" '.author' '"author"'
assert "pr-reviewのレビューを起点にする" '.last_review_commit_id' '"c1"'

assert "返信がないスレッドは全コメントが未返信" '.threads[0] | [.needs_reply, [.unanswered_comments[].id]]' '[true,[100]]'
assert "作成者が最後に返信したスレッドは返信不要" '.threads[1] | [.needs_reply, .unanswered_comments]' '[false,[]]'
assert "作成者の返信より後のpr-reviewの返信は未返信" '.threads[2] | [.needs_reply, [.unanswered_comments[].id]]' '[true,[102]]'
assert "作成者の返信より後のコメントだけが未返信" '.threads[3] | [.needs_reply, [.unanswered_comments[].id]]' '[true,[102,103]]'
assert "本文の途中の印は作成者の返信とみなす" '.threads[4].needs_reply' 'false'
assert "作成者が投稿した印のない1件目は指摘とみなす" '.threads[5] | [.needs_reply, [.unanswered_comments[].id]]' '[true,[100]]'
assert "印のない1件目への作成者の返信は返信とみなす" '.threads[6] | [.needs_reply, .unanswered_comments]' '[false,[]]'

assert "投稿者がnullのコメントは返信とみなさない" '.threads[7] | [.needs_reply, [.unanswered_comments[].id]]' '[true,[100,101]]'

assert "最後のコメントがpr-reviewの返信" '[.threads[].last_comment_is_pr_review_reply]' '[false,false,true,false,false,false,false,false]'

assert "本文が空のレビューは出力しない" '[.reviews[].id]' '[1,2,4]'
assert "pr-reviewのレビュー本文は返信不要" '.reviews[0].needs_reply' 'false'
assert "作成者の返信より前のレビュー本文は返信不要" '.reviews[1].needs_reply' 'false'
assert "作成者の返信より後のレビュー本文は未返信" '.reviews[2].needs_reply' 'true'

assert "作成者の返信より前の通常コメントは返信不要" '.issue_comments[0].needs_reply' 'false'
assert "作成者自身の通常コメントは返信不要" '.issue_comments[1].needs_reply' 'false'
assert "作成者の返信より後の通常コメントは未返信" '.issue_comments[2].needs_reply' 'true'

# botのログイン名は、RESTでは末尾に [bot] が付き、GraphQL（スレッドのコメント）では付かない
OUTPUT=$(jq -n \
  --argjson pr 1 \
  --arg author 'author[bot]' \
  --slurpfile reviews <(sed 's/"login": "author"/"login": "author[bot]"/' "$WORK_DIR/reviews.json") \
  --slurpfile issue_comments <(sed 's/"login": "author"/"login": "author[bot]"/' "$WORK_DIR/issue_comments.json") \
  --slurpfile threads "$WORK_DIR/threads.json" \
  -f "$SCRIPT_DIR/fetch-pr-review-threads.jq")

assert "botの作成者のスレッドへの返信は返信とみなす" '.threads[1].needs_reply' 'false'
assert "botの作成者の通常コメントは返信とみなす" '[.issue_comments[].needs_reply]' '[false,false,true]'

if [[ $FAILED -ne 0 ]]; then
  exit 1
fi
echo "ok"
