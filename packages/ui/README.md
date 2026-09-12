# @repo/ui

共有UIコンポーネントパッケージ。

## コンポーネントカタログ

UIコンポーネントの一覧・バリアント確認にはStorybookを使用する。

```bash
# 開発サーバーを起動（localhost:6006 で確認）
mise exec -- pnpm --filter @repo/ui storybook

# 静的ビルド
mise exec -- pnpm --filter @repo/ui build-storybook
```
