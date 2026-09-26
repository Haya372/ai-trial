# Playwright MCP 利用ガイド

## 概要

AIコーディングエージェント（Claude Codeなど）が、実装後のUI動作確認をブラウザ上で自律的に行うためのMCPサーバー。
ページ遷移・クリック・スクリーンショット取得などをエージェントから直接実行できる。

設定は `.mcp.json` にあり、リポジトリをチェックアウトしたエージェントは自動的に利用可能になる。

## セキュリティ上の制約

任意のURLへのアクセスを許可すると、意図しない外部サイトへのアクセスや認証情報の漏洩につながるリスクがある。
そのため `.mcp.json` では `--allowed-origins` オプションでアクセス先を `http://localhost:*` / `http://127.0.0.1:*` に限定している。

```json
"--allowed-origins",
"http://localhost:*;http://127.0.0.1:*"
```

これにより、ローカルで起動した `apps/web`（デフォルト: `http://localhost:5173`、バックエンドAPIは同一オリジンから `/api` にプロキシされる）へのアクセスのみを許可し、外部ホストへのアクセスは `net::ERR_BLOCKED_BY_CLIENT` として拒否される。

### 注意: セキュリティ境界ではない

Playwright MCP公式ドキュメントの記載どおり、`--allowed-origins` / `--blocked-origins` は**厳密なセキュリティ境界としては機能しない**（リダイレクトには影響しないなどの制約がある）。
あくまで誤操作による意図しない外部アクセスを防ぐための便宜的な制御であり、機密情報を扱う操作をエージェントに任せる場合は別途レビューが必要。

## 利用方法

Claude Codeでは `.mcp.json` の設定により、セッション起動時に自動でPlaywright MCPサーバーが起動する。
実装後のUI確認では、以下のようなツールを使う。

- `browser_navigate` — 指定URLへ遷移する
- `browser_snapshot` / `browser_take_screenshot` — 画面状態を取得する
- `browser_click` / `browser_type` — 要素の操作を行う

バックエンド（`cd apps/backend && go run ./cmd/server`）と `apps/web` の開発サーバー（`mise exec -- pnpm dev:web`）を起動した状態で、対象ページに遷移し、実装した変更が意図通り表示・動作するかを確認する。

## トラブルシューティング

### 初回起動が遅い / ブラウザがダウンロードされる

初回実行時にPlaywrightが対象ブラウザ（Chromiumなど）をダウンロードするため時間がかかる。以降はキャッシュされる。

### `net::ERR_BLOCKED_BY_CLIENT` が意図せず出る

`apps/web` 以外のオリジン（ポート番号違いを含む）にアクセスしようとしていないか確認する。`apps/web` の開発サーバーが起動しているホスト・ポートと `.mcp.json` の `--allowed-origins` が一致しているか見直す。
