# @ai-trial/schema

OpenAPI スキーマの管理パッケージ。フロントエンド・バックエンドの型生成に使用する。

## ファイル構成

```text
packages/schema/
├── openapi.yaml               # エントリーポイント（$ref で各ファイルを参照）
├── paths/
│   └── health.yaml            # エンドポイント定義（パスごとに1ファイル）
├── components/
│   └── schemas/
│       └── HealthResponse.yaml  # スキーマ定義（コンポーネントごとに1ファイル）
├── redocly.yaml               # lint ルール設定
├── oapi-codegen.yaml          # バックエンド (Go) 向けコード生成設定
├── orval.config.ts            # フロントエンド (TypeScript) 向けコード生成設定
└── tsconfig.json              # TypeScript 設定
```

`dist/openapi.yaml` はコード生成時に自動生成されるビルド成果物（gitignore済み）。

## スキーマビューアの起動

```bash
mise exec -- pnpm dev:schema
```

起動後、ブラウザで http://localhost:8080 を開くとスキーマが確認できる。

