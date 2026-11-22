# fetch-pr-review-threads.sh が取得したAPIのレスポンスを整形し、返信が必要なコメントを判定する。
# 入力: $pr（PR番号）, $author（PRの作成者）, $reviews・$issue_comments・$threads（gh api --paginate --slurp の出力）

# pr-reviewスキルの投稿は本文の先頭に印を付ける。作成者とpr-reviewの投稿者が同じアカウントの場合があるため、印で区別する
def is_pr_review_post: (.body // "") | startswith("<!-- pr-review -->");
def is_author_reply: .author == $author and (is_pr_review_post | not);

($reviews[0] | add // []) as $reviews
| [($issue_comments[0] | add // [])[] | {
    id, author: .user.login, body, created_at: .created_at, url: .html_url
  }] as $issue_comments
| [$reviews[] | select((.body // "") != "") | {
    id, author: .user.login, state, body, commit_id, submitted_at
  }] as $review_bodies
# レビュー本文・通常コメントはスレッドを持たないため、作成者の最後の返信（通常コメントまたはレビュー本文）より後のものを未返信とする
| ([($issue_comments[] | select(is_author_reply) | .created_at),
    ($review_bodies[] | select(is_author_reply) | .submitted_at)] | max) as $last_author_reply_at
| def needs_reply($at): (is_author_reply | not) and (is_pr_review_post | not)
    and ($last_author_reply_at == null or $at > $last_author_reply_at);
{
  pr: $pr,
  author: $author,
  # 人のレビューやスレッドへの返信（本文が空のレビュー）を起点にしないよう、pr-reviewスキルが投稿したレビューに限る
  last_review_commit_id: ([$reviews[] | select(.state != "PENDING" and is_pr_review_post)]
    | last | .commit_id // null),
  reviews: [$review_bodies[] | . + {needs_reply: needs_reply(.submitted_at)}],
  issue_comments: [$issue_comments[] | . + {needs_reply: needs_reply(.created_at)}],
  threads: [$threads[0][].data.repository.pullRequest.reviewThreads.nodes[]
    | [.comments.nodes[] | {
        id: .databaseId, author: .author.login, body, created_at: .createdAt, url,
        original_commit_id: .originalCommit.oid
      }] as $comments
    # 最初のコメントは指摘そのもの。pr-reviewはインラインの指摘に印を付けないため、2件目以降から返信を探す
    | ([$comments | to_entries[] | select(.key > 0 and (.value | is_author_reply)) | .key] | last) as $last_author_reply
    # 作成者の最後の返信より後のコメント。作成者の返信がなければすべてのコメント
    | (if $last_author_reply == null then $comments else $comments[$last_author_reply + 1:] end) as $unanswered
    | {
        is_resolved: .isResolved,
        is_outdated: .isOutdated,
        path,
        line,
        start_line: .startLine,
        # outdatedのスレッドは line が null になるため、コメントした時点の行番号も出す
        original_line: .originalLine,
        original_start_line: .originalStartLine,
        # LEFT（削除した行）の場合、行番号はベース側のファイルの行を指す
        diff_side: .diffSide,
        comments: $comments,
        needs_reply: ($unanswered | length > 0),
        unanswered_comments: $unanswered,
        # 前回のpr-reviewの返信に作成者がまだ答えていない
        last_comment_is_pr_review_reply: (($comments | length) > 1 and ($comments | last | is_pr_review_post))
      }]
}
