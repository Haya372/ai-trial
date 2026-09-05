---
name: create-issue
description: GitHubにIssueを起票するとき。
---

# Issue 作成スキル

テンプレートの内容は `.github/ISSUE_TEMPLATE/` を読み込んで参照すること。

## テンプレートの選択

| テンプレート | 用途 |
|---|---|
| `feature.md` | PRDをもとにした機能実装Issue |
| `task.md` | ドキュメント・設定・環境整備など |

## 命名規則

- **タイトル**: 必ず英語で書く。日本語は絶対に使わない（例: `feat: Add calendar view`）
- **本文**: 日本語で書く

## コマンド

`--template` フラグはインタラクティブな入力を要求するため使用しない。
以下の手順で本文を組み立てて `--body` で渡す。

1. `.github/ISSUE_TEMPLATE/<テンプレートファイル>` を Read ツールで読み込む
2. テンプレートのコメント（`<!-- ... -->`）を実際の内容で置き換える
3. `gh issue create` の `--body` フラグに渡す

```bash
gh issue create --title "<type>: <English title>" --body "<テンプレートを埋めた本文>"
```

## ルール

- **ラベル**: `--label` フラグは使わない。ラベルの存在確認もしない。
- **関連PRD**: ファイルパスではなく、GitHubのファイルURLをMarkdownリンクで記載する。
  - リポジトリのURLは `gh repo view --json nameWithOwner,defaultBranchRef` で取得する。
  - 形式: `[docs/prd/PRD-XXX.md](https://github.com/<owner>/<repo>/blob/<branch>/docs/prd/PRD-XXX.md)`
- **ユーザーへの確認**: ラベルやその他の設定について確認のために止まらない。すぐに起票する。

## チェックリスト

- [ ] タイトルが英語になっているか（日本語が混入していないか）
- [ ] タイトルがコミットメッセージ規約に沿っているか（`feat:` / `chore:` など）
- [ ] 関連PRDがGitHubファイルへのMarkdownリンクになっているか
- [ ] テンプレートの全セクションが埋まっているか
- [ ] 受け入れ条件が具体的で検証可能か
