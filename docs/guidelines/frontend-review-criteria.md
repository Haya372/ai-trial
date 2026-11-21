# フロントエンド レビュー観点

## ガイドライン準拠

以下のガイドラインに沿っているか確認する。

- `docs/guidelines/design-guidelines.md` — フロントエンド設計（コンポーネント・状態管理・API通信設計）
- `docs/guidelines/component-guidelines.md` — UIコンポーネント実装（variant・className・cva）

## レビュー固有の観点

ガイドラインに定義されていないが、PRで確認が必要な観点。

### セキュリティ

- `dangerouslySetInnerHTML`・`ref` 経由の `innerHTML`・`insertAdjacentHTML` など、HTMLを直接描画する箇所に渡すユーザー入力・APIレスポンス由来の文字列を、DOMPurifyなどのサニタイズライブラリに通しているか（手書きの置換・正規表現でのタグ除去はサニタイズとみなさない）
- ユーザー入力を含むURLを `href`・`src`・`window.location` に設定する箇所で、URLを `new URL()` で解析し、`http:`・`https:` のスキームまたは相対パスだけを設定しているか
- `eval`・`new Function`・文字列を渡す `setTimeout` に、ユーザー入力・APIレスポンス由来の文字列を渡していないか
- 遷移先（`navigate`・`redirect`・`window.location`）にクエリパラメータなどのユーザー入力を使う箇所で、`/` で始まり `//`・`/\` で始まらない同一オリジンのパスだけを許可しているか（ログイン後に元のページへ戻す処理など）
- 認証トークン・機密情報を `localStorage`・`sessionStorage`・JavaScriptから読めるCookieに保存していないか
- 機密情報（APIキー・シークレット）がクライアントのコードや、ビルド時にバンドルされる環境変数（`VITE_` で始まるもの）に含まれていないか
- URLに認証情報（共有トークンなど）を含むページで、外部サイトへのリンクに `rel="noreferrer"` を付け、外部のスクリプト・画像を読み込まず、Referer経由で認証情報を送っていないか
- 外部ドメインのスクリプト・スタイルシートを読み込む場合、npmパッケージとしてバンドルするか、`integrity` 属性（SRI）を付けているか
- `postMessage` の受信側で `event.origin` を確認し、送信側で `targetOrigin` に `'*'` ではなく送信先のオリジンを指定しているか
- 権限によって操作を制限する箇所で、ボタンの非表示・無効化だけで制限せず、対応するAPIでも認可しているか

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
