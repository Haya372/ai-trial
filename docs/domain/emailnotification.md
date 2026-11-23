# EmailNotification ドメイン

## 概要

ユーザーの登録メールアドレス宛てに送る通知メール1通分の送信履歴を表すドメイン。アプリケーション内のユースケース（パスワードリセット・アカウントロック通知等）がメールの送信を要求すると、本ドメインの集約がまず「送信待ち」で永続化され、非同期に動く送信ワーカーが実際の送信・再送・結果記録を行う。

## 用語定義

| 用語 | 説明 |
|---|---|
| EmailNotification | 送信履歴1件を表す集約ルート。ID・ユーザーID・宛先アドレスのスナップショット・メール種別・状態・試行回数・時刻群・失敗理由・メッセージID・バウンス情報・本文用パラメータを持つ |
| EmailKind | メールの種別を表す値オブジェクト。種別の識別子をとる。具体的な値はメールの種類を追加する仕様書で定義する。各値には対応する `EmailTemplate` の登録が必須 |
| EmailStatus | 送信履歴の状態を表す値オブジェクト。`pending`（送信待ち）・`sent`（送信完了）・`failed`（送信失敗）・`bounced`（バウンス）のいずれか |
| EmailTemplate | メール種別ごとの件名・本文のテンプレート。パスワードの埋め込みは禁止。アプリケーション起動時に静的レジストリへ登録し、`EmailKind` をキーに取り出す |
| TemplateParams | 本文テンプレートに埋め込む1通分のパラメータ（リセット用URL等）。`pending` の間だけ EmailNotification に保持し、`sent` / `failed` / `bounced` のいずれかに遷移した時点で空にする |
| RecipientEmail | 送信要求時点のユーザー登録メールアドレスをスナップショットした宛先アドレス。以後、ユーザー側のアドレス変更の影響を受けない |
| MessageID | メール送信サービスがメール1通ごとに払い出す一意の識別子。`sent` 以降の履歴にだけ存在し、バウンス通知の突合に使う |
| BounceType | バウンスの種別を表す値オブジェクト。`permanent`（受信者側起因の恒久的なバウンス）・`transient`（一時的なバウンス）・`undetermined`（判別不能）のいずれか |
| 送信ワーカー (EmailSenderWorker) | バックグラウンドで `pending` の送信履歴をロック取得し、テンプレートに `TemplateParams` を埋めて送信サービスへ渡すコンポーネント。送信履歴の状態遷移と `TemplateParams` のクリアを担う |

## ビジネスルール

### 送信要求と集約の生成

- 送信要求は `UserID`・`EmailKind`・`TemplateParams` を取り、以下を満たすときだけ EmailNotification を生成できる
  - 指定した `UserID` のユーザーが存在すること（存在しなければ要求はエラーになり、集約は生成しない）
  - `EmailKind` に対応する `EmailTemplate` が登録されていること
- 生成直後の EmailNotification は以下を満たす
  - `status = pending`、`attempt_count = 0`、`sent_at = null`、`message_id = null`、`bounce_type = null`、`bounced_at = null`
  - `recipient_email` は要求時点のユーザーの登録アドレスのスナップショットで、以降変更しない
  - `requested_at = next_attempt_at = now`（即時に送信試行可）
  - `template_params` には要求で受け取ったパラメータを格納する
- EmailNotification の生成は、送信を要求した側のユースケースが張るトランザクション（`TransactionManager.RunInTx`）に参加し、業務データの永続化と送信履歴の作成を同一トランザクションで扱う（アウトボックス方式）

### 状態遷移

状態は `pending → sent` / `pending → failed` / `pending → pending` / `sent → bounced` のいずれかの遷移だけを許す。各遷移は送信ワーカー（またはバウンス通知）がトリガーになる。

