# 設計ガイドライン

基本設計原則（SOLID・DRY・YAGNI・KISS）は `docs/guidelines/guidelines.md` を参照。
このドキュメントはバックエンド・フロントエンド固有の設計指針を定義する。

## バックエンド設計

### レイヤー設計

- 責務ごとにレイヤーを分割する（例: Controller / Service / Repository）
- 上位レイヤーは下位レイヤーの抽象（インターフェース）に依存し、具体実装に依存しない
- レイヤーを越えるデータはDTOで明示的に変換する

### API設計

- エンドポイントはリソースを表す名詞を使う（例: `/users`, `/orders`）
- 成功・エラーのレスポンス形式をプロジェクト全体で統一する
- 入力バリデーションはAPIの境界で行う

### APIエラー定義

エラーレスポンスの構造・フィールド定義・セキュリティ制約は [ADR-013](../adr/ADR-013-api-error-structure.md) を参照。以下はその実装指針。

#### HTTPステータスコードとエラーコードの対応

| HTTPステータス | エラーコード | 用途 |
|---|---|---|
| 400 Bad Request | `VALIDATION_ERROR` | リクエストパラメータのバリデーション失敗 |
| 400 Bad Request | `INVALID_REQUEST` | 構造的に不正なリクエスト（JSON parse失敗等） |
| 401 Unauthorized | `UNAUTHORIZED` | 認証情報なし・無効 |
| 403 Forbidden | `FORBIDDEN` | 認証済みだがリソースへのアクセス権限なし |
| 404 Not Found | `NOT_FOUND` | リソースが存在しない |
| 409 Conflict | `CONFLICT` | 一意制約違反など競合状態 |
| 500 Internal Server Error | `INTERNAL_ERROR` | サーバー内部エラー（詳細はログに記録） |

新しいエラー種別が必要になった場合はこの表を更新する。

#### エラーコードの命名規則

- 大文字スネークケースで定義する（例: `NOT_FOUND`, `INVALID_EMAIL_FORMAT`）
- トップレベルのコードはエラー種別を表す（例: `VALIDATION_ERROR`, `UNAUTHORIZED`）
- `details[].code` はフィールドの制約違反を表す（例: `TOO_SHORT`, `REQUIRED`, `INVALID_FORMAT`）

#### 新しいエラーを追加するとき

1. HTTPステータスコードとエラーコードの対応を ADR-013 の対応表に照らして決める
2. 対応するエラーコード定数をバックエンドに定義する
3. `message` はバックエンドで事前定義した英語の文言を使う（ライブラリの生エラーをそのまま渡さない）

```go
// OK: 定義済みの定数と文言を使う
const (
    ErrCodeNotFound        = "NOT_FOUND"
    ErrCodeValidation      = "VALIDATION_ERROR"
    ErrCodeInternal        = "INTERNAL_ERROR"
)

writeError(w, http.StatusNotFound, ErrCodeNotFound, "Resource not found")

// NG: 内部エラーをそのまま渡す
writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
```

#### セキュリティ制約

レスポンスに含めてはならない情報:

- スタックトレース・ファイルパス・行番号
- SQLクエリ・テーブル名・DBのエラーメッセージ
- 使用ライブラリの名称やバージョン

`500 Internal Server Error` の `message` は常に `"Internal server error"` 固定とし、詳細はログにのみ記録する。

### ドメイン設計

- ドメイン固有の用語はコード・ドキュメント全体で統一する（`docs/domain/` 参照）
- ビジネスルールはドメイン層に集約し、インフラ層に漏らさない

## フロントエンド設計

### UIコンポーネント設計

- コンポーネントの責務を1つに絞る
- Presentational（表示のみ）と Container（ロジック・状態管理）を分離する
- 再利用するコンポーネントはProps/Stateのインターフェースを先に設計する

### 状態管理設計

- グローバル状態: 複数コンポーネント間で共有するデータのみ
- ローカル状態: そのコンポーネントで完結するデータ
- 状態の更新フローを一方向に保つ（単方向データフロー）

### 画面遷移設計

- 画面一覧とルーティングを定義する
- 認証が必要な画面と不要な画面を明示する
- 遷移条件（成功時・エラー時・未ログイン時）を定義する

### API通信設計

- 呼び出すAPIのエンドポイント・リクエスト・レスポンス形式を設計ドキュメントに明示する
- ローディング状態・エラー状態のハンドリング方針を定める
- エラーレスポンスの扱いを統一する（グローバルハンドリング vs 個別ハンドリング）
