---
name: pr-review
description: PRをレビューしてインラインコメントを投稿するとき。「PRをレビューする」「PRレビューをする」のような発言や、開発フローの「7. PRレビュー」で呼び出す。
---

# PRレビュースキル

## 目的

PRの要件・仕様・既存のレビュースレッドと照らし合わせて指摘を絞り込み、インラインコメントとして投稿する。

---

## 手順

### ステップ 1: 文脈を集める

```bash
gh pr view <PR番号> --json number,title,body,author,headRefOid,headRefName,baseRefName,closingIssuesReferences
```

PR番号が会話から分からない場合は、`<PR番号>` を省略して現在のブランチのPRを取得する。

以下を読む。

- PRに紐づくIssue（`closingIssuesReferences`、またはPR本文の `#NN`）の要件・受け入れ条件。Issueが親Issueを持つ場合は親Issueも読む
- 該当する仕様書（`docs/spec/`）と設計ドキュメント（`docs/design/`）。Issue・PR本文に記載がない場合は、変更したファイルの機能名で `docs/spec/` を検索する
- 既存のレビューとスレッド

```bash
bash scripts/fetch-pr-review-threads.sh <PR番号> >| "$TMPDIR/pr-<PR番号>-threads.json"
```

出力の `threads[].is_resolved` でスレッドの解決状態、`threads[].comments` で返信の有無を確認する。

### ステップ 2: レビュー対象の差分を決める

ベースブランチとPRのコミットを取得する。ブランチは切り替えない（PRのブランチを別のworktreeで開いていても動くようにするため）。

```bash
git fetch origin <baseRefName> pull/<PR番号>/head
```

以降の手順では、作業ツリーやHEADではなく、ステップ1の `headRefOid` をPRの内容として扱う。ファイルは `git show <headRefOid>:<path>` で読み、呼び出し元は `git grep <pattern> <headRefOid>` で探す。

ステップ1の出力の `last_review_commit_id` で、上の行から順に当てはまる条件の対象を使う。

| 条件 | 対象 |
|---|---|
| `last_review_commit_id` が `null`（初回） | `git diff origin/<baseRefName>...<headRefOid>` |
| `git merge-base --is-ancestor <last_review_commit_id> <headRefOid>` が失敗する（リベース・force-push後） | `git diff origin/<baseRefName>...<headRefOid>` と、未解決のスレッド |
| `git log --merges <last_review_commit_id>..<headRefOid>` が1件以上出力する（ベースブランチをmerge済み） | `git diff origin/<baseRefName>...<headRefOid>` と、未解決のスレッド |
| 上のどれにも当たらない | `git diff <last_review_commit_id>..<headRefOid>` と、未解決のスレッド |

2回目以降で `last_review_commit_id` からの差分を対象にする場合も、変更した関数・コンポーネント・型の呼び出し元と依存先を `git grep` で探し、変更による影響（共有の状態・フック・インターフェースの変化）を確認する。

### ステップ 3: 観点ごとにレビューする

以下の観点で指摘を集める。

| 観点 | 確認方法 |
|---|---|
| 要件・スコープ | Issueの要件・受け入れ条件をすべて満たしているか。要件と無関係な変更が含まれていないか |
| 仕様 | 振る舞いが仕様書の記述に沿っているか |
| 正しさ | Claude Codeでは `code-review` スキルを `--comment` なしで、件数の上限を外して呼ぶ（例: `code-review high --max-findings all <PR番号>`）。上限があると、ステップ4で除く範囲外の指摘で枠が埋まり、対象の差分の指摘が切り捨てられる。それ以外のツールでは、対象の差分の変更ごとに、境界値・null・エラー時の分岐・並行実行で誤った結果になる入力がないかをコードを辿って確認する |
| セキュリティ | Claude Codeで、現在のブランチ（`git branch --show-current`）が `headRefName` と一致し、HEADが `headRefOid` と一致する場合は `/security-review` を実行する（`/security-review` は現在のブランチの差分をレビューする）。それ以外の場合は、対象の差分を `docs/guidelines/backend-review-criteria.md` の「セキュリティ」に沿って確認する |
| ガイドライン | 変更したファイルの種類に応じて参照する（下表） |
| テスト | `docs/guidelines/test-guidelines.md` に沿って、テストケースの網羅とテストの品質を確認する |

| 変更したファイル | 参照するレビュー観点 |
|---|---|
| `apps/backend/` | `docs/guidelines/backend-review-criteria.md` |
| `apps/web/`・`packages/ui/` | `docs/guidelines/frontend-review-criteria.md` |
| Markdownドキュメント・スキル | `docs/guidelines/document-review-criteria.md` |

### ステップ 4: 指摘を絞り込む

集めた指摘を1件ずつ、以下の表の上の行から順に当てはめて扱いを決める。

