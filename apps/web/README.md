# web

フロントエンドアプリケーション（React + TypeScript + Vite）。

## 開発

```bash
# 開発サーバー起動
pnpm dev

# ビルド
pnpm build

# Lint（oxlint + format check + knip を並列実行）
pnpm lint

# フォーマット適用
pnpm format
```

## 技術スタック

| ツール | 役割 |
|---|---|
| React 19 + TypeScript 7 | UIフレームワーク |
| Vite 8 | バンドラー・開発サーバー |
| oxlint | Linter |
| Biome | Formatter |
| knip | 未使用コード検出 |

## i18n（国際化）

`i18next` + `react-i18next` による国際化基盤を導入済み。文言追加・言語切り替えの手順は [src/i18n/README.md](src/i18n/README.md) を参照。
