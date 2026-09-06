# オブザーバビリティガイドライン

採用方針は [ADR-014](../adr/ADR-014-observability-policy.md) を参照。

---

## ログ

### ログレベルの使い分け

| レベル | 出力環境 | 使う場面 | 使わない場面 |
|---|---|---|---|
| DEBUG | 開発のみ | 処理の内部状態・変数の値確認 | 本番環境（デフォルト無効） |
| INFO | 全環境 | サービスの境界を越えるイベント（起動・終了・リクエスト受付・外部API呼び出し） | 関数の入退出・ループ内の毎回の処理 |
| WARN | 全環境 | 処理は続くが対応が必要な状態（リトライ発生・設定値のフォールバック） | エラーが確定している状況（ERROR を使う） |
| ERROR | 全環境 | 処理が失敗し呼び出し元にエラーを返す時点、または本来起こりえないデータ不整合を検知した時点 | エラーを受け取って上位に伝播するだけの箇所（二重ログを避ける） |

### ログを出すべきイベントの判断基準

イベントの種類ごとにレベルと出力有無を定める。

| イベント | レベル | 理由 |
|---|---|---|
| サーバー起動・終了 | INFO | 運用上の状態変化 |
| HTTPリクエスト受付・完了 | INFO | 証跡（ミドルウェアで一括出力） |
| 外部API呼び出し（成功・失敗とも） | INFO / ERROR | 証跡として残す。成功はINFO、失敗はERROR |
| DB操作の失敗 | ERROR | 成功は記録不要（トレーシングで把握） |
| 認証・認可の失敗 | WARN | 不正アクセス検知のために残す |
| バックグラウンドジョブの開始・終了 | INFO | 実行された証跡 |
| **本来起こりえないデータ不整合** | **ERROR** | **バグ・データ破損の可能性があるため発見箇所で即記録** |
| 関数の入退出 | 出さない | 粒度が細かすぎてノイズになる |
| DBクエリの成功 | 出さない | トレーシングで把握する |

**外部API呼び出し: 成功・失敗ともに INFO / ERROR で記録する**

```go
// 外部APIを呼ぶ前後で証跡を残す
logger.Info("calling external api", "service", "payment", "endpoint", "/charge")
resp, err := paymentClient.Charge(ctx, req)
if err != nil {
    logger.Error("external api failed", "error", err, "service", "payment")
    return err
}
logger.Info("external api succeeded", "service", "payment", "status", resp.Status)
```

**DB操作: 失敗時のみ ERROR で記録する**

```go
user, err := s.repo.FindByID(ctx, id)
if err != nil {
    logger.Error("failed to query database", "error", err, "operation", "find_user")
    return nil, err
}
// 成功時はログ不要（スパンで処理時間・成否を把握する）
```

**サーバー起動・終了・認証失敗**

```go
logger.Info("server starting", "addr", srv.Addr)
logger.Info("shutting down server")

logger.Warn("authentication failed", "reason", "invalid_token", "path", r.URL.Path)
```

**出さない: 関数の内部処理**

```go
// NG: 関数の入退出は出さない
func (s *UserService) FindByID(ctx context.Context, id string) (*User, error) {
    logger.Debug("FindByID called", "id", id)  // ← 不要
    user, err := s.repo.FindByID(ctx, id)
    logger.Debug("FindByID finished")           // ← 不要
    return user, err
}
```

### 構造化フィールドのルール

ログには必ず意味のある構造化フィールドを付与する。フリーテキストのみのログは検索できないため禁止。

```go
// NG: フリーテキストだけ
logger.Info("user login failed for user 12345")

// OK: 構造化フィールドで記録
logger.Warn("user login failed", "user_id", userID, "reason", "invalid_password")
```

**共通フィールド**

| フィールド名 | 型 | 説明 |
|---|---|---|
| `error` | string | エラー内容。`err.Error()` の値（ERROR レベル必須） |
| `request_id` | string | リクエストID（トレーシングと紐付ける） |
| `user_id` | string | 操作ユーザーのID（認証済みリクエストの場合） |
| `path` | string | HTTPリクエストのパス |
| `method` | string | HTTPメソッド |
| `status` | int | HTTPステータスコード |
| `duration_ms` | int | 処理時間（ミリ秒） |

### ログに含めてはならない情報

| 禁止情報 | 具体例 |
|---|---|
| 認証情報 | パスワード、APIキー、JWTトークンの中身 |
| 個人情報 | メールアドレス、氏名、住所、電話番号 |
| DBの内部詳細 | SQLクエリ文字列、DBエラーメッセージそのもの |
| スタックトレース | `runtime/debug.Stack()` の出力（障害調査時のデバッグログを除く） |

