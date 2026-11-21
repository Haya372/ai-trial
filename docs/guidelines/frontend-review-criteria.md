# フロントエンド レビュー観点

## ガイドライン準拠

以下のガイドラインに沿っているか確認する。

- `docs/guidelines/design-guidelines.md` — フロントエンド設計（コンポーネント・状態管理・API通信設計）
- `docs/guidelines/component-guidelines.md` — UIコンポーネント実装（variant・className・cva）

## レビュー固有の観点

ガイドラインに定義されていないが、PRで確認が必要な観点。

### セキュリティ

- `dangerouslySetInnerHTML` など、HTMLを直接描画する箇所に渡すユーザー入力・APIレスポンス由来の文字列を、DOMPurifyなどのサニタイズライブラリに通しているか（手書きの置換・正規表現でのタグ除去はサニタイズとみなさない）
- ユーザー入力を含むURLを `href`・`src`・`window.location` に設定する箇所で、`javascript:` などのスキームを拒否しているか
- 認証トークン・機密情報を `localStorage`・`sessionStorage`・JavaScriptから読めるCookieに保存していないか
- 機密情報（APIキー・シークレット）がクライアントのコードや、ビルド時にバンドルされる環境変数（`VITE_` で始まるもの）に含まれていないか

### パフォーマンス

- 不必要な再レンダリングが発生していないか（メモ化の欠如・依存配列の誤り）
- `useEffect` の依存配列が正確か（不足・過剰どちらも問題）
- 重い処理が `useMemo` / `useCallback` でメモ化されているか

### アクセシビリティ

- インタラクティブ要素（ボタン・リンク）に `aria-label` またはテキストコンテンツがあるか
- フォーム要素に `label` が関連付けられているか
- セマンティックなHTML要素が使われているか（`div` のクリックイベントより `button` を使うべき）
- キーボード操作（Tab・Enter・Escape）が機能するか

### エラー・ローディング処理

- データ取得中のローディング状態が表示されるか
- エラー時のフォールバックUIがあるか
- エラー境界（Error Boundary）が適切に配置されているか
