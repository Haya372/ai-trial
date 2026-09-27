# ADR-030: 認証任意ミドルウェアの導入

- ステータス: 承認済

## コンテキスト

予定共有機能（PRD-004・SPEC-004）の `GET /shares/{token}` は認証不要で共有リンクを閲覧できるエンドポイントだが、ログイン済みの場合はレスポンスに `isOwnEvent`（予定作成者本人か）・`isSubscribed`（既に自分のカレンダーに追加済みか）を含める必要がある（Issue #199 受け入れ条件）。

既存の `interface/middleware/auth.go::RequireAuth` はセッションが欠如／無効／エラーの場合に常に401を返す設計になっており、そのままでは「未ログインなら素通し、ログイン済みなら `ctxkey.User` をセット」という要件を満たせない。

同種の要求は今後、他の公開系（例：ゲスト閲覧を許容するリソース）でも発生し得るため、共有機能専用の場当たり実装ではなく、既存の `RequireAuth` と一貫した形でミドルウェアとして提供する。

## スコープ外

- セッション生成・失効ロジックの変更（ADR-020 の範囲）
- 認証済みユーザー向けの権限判定（本ミドルウェアは identity の取得のみ）
- レートリミット・不正アクセス検知

## 決定

`interface/middleware/auth.go` に「セッションからユーザーを解決する」ヘルパー `resolveSessionUser` を追加し、`RequireAuth` と新設 `OptionalAuth` の両方から呼び出す形に整理する。

- `OptionalAuth` は未認証／セッション無効／リポジトリエラーのいずれの場合も `next` を呼び、`ctxkey.User` はセットしない
- 認証成功時のみ `ctxkey.User` に `user.User` をセットする
- レスポンスへの401書き込みは `RequireAuth` のみが行う

### 関数シグネチャ

```go
// resolveSessionUser looks up the user for the session cookie on r. It returns
// (nil, nil) when there is no session or the session is expired/invalid/missing
// its user. It returns a non-nil error only for an unexpected repository failure.
func resolveSessionUser(
    ctx context.Context,
    r *http.Request,
    sessionRepo session.Repository,
    userRepo user.Repository,
    logger *slog.Logger,
) (user.User, error)

func OptionalAuth(
    sessionRepo session.Repository,
    userRepo user.Repository,
    logger *slog.Logger,
) func(http.Handler) http.Handler
```

### 挙動一覧

| 状態 | `RequireAuth` | `OptionalAuth` |
|---|---|---|
| クッキー無し | 401 | 素通し（user=nil） |
| クッキーのUUIDパース失敗 | 401 | 素通し（user=nil、Warnログ） |
| セッション見つからず | 401 | 素通し（user=nil、Warnログ） |
| セッションリポジトリエラー | 401 | 素通し（user=nil、Errorログ） |
| ユーザー見つからず／リポジトリエラー | 401 | 素通し（user=nil、Errorログ） |
| 認証成功 | `ctxkey.User` セット | `ctxkey.User` セット |

`OptionalAuth` の設計方針：認証に付随する不整合（DB エラー等）で公開閲覧そのものを止めないため、リポジトリエラーでも `next` を呼ぶ。ただし観測性のため Warn/Error ログは残す。

### ハンドラ側の取り出し方

`GET /shares/{token}` ハンドラは既存の `EventHandler.GetEvents` と同じパターンで `ctxkey.User` を参照する:

```go
u, _ := r.Context().Value(ctxkey.User).(user.User)  // 未ログイン時は u == nil
if u != nil {
    // isOwnEvent / isSubscribed を算出
}
```

## 検討した選択肢

### 選択肢1: `RequireAuth` と独立した `OptionalAuth` を新規追加（重複実装）

#### 概要

`RequireAuth` を触らず、新しい `OptionalAuth` に同等のセッション解決コードをコピーする。

#### メリット

- `RequireAuth` の既存テストが影響を受けない

#### デメリット

- クッキー取得・UUIDパース・セッション取得・ユーザー取得の一連のロジックが2箇所に重複し、変更漏れの原因になる
- セッション周りの実装変更（ADR-020 の見直し等）時に修正コストが二重にかかる

### 選択肢2: セッション解決を共通ヘルパーに抽出し、両ミドルウェアから呼ぶ（本決定）

#### 概要

`resolveSessionUser(ctx, r, sessionRepo, userRepo, logger) (user.User, error)` を `middleware` パッケージ内部関数として抽出し、`RequireAuth` は「nil または error → 401」、`OptionalAuth` は「常に `next`、user が取れれば context にセット」というレスポンスポリシーだけを担う。

#### メリット

- セッション→ユーザー解決のロジックが単一
- 各ミドルウェアの責務がレスポンスポリシーに絞られ、テストしやすい
- 依存関係逆転（DIP）を維持: 上位のミドルウェアは抽象（Repository）に依存

#### デメリット

- 既存の `RequireAuth` テストのうち、内部ヘルパー導入によりログレベルが変わるケースがあれば追随が必要（軽微）

### 選択肢3: ミドルウェアを使わず、ハンドラ内でセッション解決を行う

#### 概要

`/shares/{token}` ハンドラ内で直接 `sessionRepo`・`userRepo` を呼び出してユーザーを解決する。

#### メリット

- ミドルウェア数が増えない

#### デメリット

- 認証というインフラ横断的な関心事がハンドラに漏れる（単一責任原則違反）
- 今後同種の要求が出るたびに各ハンドラで同じコードを書くことになる

## 決定理由

- セッション解決という下位詳細を1箇所にまとめ、レスポンスポリシー（401 vs 素通し）だけを2つのミドルウェアが担う設計は SRP と DRY を両立できる
- 既存の `RequireAuth` の挙動は変更しないため、他エンドポイントへの影響がない
- `OptionalAuth` を独立したミドルウェアとして提供することで、ルーティング定義（`di/container.go::newRouter`）を読むだけで各エンドポイントの認証ポリシーが分かる

## 結果

### 良い影響

- 認証任意エンドポイントの追加が今後低コストで可能
- `RequireAuth` と `OptionalAuth` の挙動差が「レスポンスポリシー」のみに絞られ、テストのカバレッジが分かりやすい

### 悪い影響

- ミドルウェアが1種類増えるので、ルーター定義を読む側は「どちらを使うか」を毎回判断する必要がある。ただし用途は「認証必須か任意か」の1軸で判断できるため負担は小さい
