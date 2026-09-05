---
name: create-issue
description: GitHubにIssueを起票するとき。
---

# Issue 作成スキル

テンプレートの内容は `.github/ISSUE_TEMPLATE/` を参照すること。

## テンプレートの選択

| テンプレート | 用途 |
|---|---|
| `feature.md` | PRDをもとにした機能実装Issue |
| `task.md` | ドキュメント・設定・環境整備など |

## コマンド

```bash
gh issue create --title "<type>: <タイトル>" --template <テンプレートファイル名>
```

## チェックリスト

- [ ] テンプレートの全セクションが埋まっているか
- [ ] タイトルがコミットメッセージ規約に沿っているか（`feat:` / `chore:` など）
- [ ] 受け入れ条件が具体的で検証可能か
