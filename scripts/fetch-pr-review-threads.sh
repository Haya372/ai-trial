#!/bin/bash
# PRの既存レビュー・レビュースレッド・通常コメントをまとめてJSONで出力する。
# Usage: bash scripts/fetch-pr-review-threads.sh <PR番号>
set -euo pipefail

if [[ $# -ne 1 || ! "$1" =~ ^[0-9]+$ ]]; then
  echo "Usage: $0 <PR number>" >&2
  exit 1
fi

PR="$1"
REPO=$(gh repo view --json nameWithOwner --jq '.nameWithOwner')
OWNER="${REPO%/*}"
NAME="${REPO#*/}"

# --paginate と --slurp でページをまたいで全件を1つの配列にまとめる
REVIEWS=$(gh api "repos/$REPO/pulls/$PR/reviews" --paginate --slurp)
ISSUE_COMMENTS=$(gh api "repos/$REPO/issues/$PR/comments" --paginate --slurp)

# スレッドの解決状態はRESTでは取得できないため、GraphQLの reviewThreads を使う
# shellcheck disable=SC2016 # GraphQLの変数をシェルで展開しない
THREADS=$(gh api graphql --paginate --slurp \
  -F owner="$OWNER" -F name="$NAME" -F number="$PR" \
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
              comments(first: 100) {
                nodes { databaseId author { login } body createdAt url }
              }
            }
          }
        }
      }
    }')

jq -n \
  --argjson pr "$PR" \
  --argjson reviews "$REVIEWS" \
  --argjson issue_comments "$ISSUE_COMMENTS" \
  --argjson threads "$THREADS" \
  '
  ($reviews | add // []) as $reviews
  | {
      pr: $pr,
      # スレッドへの返信も本文が空のレビューとして記録されるため、本文のあるレビューを直前のレビューとみなす
      last_review_commit_id: ([$reviews[] | select((.body // "") != "")] | last | .commit_id // null),
      reviews: [$reviews[] | select((.body // "") != "") | {
        id, author: .user.login, state, body, commit_id, submitted_at
      }],
      issue_comments: [($issue_comments | add // [])[] | {
        id, author: .user.login, body, created_at: .created_at, url: .html_url
      }],
      threads: [$threads[].data.repository.pullRequest.reviewThreads.nodes[] | {
        is_resolved: .isResolved,
        is_outdated: .isOutdated,
        path,
        line,
        comments: [.comments.nodes[] | {
          id: .databaseId, author: .author.login, body, created_at: .createdAt, url
        }]
      }]
    }
  '
