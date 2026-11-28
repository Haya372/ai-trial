# EmailNotification ドメイン

## 概要

ユーザーの登録メールアドレス宛てに送る通知メール1通分の送信履歴を表すドメイン。アプリケーション内のユースケースがメールの送信を要求すると、本ドメインの集約がまず「送信待ち」で永続化され、非同期に動く送信ワーカーが `EmailTransport` を通して送信・再送・結果記録を行う。

## 用語定義

| 用語 | 説明 |
|---|---|
| EmailNotification | 送信履歴1件を表す集約ルート。ID・ユーザーID・宛先アドレスのスナップショット・メール種別・状態・試行回数・時刻群・失敗コード・本文用パラメータを持つ |
| EmailKind | メールの種別を表す値オブジェクト。種別の識別子をとる。具体的な値はメールの種類を追加する仕様書で定義する。各値には対応する `EmailTemplate` の登録が必須 |
| EmailStatus | 送信履歴の状態を表す値オブジェクト。`pending`（送信待ち）・`sent`（送信完了）・`failed`（送信失敗）のいずれか |
| EmailTemplate | メール種別ごとの件名・本文のテンプレートと、埋め込むパラメータのキーの定義。パスワードの埋め込みは禁止。アプリケーション起動時に静的レジストリへ登録し、`EmailKind` をキーに取り出す。形式は [ADR-041](../adr/ADR-041-email-template-format.md) を参照 |
| TemplateParams | 本文テンプレートに埋め込む1通分のパラメータ（リセット用URL等）。`pending` の間だけ EmailNotification に保持し、`sent` / `failed` に遷移した時点で空にする |
| FailureCode | 送信失敗の原因を表す値オブジェクト。値と一時的・恒久的の分類は仕様書で定義する。送信先が返したエラー文は宛先アドレスを含みうるため保持しない |
| RecipientEmail | 送信要求時点のユーザー登録メールアドレスをスナップショットした宛先アドレス。以後、ユーザー側のアドレス変更の影響を受けない |
| EmailTransport | 件名・本文・宛先・差出人を受け取ってメールを送るインターフェース。結果を「受け付けた」「一時的な失敗」「恒久的な失敗」のいずれかで返す。送信サービス・送信プロトコルに依存する処理はこの実装に閉じ込める |
| 送信ワーカー (EmailSenderWorker) | バックグラウンドで `pending` の送信履歴をロック取得し、テンプレートに `TemplateParams` を埋めて `EmailTransport` へ渡すコンポーネント。送信履歴の状態遷移と `TemplateParams` のクリアを担う |

## ビジネスルール

### 送信要求と集約の生成

- 送信要求は `UserID`・`EmailKind`・`TemplateParams` を取り、以下を満たすときだけ EmailNotification を生成できる
  - 指定した `UserID` のユーザーが存在すること（存在しなければ要求はエラーになり、集約は生成しない）
  - `EmailKind` に対応する `EmailTemplate` が登録されていること
  - `TemplateParams` のキーが `EmailTemplate` の定義と一致すること（不足・余分がないこと）
- 生成直後の EmailNotification は以下を満たす
  - `status = pending`、`attempt_count = 0`、`sent_at = null`
  - `recipient_email` は要求時点のユーザーの登録アドレスのスナップショットで、以降変更しない
  - `requested_at = next_attempt_at = now`（即時に送信試行可）
  - `template_params` には要求で受け取ったパラメータを格納する
- EmailNotification の生成は、送信を要求した側のユースケースが張るトランザクション（`TransactionManager.RunInTx`）に参加し、業務データの永続化と送信履歴の作成を同一トランザクションで扱う（アウトボックス方式）

### 状態遷移

状態は `pending → sent` / `pending → failed` / `pending → pending` のいずれかの遷移だけを許す。各遷移は送信ワーカーがトリガーになる。

