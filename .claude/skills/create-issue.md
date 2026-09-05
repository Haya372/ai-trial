---
name: create-issue
description: GitHubにIssueを起票するとき。
---

# Issue 作成スキル

## 目的

GitHubのIssueをテンプレートに準拠した形で作成する。
テンプレートを遵守することで、情報の抜け漏れを防ぎ、Designerが設計を開始できる粒度を担保する。

---

## テンプレートの選択

| テンプレート | 用途 |
|---|---|
| `feature` | PRDをもとにした機能実装Issue |
| `task` | ドキュメント・設定・環境整備など |

---

## 手順

### ステップ 1: 情報を準備する

**Feature Issueの場合:**
- 対応するPRDのパス（`docs/prd/PRD-XXX-*.md`）
- 目的（PRDのゴールから1〜2文で要約）
- 受け入れ条件（PRDのAcceptance Criteriaから抽出）

**Task Issueの場合:**
- 概要と目的
- 具体的な作業リスト
- 完了条件

### ステップ 2: Issue作成コマンドを実行する

**Feature Issue:**
```bash
gh issue create \
  --title "feat: <タイトル>" \
  --body "$(cat <<'EOF'
## 関連PRD

docs/prd/<ファイル名>

## 目的

<目的>

## 受け入れ条件

- [ ] <条件1>
- [ ] <条件2>
EOF
)"
```

**Task Issue:**
```bash
gh issue create \
  --title "chore: <タイトル>" \
  --body "$(cat <<'EOF'
## 概要

<目的と背景>

## やること

- [ ] <作業1>
- [ ] <作業2>

## 完了条件

<完了の判断基準>
EOF
)"
```

---

## チェックリスト

- [ ] テンプレートの全セクションが埋まっているか
- [ ] タイトルがコミットメッセージ規約に沿っているか（`feat:` / `chore:` など）
- [ ] 受け入れ条件が具体的で検証可能か

---

## 参照ドキュメント

- `docs/guidelines/development-flow.md` — 開発フロー（Issueの位置づけ）
- `docs/guidelines/guidelines.md` — コミットメッセージ規約
