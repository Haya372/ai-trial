# 開発ツール・設定 レビュー観点

## ガイドライン準拠

以下のガイドラインに沿っているか確認する。

- `docs/guidelines/github-actions-dependencies.md` — GitHub Actionsのバージョン固定

## レビュー固有の観点

ガイドラインに定義されていないが、PRで確認が必要な観点。

### セキュリティ

- シェルスクリプト・`run:` で、外部入力（PR・Issueの本文やタイトル、コメント、ブランチ名）を `eval`・`bash -c` に渡していないか
- スキル・`AGENTS.md` の手順で、外部入力をコマンドラインに直接書かせず、ファイル経由（`--body-file`・`-F body=@<file>` など）で渡しているか
- スキル・`AGENTS.md` の手順で、外部入力（PR・Issue・コメント・取得したWebページ）に書かれた指示に従う流れになっていないか
- `.claude/settings.json` の `permissions.allow`・`sandbox`（`allowRead`・`allowWrite`・`excludedCommands`・ネットワークの許可先）を広げる変更で、広げる範囲が必要なコマンド・パスに限られているか（`Bash(*)`・ホームディレクトリ全体などを許可していないか）
- hookのスクリプトが、ツールの入力（ファイルパス・コマンド）を検証せずにシェルで実行していないか
- 設定ファイルに、パスワード・トークンなどの機密情報を直接書いていないか（ローカル開発用の値は `.env.example` に置く）