```go
// NG: 認証情報をそのまま記録
logger.Info("user logged in", "password", req.Password)  // 絶対禁止

// OK: IDのみ記録
logger.Info("user logged in", "user_id", user.ID)
```

### エラーログは末端で1回だけ出す

ログはエラーが最終的に処理される箇所（ハンドラー・ゴルーチンの末端）で1回だけ出す。中間層では `fmt.Errorf` でコンテキストをエラーに乗せて返す。こうすることで、ハンドラーの1行のログにエラーの発生経路が含まれる。

```go
// 中間層: ログを出さず、コンテキストをエラーに乗せてラップする
func (r *UserRepository) FindByID(ctx context.Context, id string) (*User, error) {
    user, err := r.db.QueryRowContext(ctx, query, id)
    if err != nil {
        return nil, fmt.Errorf("user_repository.find_by_id id=%s: %w", id, err)
    }
    return user, nil
}

func (s *UserService) FindByID(ctx context.Context, id string) (*User, error) {
    user, err := s.repo.FindByID(ctx, id)
    if err != nil {
        return nil, fmt.Errorf("user_service.find_by_id: %w", err)
    }
    return user, nil
}

// ハンドラー（末端）で1回だけ記録
// error の文字列に経路が含まれている:
// "user_service.find_by_id: user_repository.find_by_id id=123: connection refused"
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
    user, err := h.service.FindByID(r.Context(), id)
    if err != nil {
        logger.Error("failed to get user", "error", err)
        writeInternalError(w)
        return
    }
    writeJSON(w, user)
}
```

### データ不整合は発見箇所で即記録する

通常のエラー伝播とは異なり、「本来起こりえない状態」を検知した場合は中間層であっても発見箇所で ERROR を出す。ハンドラーまで伝播するころには文脈が失われており、バグ調査が困難になるため。

**本来起こりえない状態の例**

- 外部キーが存在するはずのレコードが見つからない
- ステータスが有効な値の列挙に含まれない
- 集計結果が負数になっている

```go
// 注文は必ず対応するユーザーが存在するはずなのに見つからない
user, err := s.userRepo.FindByID(ctx, order.UserID)
if err != nil {
    // データ不整合: 発見箇所でERRORを出し、詳細を記録する
    logger.Error("data inconsistency: user not found for existing order",
        "order_id", order.ID,
        "user_id", order.UserID,
    )
    return nil, fmt.Errorf("user not found for order %s: %w", order.ID, err)
}
```

---

## メトリクス

採用: OpenTelemetry Go SDK + Prometheus exporter

### 収集するメトリクス（RED方式）

| メトリクス名 | 種別 | ラベル | 説明 |
|---|---|---|---|
| `http_requests_total` | Counter | `method`, `path`, `status` | リクエスト総数 |
| `http_errors_total` | Counter | `method`, `path`, `status` | 4xx/5xx のエラー数 |
| `http_request_duration_seconds` | Histogram | `method`, `path` | リクエスト処理時間 |

### 命名規則

- スネークケースで記述する（例: `http_request_duration_seconds`）
- 単位をサフィックスに含める（例: `_seconds`, `_bytes`, `_total`）
- ラベル値はパスのパラメーター部分を固定文字列に置換する（カーディナリティ爆発の防止）

```go
// NG: パスパラメーターをそのままラベルに使う
// → /users/1, /users/2, /users/3 ... でラベルが無限に増える
attrs := []attribute.KeyValue{
    attribute.String("path", r.URL.Path),  // /users/123
}

// OK: ルートパターンを使う
attrs := []attribute.KeyValue{
    attribute.String("path", "/users/{id}"),  // chi.RouteContext から取得
}
```

---

## トレーシング

採用: OpenTelemetry Go SDK + OTLP エクスポーター

### スパンを作るタイミング

| 作る | 作らない |
|---|---|
| HTTPハンドラーの処理全体 | ユーティリティ関数の内部 |
| DBクエリ1回 | ループの各イテレーション |
| 外部APIの呼び出し1回 | 単純な値変換・マッピング |

### スパン名と属性の規則

- スパン名は `<コンポーネント>.<操作>` の形式（例: `db.query`, `http.request`）
- HTTPスパンには `http.method`, `http.route`, `http.status_code` を付与する
- DBスパンには `db.system`, `db.operation`, `db.table` を付与する（SQLクエリ文字列は含めない）

```go
ctx, span := tracer.Start(ctx, "db.query")
defer span.End()

span.SetAttributes(
    attribute.String("db.system", "postgresql"),
    attribute.String("db.operation", "SELECT"),
    attribute.String("db.table", "users"),
    // NG: attribute.String("db.statement", query) ← SQLは含めない
)

if err != nil {
    span.RecordError(err)
    span.SetStatus(codes.Error, err.Error())
}
```
