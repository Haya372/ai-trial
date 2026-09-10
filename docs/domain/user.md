# User ドメイン

## 概要

ユーザーアカウントを表すドメイン。認証情報（メールアドレス・パスワードハッシュ）とプロフィール情報（表示名）を保持する。

## 用語定義

| 用語 | 説明 |
|---|---|
| User | システムに登録されたアカウント。`ID`, `Email`, `DisplayName`, `PasswordHash` を持つ |
| Email | メールアドレスを表す値オブジェクト。RFC準拠の形式検証済み |

## エンティティ

### User（インターフェース）

```go
type User interface {
    ID() uuid.UUID
    Email() Email
    DisplayName() string
    PasswordHash() string
}
```

インスタンス生成には `user.New(id, email, displayName, passwordHash)` を使用する。内部実装（`userEntity`）は非公開。

### Email（値オブジェクト）

```go
type Email string

func NewEmail(s string) (Email, error)
```

- 生成時に正規表現でフォーマット検証を行う
- 検証失敗時はエラーを返す（ゼロ値の Email は存在しない）

## ビジネスルール

- メールアドレスはシステム全体で一意でなければならない（`ErrEmailTaken`）
- パスワードは保存前に必ずハッシュ化する（平文は保持しない）

## エラー

| エラー | 意味 |
|---|---|
| `ErrUserNotFound` | 指定した ID またはメールアドレスのユーザーが存在しない |
| `ErrEmailTaken` | 登録しようとしたメールアドレスがすでに使われている |

## 関連ドキュメント

- [ADR-001: クリーンアーキテクチャ採用](../adr/ADR-001.md)
- [Session ドメイン](./session.md)
