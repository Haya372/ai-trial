# RecurringEventSeries ドメイン

## 概要

繰り返し予定の1まとまり（シリーズ）を表すドメイン。繰り返しパターン・終了条件・シリーズ作成時の端末タイムゾーン・除外日集合を持ち、シリーズに属する各回を [Event](./event.md) として束ねる。
Event とは別集約のルートとして扱い、Event 側は RecurringEventSeries の ID を参照する。
シリーズの各回は物理的に Event として保存する方式（実体化方式）を採る。選定理由は [SPEC-009](../spec/SPEC-009-recurring-events.md) を参照。

## 用語定義

| 用語 | 説明 |
|---|---|
| RecurringEventSeries（シリーズ） | 繰り返し予定の1まとまりを表す集約ルート。繰り返しルール・終了条件・タイムゾーン・除外日集合を持つ |
| 回 (Occurrence) | シリーズから生成される予定の1回。Event として実体化し、タイトル・日時・場所・メモ・タグなどを個別に持つ |
| RecurrencePattern | 繰り返しパターンを表す値オブジェクト。`daily` / `weekly` / `biweekly` / `monthly_day` / `monthly_nth_weekday` の5種のいずれかで、バリアントごとに必要な付随パラメータ（曜日・日付・第N週）を持つ |
| RecurrenceEnd | 終了条件を表す値オブジェクト。`indefinite`（無期限）／`until_date`（終了日指定、SeriesTimezone のローカル日付）／`count`（回数指定、整数）のいずれかを取る |
| MonthlyNth | 「毎月 第N曜日」の N を表す値オブジェクト。`1` / `2` / `3` / `4` / `last` のいずれかを取る |
| SeriesTimezone | シリーズ作成時の端末タイムゾーン（IANA 名、例: `Asia/Tokyo`）を表す値オブジェクト。各回の壁時計時刻（開始・終了）の計算と除外日の解釈に使う |
| 除外日 (ExcludedDate) | パターンから本来生成される日のうち、生成・保持しない日を表す値オブジェクト。SeriesTimezone のローカル日付（DATE）で保持する |
| 切り離し (Detach) | 「この予定のみ変更」で、対象回の Event をシリーズとの関連から外し、単発の予定にする操作。シリーズ側にはその回の日付が除外日として残る |

## ビジネスルール

### 集約の境界

- RecurringEventSeries は Event とは別集約のルートとする。シリーズの各回は Event 集約で実体化し、シリーズへの所属は Event からの参照（`series_id`）で表現する
- RecurringEventSeries 集約は、シリーズ自身の属性（所有ユーザー・RecurrencePattern・RecurrenceEnd・SeriesTimezone）と除外日集合（ExcludedDate の集合）を自分の内側に保持する
- Event 側の生存・編集・削除・タグ付与・共有は、シリーズとの関連の有無にかかわらず Event 集約として完結する。切り離された回も単発の Event として存在し続ける
- 複数の Event と RecurringEventSeries を同時に操作する場面（シリーズ作成・「以降すべて」「すべて」の変更/削除・切り離し）は、UseCase 層でトランザクション境界を張る（[ADR-024](../adr/ADR-024-transaction-management.md)）

### シリーズの構成

- シリーズは所有ユーザー1人に属し、そのユーザーだけが変更・削除できる
- シリーズは RecurrencePattern と RecurrenceEnd をちょうど1つずつ持つ
- RecurrencePattern の付随パラメータ（曜日・日付・第N週）はシリーズの最初の回の壁時計日時から確定する。パターン・終了条件を編集で変更した場合は、既存シリーズをそのまま更新するのではなく、「以降すべて変更」は別シリーズの新規作成、「すべて変更」は同じシリーズ ID のまま除外日集合と回を破棄して再生成、として扱う
- SeriesTimezone はシリーズ作成時の端末タイムゾーン（IANA 名）を1つ保持し、以後変更しない
- 各回の壁時計時刻（開始・終了）はシリーズ内で同一とし、SeriesTimezone のもとで計算する。夏時間の切り替えをまたいでも壁時計時刻は変えない（UTC の瞬間時刻はずれる）
- 各回の所要時間（終了 − 開始）はシリーズ内で同一とする

### 期間と回数

- シリーズに含める回は、最初の回の開始日から2年以内（最初の回を含み、2年後の同日を超えない）に開始するものだけとする
- RecurrenceEnd.Count（回数指定）は 2 以上の整数とし、除外日・スキップされた月の回は回数に含めない
- RecurrenceEnd.UntilDate は最初の回の開始日以降、かつ最初の回から2年以内のローカル日付とする（終了日当日に開始する回までを含む）
- ユーザーが繰り返し設定を入力してシリーズを作る場合（新規作成・単発からの変換・繰り返し設定を変更する「以降すべて変更」「すべて変更」）は、2回以上の回を含まなければならない。0 回・1 回しか候補がないパターン・終了条件の組み合わせは生成できない
- 「以降すべて変更」「以降すべて削除」「この予定のみ」の操作の結果、シリーズの回が1回になることは許容する
- 繰り返し設定を変更しない「以降すべて変更」で、元のシリーズの RecurrenceEnd が `count` の場合、新しいシリーズの Count は対象回を含む残りの回数とする

