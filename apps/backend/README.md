# backend

バックエンドAPIサーバー（Go）。

## 開発

```bash
# DB起動（プロジェクトルートで実行）
docker compose up -d

# DBマイグレーション
make migrate-up

# 開発サーバー起動
go run ./cmd/server/
```

## その他のコマンド

```bash
# ビルド
make build

# テスト
make test

# Lint
make lint

# フォーマット
make fmt
```
