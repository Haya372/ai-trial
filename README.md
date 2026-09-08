# スケジュール管理Webカレンダーアプリ

見やすさと予定登録のしやすさに特化したWebカレンダーアプリ。

## 概要

既存のカレンダーアプリ（Google Calendar等）の「見づらさ」「登録のしにくさ」を解決する。
月/週ビューでの一覧性、ワンクリックでの素早い登録、予定の共有をコアとして提供する。

## ペルソナ

**さくら（28歳）** — ハイブリッド勤務のマーケター。MTG中に次の予定をすぐ押さえたい。リモート/出社を視覚的に区別したい。

**拓也（42歳）** — 外回り営業部長。朝にその日の流れをパッと把握したい。アポを即登録したい。

## ドキュメント

- [開発ガイドライン](docs/guidelines/guidelines.md)
- [ADR](docs/adr/)

## Claude Code スキル

### PRレビュー (`review-pr`)

PR差分をカテゴリ別（frontend / backend / document）にサブAgentで並列レビューし、スコア7以上の指摘のみGitHub PRにインラインコメントとして投稿する。

**使い方:**

```
/review-pr 42
```

または

```
PR #42 をレビューして
```

**レビュー観点:**
- [フロントエンド](docs/guidelines/review-criteria/frontend.md)
- [バックエンド](docs/guidelines/review-criteria/backend.md)
- [ドキュメント](docs/guidelines/review-criteria/document.md)
