# @ai-trial/schema

OpenAPI スキーマの管理パッケージ。フロントエンド・バックエンドの型生成に使用する。

## ファイル構成

```
packages/schema/
├── openapi.yaml               # エントリーポイント（$ref で各ファイルを参照）
├── paths/
│   └── health.yaml            # エンドポイント定義（パスごとに1ファイル）
├── components/
│   └── schemas/
│       └── HealthResponse.yaml  # スキーマ定義（コンポーネントごとに1ファイル）
├── redocly.yaml               # lint ルール設定
├── oapi-codegen.yaml          # バックエンド (Go) 向けコード生成設定
└── orval.config.ts            # フロントエンド (TypeScript) 向けコード生成設定
```

`dist/openapi.yaml` はコード生成時に自動生成されるビルド成果物（gitignore済み）。

## スキーマビューアの起動

```bash
mise exec -- pnpm dev:schema
```

起動後、ブラウザで http://localhost:8080 を開くとスキーマが確認できる。

## スキーマの lint

```bash
mise exec -- pnpm lint:schema
```

コミット時に自動で実行される（lefthook の pre-commit フック）。

## コード生成

スキーマを更新したら以下を実行して各言語のコードを再生成する。  
内部で `redocly bundle` が走り `dist/openapi.yaml` を生成してからコード生成する。

```bash
# フロントエンド・バックエンド両方
mise exec -- pnpm generate

# フロントエンドのみ
mise exec -- pnpm generate:frontend

# バックエンドのみ
mise exec -- pnpm generate:backend
```

## 新しいエンドポイントを追加するとき

1. `paths/<endpoint>.yaml` にパス定義を追加
2. 必要に応じて `components/schemas/<Schema>.yaml` にスキーマを追加
3. `openapi.yaml` の `paths` と `components.schemas` に `$ref` を追加
4. `mise exec -- pnpm lint:schema` でバリデーション
5. `mise exec -- pnpm generate` でコード再生成