| 遷移 | トリガー | 副作用 |
|---|---|---|
| `pending → pending`（再送待ち） | 送信試行が一時的な失敗で終わり、かつ `attempt_count` がまだ上限未満 | `attempt_count` を1加算、`last_failure_reason` を更新、`next_attempt_at` を次の試行時刻に進める（初回失敗から10秒後、2回目失敗から20秒後。詳細は仕様書を参照） |
| `pending → sent` | メール送信サービスがメッセージを受け付け、`message_id` を返した | `sent_at = now`、`message_id` を保存、`next_attempt_at = null`、`template_params = '{}'` |
| `pending → failed` | 送信試行が恒久的な失敗で終わった、または `attempt_count` が上限に達した一時的な失敗で終わった | `last_failure_reason` を更新、`next_attempt_at = null`、`template_params = '{}'` |
| `sent → bounced` | `message_id` と一致するバウンス通知を受信した | `bounce_type` を保存、`bounced_at = now` |

- `failed` および `bounced` は終端状態で、以後の送信・状態変化は行わない
- `template_params` は `sent` / `failed` / `bounced` のどれに遷移するときでも必ずクリアする（送信待ちの間だけ保持）
- `message_id` は `pending → sent` の遷移でだけ割り当てる。`failed` には `message_id` を保存しない

### 再送

- 試行回数は最大3回（初回1回 + 再送2回）。それ以上の再送は `failed` へ遷移させる
- 送信ワーカーは `status = pending AND next_attempt_at <= now` の履歴だけを取り出す
- 同時に複数のワーカーインスタンスが動く場合でも、1件の履歴を同時に取り出さないこと（取り出しは `FOR UPDATE SKIP LOCKED` などで排他する）

### 冪等性と重複送信

- 送信ワーカーがメール送信サービスへ渡した後、`status = sent` へ遷移させる前に停止した場合、再起動後に同じ履歴が再送される可能性がある。重複送信は許容し、未送信を防ぐことを優先する（仕様書の該当エッジケースと整合）
- バウンス通知は `message_id` で突合する。既に `bounced` の履歴に再度バウンス通知が届いた場合は無視する（最初の通知だけを記録する）
- `message_id` に一致する履歴がない場合、またはその履歴が `sent` 以外の場合は、通知を記録せず無視する（受信者から任意の `message_id` を送られても他の履歴に影響を与えないため）

### 不変条件

- `status = pending` のとき、`next_attempt_at` は NULL ではない
- `status <> pending` のとき、`next_attempt_at` は NULL
- `status = sent` または `status = bounced` のとき、`message_id` は NULL ではない
- `status = bounced` のとき、`bounce_type`・`bounced_at` は NULL ではない
- 任意の終端状態（`sent` / `failed` / `bounced`）のとき、`template_params` は空オブジェクト

## エラー

| エラー | 発生条件 |
|---|---|
| `ErrEmailRecipientNotFound` | 送信要求時の `UserID` に対応するユーザーが存在しない |
| `ErrEmailTemplateNotRegistered` | 指定した `EmailKind` に対応する `EmailTemplate` が登録されていない |
| `ErrEmailNotificationNotFound` | 指定した `message_id` に対応する EmailNotification が存在しない（バウンス突合） |

## 関連ドキュメント

- [PRD-010: メール通知基盤](../prd/PRD-010-email-notification.md)
- [SPEC-010: メール通知基盤 仕様書](../spec/SPEC-010-email-notification.md)
- [ADR-019: ENUM値の管理方針](../adr/ADR-019-enum-management.md)
- [ADR-024: トランザクション管理層](../adr/ADR-024-transaction-management.md)
- [ADR-038: メール通知の送信方式](../adr/ADR-038-email-transport-protocol.md)
- [ADR-039: メール通知の非同期送信とアウトボックス方式](../adr/ADR-039-email-async-outbox.md)
- [ADR-040: メール送信待ちパラメータの保存方式](../adr/ADR-040-email-template-params-storage.md)
- [ADR-041: バウンス通知受信エンドポイントの認証と内部形式](../adr/ADR-041-bounce-webhook-auth.md)
- [情報取り扱いガイドライン](../guidelines/data-handling-guidelines.md)
- [User ドメイン](./user.md)
