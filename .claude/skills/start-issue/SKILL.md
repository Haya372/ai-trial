---
name: start-issue
description: GitHubのIssueに着手するとき。Issueを読んで状況を理解し、適切な名前のブランチを作成して作業を開始する。「Issue〇〇を対応したい」「Issue〇〇に着手する」「Issue〇〇をやる」「〇〇番のIssueを進めたい」のような発言で呼び出す。
---

# Issue 着手スキル

## 目的

GitHubのIssueを読んで作業内容を理解し、適切なブランチを作成して実装に入れる状態を整える。

---

## 手順

### ステップ 1: Issueを読む

```bash
gh issue view <number>
```

- タイトル・概要・作業内容・受け入れ条件を把握する
- Issueの種類（機能追加・バグ修正・チョア・ドキュメント）を判断する
- Issue本文に「関連PRD」の記載がないか確認する。記載がない場合でも、「関連Issue: #NN（親Issue）」のように親Issueへの参照があれば `gh issue view <親Issue番号>` で親Issueの本文も確認し、関連PRDが紐づいていないか調べる（サブイシュー自身の本文にはPRDへの直接参照がないことが多いため）

### ステップ 2: ブランチ名を決める

以下の規則でブランチ名を決める。

| Issueの種類 | プレフィックス | 例 |
|---|---|---|
| 機能追加 | `feat/` | `feat/issue-30-add-dependabot` |
| バグ修正 | `fix/` | `fix/issue-15-null-check` |
| チョア・設定・環境 | `chore/` | `chore/issue-30-dependabot` |
| ドキュメント | `docs/` | `docs/issue-30-update-readme` |
| リファクタリング | `refactor/` | `refactor/issue-22-extract-helper` |

- Issue番号を必ず含める（例: `issue-30`）
- 内容が分かる短い説明をケバブケースで付ける
- 英語で書く

### ステップ 3: ブランチを作成する

```bash
git checkout -b <branch-name>
```

mainブランチの最新状態から作成する。

### ステップ 4: 作業内容を整理してユーザーに報告する

以下の内容を簡潔にまとめてユーザーに伝える。

- ブランチ名
- Issueの作業内容（何をするか）
- 次に使うべきスキルまたはすぐ実装できるか

### 次のステップの判断

判断手順:

1. ステップ1で確認した「関連PRD」の有無を確認する（Issue自身、または親Issueが持つ関連PRDを含む）
2. 関連PRDがある場合、`docs/spec/` にそのPRDに対応する仕様書が存在し、かつPDM合意済かを確認する
   - 存在しない／未合意 → まず `docs/guidelines/development-flow.md` を読んでフローを確認し、`specification` スキルでの仕様定義をユーザーに提示する（このIssueの実装には着手しない。実装が技術的に単純であっても例外にしない）
   - 存在し合意済み → 以下の表で通常通り判断する
3. 関連PRDがない場合は以下の表で判断する

| 作業の性質 | 次のスキル |
|---|---|
| アプリケーション機能が複雑・設計が必要 | `docs/guidelines/development-flow.md` を読んでフローを確認してからユーザーに提示する |
| Claude Codeスキル・ガイドライン・ワークフロー等の開発者ツール系 | そのまま実装（specification / detailed-design はスキップ） |
| シンプルな設定・ファイル追加（関連PRDがない、独立したIssue） | そのまま実装 |
| バグ修正 | `systematic-debugging` スキル |

---

## チェックリスト

- [ ] Issueの内容を把握したか
- [ ] ブランチ名にIssue番号が含まれているか
- [ ] ブランチ名が英語・ケバブケースになっているか
- [ ] ユーザーに作業内容と次のステップを伝えたか