### 除外日・切り離し

- 除外日は SeriesTimezone のローカル日付で保持し、1シリーズ内で重複しない
- 「この予定のみ削除」は、対象回の Event を削除してシリーズに同じ日付の除外日を追加する
- 「この予定のみ変更」（切り離し）は、対象回の Event からシリーズへの関連（`series_id`）を外し、変更後の内容で単発の予定として残す。あわせてシリーズに同じ日付の除外日を追加する
- 切り離した回は、以降の「以降すべて」「すべて」の変更・削除の対象に含まれない

### 範囲変更・範囲削除

- 「以降すべて変更」は、対象回の直前でシリーズを終了させ（RecurrenceEnd を対象回の前日までの `until_date` に書き換え、対象回以降の回と除外日を取り除く）、対象回以降を新しいルールの別シリーズとして扱う
- 「すべて変更」でパターン・終了条件を変更した場合、シリーズの既存の回と除外日を破棄し、新しいルールで回を再生成する
- 「すべて変更」でタイトル・メモ・場所・URL・タグだけを変更した場合、シリーズに属するすべての回の該当フィールドを同じ値に揃える
- 「すべて変更」で開始日時だけを変更した場合、シリーズに属するすべての回の開始・終了日時を変更前後の差分だけずらす
- 「すべて削除」はシリーズとシリーズに属する回・除外日を一括で削除する
- シリーズに属する回が「この予定のみ削除」「この予定のみ変更」等で 0 件になった場合、シリーズも削除する

### タグとの関係

- タグ付与はシリーズ単位の意味論を持ち、同じシリーズに属するすべての回で Tag 集合（付与順を含む）を揃える。保存は従来どおり各回の [EventTag](./event.md)（Event 集約配下）に行う
- シリーズのタグを変更する操作（「すべて変更」）では、シリーズに属するすべての回の EventTag を同じ内容に揃える（UseCase 層が保証する）
- 切り離された回は、切り離し時点で付いていた EventTag をそのまま保持する。シリーズ側のタグが後に変更されても、切り離された回のタグには影響しない

## エラー

| エラー | 発生条件 |
|---|---|
| `ErrRecurringSeriesNotFound` | 指定した ID のシリーズが存在しない、または呼び出しユーザーに属していない |
| `ErrRecurringPatternInvalid` | RecurrencePattern のパラメータが不正（`monthly_day` の日付が 1〜31 の範囲外、`monthly_nth_weekday` の Nth が `1`〜`4` / `last` のいずれでもない、曜日が 0〜6 の範囲外、など） |
| `ErrRecurringEndInvalid` | RecurrenceEnd のパラメータが不正（`count` が 2 未満、`until_date` が最初の回の開始日より前、`until_date` が最初の回の開始日から2年を超える、など） |
| `ErrRecurringTimezoneInvalid` | SeriesTimezone の値が IANA タイムゾーン名として解釈できない |
| `ErrRecurringSeriesTooFewOccurrences` | シリーズに含まれる候補回が 2 回未満になる（例: 無期限・終了日指定で 2 年以内に 2 回目の候補が発生しないパターンと開始日時の組み合わせ） |
| `ErrRecurringSeriesSpanExceeded` | シリーズに含めようとした回のうち、最初の回の開始日から 2 年を超える位置に開始するものがある |
| `ErrRecurringExcludedDateDuplicated` | 同一の除外日を同一シリーズに追加しようとした |

## 関連ドキュメント

- [PRD-009: 繰り返し予定](../prd/PRD-009-recurring-events.md)
- [SPEC-009: 繰り返し予定仕様](../spec/SPEC-009-recurring-events.md)
- [Event ドメイン](./event.md)
- [Tag ドメイン](./tag.md)
- [EventShare ドメイン](./eventshare.md)
- [EventSubscription ドメイン](./eventsubscription.md)
- [ADR-022: CQRS方針](../adr/ADR-022-cqrs.md)
- [ADR-024: トランザクション管理層](../adr/ADR-024-transaction-management.md)
- [ADR-036: APIで扱う日時のタイムゾーン](../adr/ADR-036-api-datetime-utc.md)
- [ADR-037: 繰り返し予定の日時計算をサーバーで行う](../adr/ADR-037-recurrence-timezone-calculation.md)
