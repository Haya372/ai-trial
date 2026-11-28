# User ドメイン

## 概要

ユーザーアカウントを表すドメイン。認証情報（メールアドレス・パスワードハッシュ）とプロフィール情報（表示名）を保持する。

## 用語定義

| 用語 | 説明 |
|---|---|
| User | システムに登録されたアカウント。ID・メールアドレス・表示名・パスワードハッシュ・連続ログイン失敗回数・ロック解除時刻を持つ |
| Email | メールアドレスを表す値オブジェクト。生成時にフォーマット検証を行い、不正な値は生成できない |
| Password | パスワードを表す値オブジェクト。生成時に長さを検証し bcrypt でハッシュ化する。User が保持するのはハッシュのみ |
| 連続ログイン失敗回数 (failed_login_count) | 直近のログイン成功・パスワード再設定・ロックの自動解除のいずれかから、連続して失敗した回数。0 以上の整数 |
| ロック解除時刻 (locked_until) | この時刻までログインを拒否する。NULL または `now` 以下の値は未ロック扱い |

## ビジネスルール

- メールアドレスはシステム全体で一意でなければならない
- パスワードは 8 文字以上 128 文字以下でなければならない
- パスワードは平文のまま保存してはならない（常にハッシュ化して保存する）
- ユーザー作成時には Password 値オブジェクトを介してのみパスワードを受け付ける

### アカウントロック

- `failed_login_count >= 5` に到達した時点でアカウントをロックする。ロックは `locked_until = now + 15分` を設定して表現する
- ロック中の判定は `locked_until IS NOT NULL AND locked_until > now`。`locked_until <= now` は自動解除済みとして未ロック扱い
- 1回のロック期間（`locked_until` が更新されてから自動解除までの間）に、ロック通知メールは1通だけ送る
- 以下のいずれかで `failed_login_count` を 0 に、`locked_until` を NULL に戻す
  - ログインに成功した（ロック中でなく、パスワードが一致した）とき
  - 自動解除後の最初のログイン試行で状態を評価したとき（試行自体の成否は別に判定する）
  - パスワードを再設定したとき
- 加算・ロック・ロック通知メールの送信要求は同一トランザクション（[ADR-024](../adr/ADR-024-transaction-management.md)）で行い、同じアカウントへの同時ログイン失敗でも加算漏れが起きないよう直列化する。送信要求は [SPEC-010](../spec/SPEC-010-email-notification.md) のアウトボックスに参加する
- 未登録のメールアドレスに対する失敗は数えない（登録済みの User に対してだけ加算する）

## 不変条件

- `failed_login_count >= 0`
- `locked_until` は NULL か、`created_at` より未来の時刻

## エラー

| エラー | 発生条件 |
|---|---|
| `ErrUserNotFound` | 指定した ID またはメールアドレスのユーザーが存在しない |
| `ErrEmailTaken` | 登録しようとしたメールアドレスがすでに使われている |
| `ErrInvalidEmail` | メールアドレスのフォーマットが不正 |
| `ErrPasswordTooShort` | パスワードが 8 文字未満 |
| `ErrPasswordTooLong` | パスワードが 128 文字を超えている |

## 関連ドキュメント

- [PRD-001: ユーザー認証](../prd/PRD-001-auth.md)
- [PRD-011: アカウントロック・パスワードリセット](../prd/PRD-011-account-lock-password-reset.md)
- [SPEC-001: ユーザー認証仕様](../spec/SPEC-001-auth.md)
- [SPEC-011: アカウントロック・パスワードリセット 仕様書](../spec/SPEC-011-account-lock-password-reset.md)
- [ADR-001: クリーンアーキテクチャ採用](../adr/ADR-001-clean-architecture.md)
- [ADR-024: トランザクション管理層](../adr/ADR-024-transaction-management.md)
- [Session ドメイン](./session.md)
- [PasswordResetToken ドメイン](./passwordresettoken.md)