| 指摘の状態 | 扱い |
|---|---|
| 既存スレッドと同じ内容で、そのスレッドが解決済み | 除く |
| 既存スレッドと同じ内容で、そのスレッドに返信がある（対応済み・対応しない理由の説明済み） | 除く。返信の内容でまだ問題が残る場合だけ、そのスレッドへの返信として投稿する |
| 既存スレッドと同じ内容で、返信がなく未解決、かつ `is_outdated` が `true` | 指摘したコード（`git show <comments[0].original_commit_id>:<path>` の `original_line` 行目）が今どうなっているかを確認する。問題が残っていれば、そのスレッドへの返信として投稿する。残っていなければ除く |
| 既存スレッドと同じ内容で、返信がなく未解決、かつ `is_outdated` が `false` | 新しく投稿せず、ステップ6の報告で未対応のスレッドとして挙げる |
| ステップ2の対象の差分に含まれない行への指摘で、対象の差分の変更が原因ではない | 除く |
| 仕様書に定められた振る舞いを問題とする指摘 | 投稿せず、仕様への質問としてステップ6でユーザーに伝える |
| 上のどれにも当たらない | 投稿する |

既存スレッドとの照合では、`threads[].path` とファイルが一致するスレッドに加え、`reviews[].body`・`issue_comments[].body` の内容も確認する。

「投稿する」になった指摘は、該当するコードを読み直して確信度を判断する。

| 確信度 | 基準 | 扱い |
|---|---|---|
| 高 | 問題が起きる入力・状態と、そのときの誤った結果を、コードを辿って具体的に示せる | 投稿する |
| 中 | 問題が起きるかが、コードから読み取れない条件（実行環境・外部APIの挙動・作成者の意図）に依存する | 投稿せず、確認が必要な指摘としてステップ6でユーザーに伝える |
| 低 | 問題が起きる具体的な条件を示せない、または好みの違い | 除く |

### ステップ 5: 投稿する

指摘を1つのレビューとして投稿する。payloadをJSONファイルに書き、その絶対パスを `--input` に渡す。Claude Codeでは、Writeツールでスクラッチパッドディレクトリ（システムプロンプトに記載の絶対パス）に作成する。それ以外のツールでは `$TMPDIR` に作成する。

| 投稿する内容 | `comments` | `body` |
|---|---|---|
| 新しい指摘が1件以上 | 指摘を1件ずつインラインコメントにする | 確認した観点と指摘の件数 |
| 新しい指摘が0件で、既存スレッドへの返信が1件以上 | 空配列 | 確認した観点と、問題が残っているとして返信したスレッドの件数 |
| 新しい指摘も既存スレッドへの返信も0件 | 空配列 | 確認した観点を挙げ、「問題なし」と書く |

```json
{
  "commit_id": "<headRefOid>",
  "event": "COMMENT",
  "body": "<!-- pr-review -->\n<確認した観点と、指摘件数の要約>",
  "comments": [
    { "path": "<ファイルパス>", "line": 42, "side": "RIGHT", "body": "<指摘>" }
  ]
}
```

投稿の直前に `gh pr view <PR番号> --json headRefOid --jq .headRefOid` を実行する。ステップ1の `headRefOid` と異なる場合は、レビュー中にpushされているので、ステップ1からやり直す。

```bash
gh api repos/{owner}/{repo}/pulls/<PR番号>/reviews --method POST --input "<payloadの絶対パス>"
```

- 指摘・`body` は日本語で書く。指摘には、問題が起きる具体的な条件と期待する振る舞いを書く
- 複数行への指摘は `start_line` と `line` で範囲を指定する
- 指摘は `body` に書かず、すべてインラインコメントにする。差分に含まれない行への指摘（呼び出し元への影響など）は、原因になった変更行にコメントし、影響を受ける箇所をファイルパスと行番号で書く。PRで変更していないファイルだけに関わる指摘は、投稿せずステップ6でユーザーに伝える
- `body` の先頭に `<!-- pr-review -->` を書く。この印のあるレビューだけが、次回のレビューで `last_review_commit_id` の起点になる
- 既存スレッドへの返信は、本文をpayloadと同じ場所のファイルに書き、`gh api repos/{owner}/{repo}/pulls/<PR番号>/comments/<comment_id>/replies --method POST -F body=@<返信ファイルの絶対パス>` で投稿する（`comment_id` は `threads[].comments[0].id`）。本文をコマンドラインに書くと、バッククォートがシェルのコマンド置換として実行される

### ステップ 6: ユーザーに報告する

以下を報告する。

- 投稿した指摘の件数とレビューのURL
- 仕様への質問（ステップ4で投稿しなかったもの）
- 確信度が中の指摘と、PRで変更していないファイルだけに関わる指摘
- 返信がなく未解決のスレッド

---

## チェックリスト

- [ ] Issueの要件・受け入れ条件、仕様書、既存スレッドを読んだか
- [ ] 2回目以降は、差分の起点を `last_review_commit_id` から決めたか
- [ ] 変更したコードの呼び出し元・依存先を確認したか
- [ ] 正しさを確認したか（Claude Codeでは `code-review` を `--comment` なしで呼んだか）
- [ ] 返信済みスレッドと同じ指摘を除いたか
- [ ] 仕様どおりの振る舞いへの指摘を、仕様への質問としてユーザーに伝えたか
- [ ] 投稿する指摘の確信度を判断し、高のものだけを投稿したか
- [ ] 指摘をすべてインラインコメントにしたか
- [ ] 指摘が0件の場合に、確認した観点を添えて「問題なし」を投稿したか
- [ ] レビューの `body` の先頭に `<!-- pr-review -->` を書いたか