送信ワーカーは送信履歴を確保するときに `attempt_count` を1加算し、`next_attempt_at` を確保の有効期限（値は仕様書を参照）まで進めてコミットしてから送信する。確保時点で `attempt_count` がすでに3の場合は加算せず、同じトランザクションで `failed` へ遷移させる。以下の「`attempt_count`」は、確保時に加算した後の値を指す。

| 遷移 | トリガー | 副作用 |
|---|---|---|
| `pending → pending`（再送待ち） | `EmailTransport` が一時的な失敗を返し、かつ `attempt_count < 3` | `last_failure_code` を更新、`next_attempt_at` を次の試行時刻に進める（間隔は仕様書を参照） |
| `pending → sent` | `EmailTransport` がメールを受け付けた | `sent_at = now`、`last_failure_code = null`、`next_attempt_at = null`、`template_params = '{}'` |
| `pending → failed` | `EmailTransport` が恒久的な失敗を返した、`attempt_count = 3` で一時的な失敗を返した、テンプレートへの埋め込みに失敗した、または確保時点で `attempt_count` がすでに3だった（前回の試行中に停止した） | `last_failure_code` を更新、`next_attempt_at = null`、`template_params = '{}'` |

- `sent` および `failed` は終端状態で、以後の送信・状態変化は行わない

### 再送

- 試行回数は最大3回（初回1回 + 再送2回）。3回目の試行が一時的な失敗で終わったら `failed` へ遷移させる
- 送信結果は、確保したときの `attempt_count` と一致し、`status = pending` の場合だけ記録する。一致しない場合（確保の有効期限が切れて別のワーカーが確保した等）は記録しない
- 送信ワーカーは `status = pending AND next_attempt_at <= now` の履歴だけを取り出す
- 同時に複数のワーカーインスタンスが動く場合でも、1件の履歴を同時に確保しない（確保は `FOR UPDATE SKIP LOCKED` で排他し、確保済みの履歴は有効期限まで `next_attempt_at` が未来になるため取り出されない）

### 重複送信

- 送信ワーカーが `EmailTransport` へ渡した後、`status = sent` へ遷移させる前に停止した場合、確保の有効期限が切れたあとに同じ履歴が再送される可能性がある。重複送信は許容し、未送信を防ぐことを優先する（配送は少なくとも1回）

### 不変条件

- `status = pending` のとき、`next_attempt_at` は NULL ではない
- `status <> pending` のとき、`next_attempt_at` は NULL
- `status = sent` のとき、`sent_at` は NULL ではない
- `status = failed` のとき、`last_failure_code` は NULL ではない
- 終端状態（`sent` / `failed`）のとき、`template_params` は空オブジェクト

## エラー

| エラー | 発生条件 |
|---|---|
| `ErrEmailRecipientNotFound` | 送信要求時の `UserID` に対応するユーザーが存在しない |
| `ErrEmailTemplateNotRegistered` | 指定した `EmailKind` に対応する `EmailTemplate` が登録されていない |
| `ErrInvalidTemplateParams` | `TemplateParams` のキーが `EmailTemplate` の定義と一致しない |

## 関連ドキュメント

- [PRD-010: メール通知基盤](../prd/PRD-010-email-notification.md)
- [SPEC-010: メール通知基盤 仕様書](../spec/SPEC-010-email-notification.md)
- [ADR-019: ENUM値の管理方針](../adr/ADR-019-enum-management.md)
- [ADR-024: トランザクション管理層](../adr/ADR-024-transaction-management.md)
- [ADR-038: メール通知の非同期送信とアウトボックス方式](../adr/ADR-038-email-async-outbox.md)
- [ADR-039: メール送信待ちパラメータの保存方式](../adr/ADR-039-email-template-params-storage.md)
- [ADR-041: メールテンプレートの形式とテンプレートエンジン](../adr/ADR-041-email-template-format.md)
- [情報取り扱いガイドライン](../guidelines/data-handling-guidelines.md)
- [User ドメイン](./user.md)
