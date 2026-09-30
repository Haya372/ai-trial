# GitHub Actions 依存関係管理

## 概要

`.github/workflows/` 配下のワークフローが参照する Actions は、[gh-actions-lock](https://github.com/github/gh-actions-lock)（`gh` CLI拡張機能）でコミットSHAにロックし、`.github/workflows/actions.lock` で管理する。

ワークフロー内の `uses:` はタグ参照（例: `actions/checkout@v7.0.1`）のまま保つ。実体のコミットSHAは `actions.lock` に記録される。GitHub Actions自体がワークフロー実行時にこのロックファイルを検証し、記述と食い違う場合はジョブ開始前にrunをブロックする（`Invalid lockfile` エラー）。これに加えてCI上でも `gh actions-lock --no-fix` による検証を行い、PRレビュー時点で乖離を検出する。

`actions.lock` は `gh actions-lock` が生成するため、手動で編集しない。

zizmorの `unpinned-uses` ルールはデフォルトでハッシュピン留めを要求するが、SHAの実体は `actions.lock` が管理するため `.github/zizmor.yml` で `ref-pin` に緩和している。タグ参照とSHAの一致検証は `gh actions-lock` 側の責務。

以前使用していた `pinact`（ワークフローに直接SHAを書き込む方式）は役割が重複するため `mise.toml` から削除した。

## セットアップ

```bash
mise run actions-lock-setup
```

CIと同じバージョン（`v0.1.6`）にピン留めしてインストールする。

## 運用手順

### 新規Actionsを追加したとき

ワークフローに `uses:` を追記したら、以下を実行してロックファイルを更新する。

```bash
gh actions-lock
```

タグ参照へのピン留めとロックファイルの更新が自動で行われる。差分（ワークフローとロックファイルの両方）をコミットする。

### Actionsのバージョンを更新したとき（Dependabot含む）

DependabotがActionsのタグを更新するPRを作成した場合、そのPRのブランチで以下を実行してロックファイルを再生成し、コミットを追加する。

```bash
gh actions-lock --relock
```

手動でタグを更新した場合も同様。

### CI上の検証

`.github/workflows/action-lint.yml` の `actionlint` ジョブが、PRごとに以下を実行する。

```bash
gh actions-lock --no-fix
```

ワークフロー定義とロックファイルの内容に乖離がある場合、このジョブが失敗する。なお、この検証を経ずにワークフローが実行された場合でも、GitHub Actions自体のネイティブ検証（上記）により乖離があればrunは起動しない。
