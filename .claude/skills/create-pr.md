---
name: create-pr
description: GitHubにPRを作成するとき。
---

# PR 作成スキル

テンプレートの内容は `.github/PULL_REQUEST_TEMPLATE.md` を参照すること。

## コマンド

```bash
gh pr create --title "<type>: <タイトル>" --body "$(cat .github/PULL_REQUEST_TEMPLATE.md)"
```

`--body` にテンプレートを渡した上で、各セクションを実際の内容に書き換えて実行する。

## チェックリスト

- [ ] テンプレートの全セクションが埋まっているか
- [ ] タイトルがコミットメッセージ規約に沿っているか（`feat:` / `fix:` / `chore:` など）
- [ ] Issue番号が正しく参照されているか
