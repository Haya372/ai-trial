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

`.github/PULL_REQUEST_TEMPLATE.md` の構成を参照しながら、各セクションの内容を実際の値で埋めてbodyを構築する。

```bash
gh pr create \
  --title "<type>: <English title>" \
  --body "$(cat <<'EOF'
## 概要

<変更内容の簡潔な説明>

...（テンプレートのセクションを埋める）
EOF
)"
```

## チェックリスト

- [ ] テンプレートの全セクションが埋まっているか
- [ ] タイトルが英語かつコミットメッセージ規約に沿っているか（`feat:` / `fix:` / `chore:` など）
- [ ] Issue番号が正しく参照されているか
