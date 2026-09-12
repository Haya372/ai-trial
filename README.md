# スケジュール管理Webカレンダーアプリ

見やすさと予定登録のしやすさに特化したWebカレンダーアプリ。

## 概要

既存のカレンダーアプリ（Google Calendar等）の「見づらさ」「登録のしにくさ」を解決する。
月/週ビューでの一覧性、ワンクリックでの素早い登録、予定の共有をコアとして提供する。

## ペルソナ

**さくら（28歳）** — ハイブリッド勤務のマーケター。MTG中に次の予定をすぐ押さえたい。リモート/出社を視覚的に区別したい。

**拓也（42歳）** — 外回り営業部長。朝にその日の流れをパッと把握したい。アポを即登録したい。

## 環境構築

### 必要なツール

| ツール | 用途 |
|---|---|
| [mise](https://mise.jdx.dev/) | ツールバージョン管理 |
| [gh-stack](https://github.com/github/gh-stack) | スタックPR管理（`gh extension install github/gh-stack`） |

### セットアップ手順

```bash
# 1. リポジトリをクローン
git clone <repository-url>
cd ai-trial

# 2. mise でツールをインストール（Node.js, pnpm, gh 等が一括でインストールされる）
mise install

# 3. 依存パッケージをインストール
mise exec -- pnpm install

# 4. git hooks をセットアップ
mise exec -- pnpm exec lefthook install

# 5. gh-stack をインストール（スタックPR管理用）
gh extension install github/gh-stack
gh skill install github/gh-stackも
```

### コマンド実行の注意事項

pnpm は corepack 経由で実行するとシグネチャ検証エラーが発生する場合がある。
必ず `mise exec --` を前置して実行すること。

```bash
# OK
mise exec -- pnpm <command>

# NG（シグネチャ検証エラーが発生する可能性あり）
pnpm <command>
```

## Claude レビュー環境のセットアップ

PRが作成・更新されると Claude が自動でコードレビューを実行する。初回のみ以下の手順が必要。

**前提条件:** リポジトリの管理者権限が必要。

```bash
# Claude Code CLI で GitHub App をインストール（GitHub App + Secrets を一括設定）
/install-github-app
```

コマンドを実行すると以下が自動でセットアップされる:

- Anthropic GitHub App のリポジトリへのインストール
- `ANTHROPIC_API_KEY` の GitHub Secrets への登録

> **注意:** フォークからのPRでは GitHub Secrets にアクセスできないため、自動レビューは実行されない。

## ドキュメント

- [開発ガイドライン](docs/guidelines/guidelines.md)
- [ADR](docs/adr/)
