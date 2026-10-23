---
name: create-pr
description: GitHubにPRを作成するとき。
---

# PR 作成スキル

テンプレートの構成は `.github/PULL_REQUEST_TEMPLATE.md` を参照すること。

## 命名規則

- **タイトル**: 英語で書く（例: `feat: Add PR template`）
- **概要など本文**: 日本語で書く

## コマンド

> **注意**: `git -C <path>` は allowlist にマッチしないため使わない。カレントディレクトリがプロジェクトルートなら `git push` などをそのまま実行する。パス移動が必要な場合は `cd <path> && git <command>` の形にする。

`.github/PULL_REQUEST_TEMPLATE.md` の構成を参照しながら、各セクションの内容を実際の値で埋めてbodyを構築する。

body ファイルは、Write ツールでスクラッチパッドディレクトリ（システムプロンプトに記載の絶対パス）に作成し、その絶対パスを `--body-file` に渡す。

`gh` はsandbox外で実行されるため、`$TMPDIR` の値がsandbox内と異なり、`$TMPDIR/pr-body.md` を渡すと「no such file or directory」で失敗する。

```bash
# ※ --body "$(cat <<EOF ... EOF)" はsandbox環境で不安定なため使わない
# 事前に Write ツールで <scratchpad>/pr-body.md を作成する（テンプレートのセクションを埋める）

gh pr create \
  --title "<type>: <English title>" \
  --body-file "<scratchpad>/pr-body.md"
```

## チェックリスト

- [ ] テンプレートの全セクションが埋まっているか
- [ ] タイトルが英語かつコミットメッセージ規約に沿っているか（`feat:` / `fix:` / `chore:` など）
- [ ] Issue番号が正しく参照されているか
