# Tag ドメイン

## 概要

ユーザーが予定を分類するために作成する個人用ラベルを表すドメイン。各タグは所有ユーザー・タグ名・色を持ち、予定（Event）に付与されてカレンダー上の色分け表示と絞り込みに使われる。タグは個人用の設定として扱い、他ユーザーとは共有・参照しない。

## 用語定義

| 用語 | 説明 |
|---|---|
| Tag | ユーザーが作成するタグ1件を表すエンティティ。ID・所有ユーザーID・タグ名・色・作成/更新日時を持つ |
| TagName | タグの表示名を表す値オブジェクト。前後空白を除去した後に 1〜20 文字であること |
| TagColor | タグの色を表す値オブジェクト。プリセット12色の HEX 6 文字（大文字、`#` 無し）のいずれかの値を取る |
| プリセット12色 | 文字とのコントラスト比を確保した 12 色のパレット。具体の RGB 値は OpenAPI スキーマの `TagColor` enum で一元管理する |
| EventTag | 予定（Event）へのタグ付与 1 件を表す関連エンティティ。Event 集約配下に置き、Position（1〜5）で付与順を保持する |
| 付与順 (Position) | 1 つの予定に複数のタグを付与したときの並び順。カレンダー上の背景色は Position=1 のタグ、色ドットは Position=2 以降のタグで表現する |

## ビジネスルール

- タグは所有ユーザー個人の設定であり、他ユーザーとは共有・参照できない
- TagName は空文字（前後空白除去後）を許容しない
- TagName は所有ユーザー内で一意でなければならない。一意性判定では大文字・小文字を区別しない
- 1 ユーザーが保有できる Tag は最大 50 件とする
- TagColor はプリセット12色のいずれかでなければならず、任意色は許容しない
- 異なる Tag に同じ色を付けてよい（色の重複は制限しない）
- EventTag は、紐づく Tag の所有ユーザーが紐づく Event の所有ユーザーと一致するものだけ生成できる（他ユーザーのタグを自分の予定に付与できない）
- 1 件の Event に付与できる Tag は最大 5 件、同一 Tag の重複付与は不可
- 1 件の Event 内で Position は 1〜5 の範囲で一意であり、Position の昇順を付与順とする（Tag の削除で欠番が生じても順序は保たれる）
- Tag が削除された場合、その Tag を参照する全ての EventTag も削除される（Event 自体は削除しない）
- 所有ユーザーが削除された場合、そのユーザーの Tag および紐づく EventTag も削除される

## エラー

| エラー | 発生条件 |
|---|---|
| `ErrTagNotFound` | 指定した ID の Tag が存在しない、または呼び出しユーザーに属していない |
| `ErrTagNameEmpty` | TagName が空文字、または前後空白除去後に空 |
| `ErrTagNameTooLong` | TagName が 20 文字を超えている |
| `ErrTagNameTaken` | 同一ユーザー内に同名（大小区別なし）の Tag が既に存在する |
| `ErrTagLimitExceeded` | ユーザーの保有 Tag が既に 50 件ある状態で新規作成しようとした |
| `ErrTagColorInvalid` | TagColor の値がプリセット12色のいずれでもない |
| `ErrEventTagLimitExceeded` | 1 Event に付与しようとした Tag が 5 件を超える |
| `ErrEventTagDuplicated` | 1 Event への付与指定で同一 Tag ID が複数回現れた |
| `ErrEventTagForeign` | Event の所有者と異なるユーザーの Tag を付与しようとした |

## 関連ドキュメント

- [PRD-008: タグ＆カラーラベル](../prd/PRD-008-tags-labels.md)
- [SPEC-008: タグ＆カラーラベル仕様](../spec/SPEC-008-tags-labels.md)
- [Event ドメイン](./event.md)
- [User ドメイン](./user.md)
- [ADR-013: APIエラーレスポンス構造](../adr/ADR-013-api-error-structure.md)
- [ADR-016: スキーマ駆動開発](../adr/ADR-016-schema-driven-development.md)
- [ADR-019: ENUM値の管理方針](../adr/ADR-019-enum-management.md)
- [ADR-022: CQRS方針](../adr/ADR-022-cqrs.md)
- [ADR-024: トランザクション管理層](../adr/ADR-024-transaction-management.md)
