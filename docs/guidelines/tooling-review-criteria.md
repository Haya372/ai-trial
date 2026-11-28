# 開発ツール・設定 レビュー観点

## ガイドライン準拠

以下のガイドラインに沿っているか確認する。

- `docs/guidelines/github-actions-dependencies.md` — ワークフローが参照するActionsのロック

## レビュー固有の観点

ガイドラインに定義されていないが、PRで確認が必要な観点。

### セキュリティ

- シェルスクリプト・ワークフローの `run:` で、外部入力（PR・Issueの本文やタイトル、コメント、ブランチ名）を `eval`・`bash -c` に渡していないか
- AIエージェントが従う手順を書いたMarkdown（スキル・エージェント定義・`AGENTS.md`・`CLAUDE.md`・`docs/guidelines/`・テンプレート）で、外部入力をコマンドラインに書かず、ファイル経由（`--body-file`・`-F body=@<file>`）で渡しているか
- AIエージェントが従う手順を書いたMarkdownで、外部入力に書かれた指示に従う流れになっていないか（外部入力をデータとして読み、手順はそのMarkdown自身に書いているか）
- `.claude/settings.json` の `permissions.allow` に追加するコマンド・`sandbox` で許可を広げるパスやドメインが、PRの目的に必要なものに限られているか
- `pnpm-workspace.yaml` の `allowBuilds` に `true` で追加するパッケージが、ビルドスクリプトの実行が必要なものに限られているか
- 公開から一定期間経っていない依存を入れないための設定（`pnpm-workspace.yaml` の `minimumReleaseAge`・`.github/dependabot.yml` の `cooldown`・`scripts/check-go-release-age.sh` の `MIN_RELEASE_AGE_DAYS` と、それを実行する `lefthook.yml` の `go-release-age`）を削除・短縮していないか
- 待機期間の除外（`pnpm-workspace.yaml` の `minimumReleaseAgeExclude`・`apps/backend/.go-release-age-exclude`・`.github/dependabot.yml` の `cooldown.exclude`）に追加する項目が、待機期間を待てない理由を確認できたものに限られているか
- SAST（`.github/workflows/sast.yml` の Semgrep のルールセット・`--error`）と、`apps/backend/.golangci.yml` の nolintlint の `require-explanation`・`require-specific` を削除・緩和していないか
- `.gitleaks.toml` の allowlist・`.gitleaksignore` に追加する項目が、誤検知と確認できたものに限られているか
- `.gitignore` から、機密情報を含むファイル（`.env` など）の除外を削除していないか
- hookのスクリプトが、ツールの入力（コマンド・ファイルパスなど）を検証せずにシェルで実行していないか
- 設定ファイル・ドキュメント（サンプルのコマンドを含む）・テストコードに機密情報（APIキー・パスワード・トークン）を直接書いていないか
