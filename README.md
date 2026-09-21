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

# 6. Claude Code プラグインをインストール（PRレビュー・TDD支援用）
# Claude Code 上で実行する
/plugin install code-review@claude-plugins-official
/plugin install pr-review-toolkit@claude-plugins-official
/plugin install superpowers@claude-plugins-official
```

`.claude/settings.json` の `enabledPlugins` はインストール済みプラグインを自動有効化するだけで、未インストールの場合は自動インストールされない（Claude Code の既知の制限）。上記コマンドで各自インストールすること。

### コマンド実行の注意事項

pnpm は corepack 経由で実行するとシグネチャ検証エラーが発生する場合がある。
必ず `mise exec --` を前置して実行すること。

```bash
# OK
mise exec -- pnpm <command>

# NG（シグネチャ検証エラーが発生する可能性あり）
pnpm <command>
```

## 他のAIエージェントでの開発

このプロジェクトはClaude Code以外のAIコーディングエージェントでも開発できる。

| ツール | 参照する設定ファイル |
|---|---|
| Claude Code | `CLAUDE.md`（`AGENTS.md` を読み込む） |
| Codex | `AGENTS.md`（標準で自動読み込み） |
| Gemini CLI | `AGENTS.md`（`.gemini/settings.json` で読み込み対象に設定済み） |

プロジェクト共通の規約・開発フロー・制約は `AGENTS.md` に集約している。
Claude Code固有のSubagent/Skillの使い方は `CLAUDE.md` に記載している。

## ドキュメント

- [開発ガイドライン](docs/guidelines/guidelines.md)
- [ADR](docs/adr/)
