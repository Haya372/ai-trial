# Session ドメイン

## 概要

認証済みユーザーのセッションを表すドメイン。セッションは PostgreSQL に保存され、HTTP-Only Cookie でセッション ID を管理する（ADR-020）。

## 用語定義

| 用語 | 説明 |
|---|---|
| Session | ユーザーの認証セッション。`ID`, `UserID`, `ExpiresAt` を持つ |

## エンティティ

### Session（インターフェース）

```go
type Session interface {
    ID() uuid.UUID
    UserID() uuid.UUID
    ExpiresAt() time.Time
}
```

インスタンス生成には `session.New(id, userID, expiresAt)` を使用する。内部実装（`sessionEntity`）は非公開。

## ビジネスルール

- セッションは有効期限（`ExpiresAt`）を持ち、期限切れのセッションは無効とみなす
- ログアウト時はセッションレコードを削除する
- ユーザー削除時はそのユーザーの全セッションを削除する（DB の CASCADE DELETE）
- ユーザー情報（`Email`, `DisplayName`）はセッションドメインには含めない。認証ミドルウェアが必要に応じて User リポジトリから取得する

## 関連ドキュメント

- [ADR-020: セッション管理](../adr/ADR-020.md)
- [ADR-001: クリーンアーキテクチャ採用](../adr/ADR-001.md)
- [User ドメイン](./user.md)
