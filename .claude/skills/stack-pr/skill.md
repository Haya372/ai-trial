---
name: stack-pr
description: gh-stackを使ってスタックPRを作成するとき。複数の独立した関心事が混在する場合にPRを分割して作成する。
---

# スタックPR作成スキル

複数の独立した関心事（APIエンドポイントごと・DBスキーマ変更・UI実装など）が混在する場合に、gh-stackを使ってスタックPRに分割する。

## 前提

gh-stack がインストール済みであること。未インストールの場合は以下を実行する。

```bash
gh extension install github/gh-stack
```

## 手順

### ステップ 1: 分割単位を決める

関心事ごとにPRを分割する。各PRが独立してレビューできる単位になるよう設計する。

例（ユーザー登録APIを実装する場合）:
1. DBスキーマ変更 + マイグレーション
2. `POST /users` エンドポイント実装
3. フロントエンドの登録フォーム

### ステップ 2: スタックを初期化する

```bash
# mainブランチの最新状態から開始
git checkout main && git pull

# 最初のブランチを作成してスタックを初期化
git checkout -b <first-branch>
gh stack init
```

### ステップ 3: ブランチを積み上げる

1つ目の関心事を実装・コミットしたら、次のブランチを追加する。

```bash
# 1つ目の関心事を実装・コミット
git add <files> && git commit -m "<message>"

# 2つ目のブランチをスタックに追加
gh stack add <second-branch>

# 2つ目の関心事を実装・コミット
git add <files> && git commit -m "<message>"

# 必要なだけ繰り返す
gh stack add <third-branch>
```

### ステップ 4: GitHubにプッシュしてPRを作成する

```bash
gh stack submit
```

PRタイトル・本文の入力を求められる。`.github/PULL_REQUEST_TEMPLATE.md` の構成に沿って各PRの本文を記述する。

### ステップ 5: レビュー後の修正

レビューコメントを受けたら、対象ブランチで修正してコミットし、スタック全体をリベースする。

```bash
# 修正対象のブランチをチェックアウト
gh stack checkout <branch-name>

# 修正・コミット
git add <files> && git commit -m "fix: <修正内容>"

# スタック全体をリベースして整合性を保つ
gh stack rebase

# GitHubに反映
gh stack submit
```

### ステップ 6: マージ

全PRがApproveされたら、下から順にマージする。

```bash
gh stack merge
```

## スタックの確認

```bash
# 現在のスタック構成を確認
gh stack view
```

## チェックリスト

- [ ] 分割単位が「1PR = 1つの関心事」になっているか
- [ ] 各PRが単独でレビュー可能な状態か
- [ ] PRの本文にスタックの文脈（何番目のPRか）が記載されているか
- [ ] `gh stack submit` 後に全PRのURLを確認したか
