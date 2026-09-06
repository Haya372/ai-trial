# ADR-013: APIエラーレスポンス構造

- ステータス: 承認済

## コンテキスト

バックエンドAPIのエラーレスポンス構造が未定義のため、実装者ごとに異なるフォーマットが生まれるリスクがある。認証・予定CRUD・タグ管理などのエンドポイントを実装するにあたり、統一されたエラー構造を事前に決定・記録する必要がある。

検討にあたっての主な関心事は以下の通り。

- フロントエンドがエラー種別をプログラム的に判断できるか（コード体系）
- バリデーションエラーなど複数フィールドに跨るエラーを表現できるか
- Clean Architecture（ADR-001）のAdapter層でエラーをドメイン非依存に変換できるか
- 将来の仕様変更・拡張に対して後方互換性を保ちやすいか
- 内部実装の詳細が漏洩しないセキュリティ設計になっているか

## スコープ外

- エラーログの出力形式・ログ収集基盤の設計
- フロントエンド側のエラーハンドリングUI・UXの設計
- gRPCなどHTTP以外のプロトコルにおけるエラー表現
- 認証ライブラリ・バリデーションライブラリの選定

## 決定

APIエラーレスポンスは以下の独自JSONフォーマットを採用する。

```json
{
  "code": "VALIDATION_ERROR",
  "message": "Validation failed",
  "details": [
    {
      "field": "email",
      "code": "INVALID_EMAIL_FORMAT",
      "message": "Invalid email format"
    }
  ]
}
```

### フィールド定義

| フィールド | 型 | 必須 | 説明 |
|---|---|---|---|
| `code` | string | 必須 | マシン可読なエラーコード。大文字スネークケース（例: `NOT_FOUND`） |
| `message` | string | 必須 | エラーの説明。英語で記述する。表示言語の制御はフロントエンド（i18n）の責務 |
| `details` | array | 任意 | フィールドレベルのエラー情報。バリデーションエラー時に使用 |
| `details[].field` | string | 必須（details内） | エラーが発生したフィールド名 |
| `details[].code` | string | 必須（details内） | フィールドレベルのマシン可読エラーコード。大文字スネークケース（例: `INVALID_EMAIL_FORMAT`） |
| `details[].message` | string | 必須（details内） | フィールドに対するエラー説明。英語で記述する |

### セキュリティ制約

エラーレスポンスに含めてはならない情報を以下に定める。

**絶対に含めてはならない情報**

| 情報の種類 | 具体例 |
|---|---|
| スタックトレース | Goのランタイムエラー、`runtime/debug.Stack()` の出力 |
| DB・SQLの詳細 | SQLクエリ文字列、テーブル名、カラム名、DBエラーメッセージ |
| ファイルシステム情報 | ファイルパス、ディレクトリ構造 |
| ライブラリ・内部実装の詳細 | 使用ライブラリ名、バージョン、内部関数名 |
| 他ユーザーの情報 | 他ユーザーのID・メールアドレスなど |
| 認証情報の断片 | トークンの一部、パスワードハッシュ |

**`INTERNAL_ERROR` の扱い**

`500 Internal Server Error` を返す場合、`message` は常に固定文字列 `"Internal server error"` を使用する。実際のエラー内容はログにのみ記録し、レスポンスには一切含めない。

### メッセージの制御方針

`message` および `details[].message` はバックエンドが事前に定義した文言のみを使用する。ライブラリ・DB・バリデーターが返す内部エラーメッセージをそのまま渡してはならない。これは400系エラーについても同様。

**NG例（内部エラーをそのまま返す）**

```go
// 絶対にやってはいけない（500）
w.Write([]byte(`{"code": "INTERNAL_ERROR", "message": "` + err.Error() + `"}`))

// 絶対にやってはいけない（400: バリデーターの生メッセージをそのまま渡す）
// → "json: cannot unmarshal string into Go struct field User.Age of type int" などが漏洩する
writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", validator.Error())
```

**OK例（事前定義した文言を使う）**

```go
// 定義済みの文言を使う
writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Internal server error")

// フィールドごとに定義済みの文言をマッピングして返す
details := []ErrorDetail{
    {Field: "email", Message: "Invalid email format"},
}
writeValidationError(w, "Validation failed", details)
```

**言語について**

`message` および `details[].message` は英語で記述する。ユーザーへの表示言語の制御（i18n）はフロントエンドの責務とする。フロントエンドは `code` を使って独自のローカライズメッセージを定義してもよいし、バックエンドの英語 `message` をそのまま使ってもよい。

### レスポンス例

**バリデーションエラー（400）**

```json
{
  "code": "VALIDATION_ERROR",
  "message": "Validation failed",
  "details": [
    {
      "field": "email",
      "code": "INVALID_EMAIL_FORMAT",
      "message": "Invalid email format"
    },
    {
      "field": "password",
      "code": "TOO_SHORT",
      "message": "Password must be at least 8 characters"
    }
  ]
}
```

**リソース未発見（404）**

```json
{
  "code": "NOT_FOUND",
  "message": "Resource not found"
}
```

**認証エラー（401）**

```json
{
  "code": "UNAUTHORIZED",
  "message": "Authentication required"
}
```

**サーバー内部エラー（500）**

```json
{
  "code": "INTERNAL_ERROR",
  "message": "Internal server error"
}
```

## 検討した選択肢

### 選択肢1: RFC 7807 Problem Details（`application/problem+json`）

#### 概要

IETF RFC 7807で定義された標準的なエラー形式。`type`（URI）・`title`・`status`・`detail`・`instance` フィールドを持つ。

```json
{
  "type": "https://example.com/errors/validation-error",
  "title": "Validation Error",
  "status": 400,
  "detail": "リクエストのバリデーションに失敗しました",
  "instance": "/api/users"
}
```

#### メリット

- IETFによる標準化で相互運用性が高く、ツールやライブラリのサポートが期待できる
- `type` フィールドでエラー種別のドキュメントURLを提供できる
- Content-Type `application/problem+json` によりエラーレスポンスを明示的に識別できる

#### デメリット

- `type` フィールドにURIを使うため、エラー種別ごとにドキュメントURLを維持する運用コストが発生する（省略して文字列識別子を使う場合は標準の意味が薄れる）
- フィールドレベルのバリデーションエラーを表現するための拡張フィールド定義が別途必要で、RFC標準の範囲を超える
- プロジェクトの規模に対してやや重厚

### 選択肢2: 独自フォーマット（本決定）

#### 概要

`code`・`message`・`details` のシンプルな構造を独自定義する。

#### メリット

- フィールド数が少なく、フロントエンド・バックエンド双方の実装がシンプルになる
- `code` による機械可読な識別と `message` による人間可読な説明を明確に分離できる
- `details` 配列でバリデーションエラーのフィールドレベル情報を自然に表現できる
- URIベースの `type` フィールドを維持する運用コストがない
- Clean ArchitectureのAdapter層でドメインエラーをこのフォーマットに変換する処理がシンプルに保てる

#### デメリット

- RFC 7807非準拠のため、標準ツールとの相互運用性は期待できない
- エラーコード体系の設計・管理はプロジェクト側の責務になる

### 選択肢3: ハイブリッド（RFC 7807ベース＋独自フィールド）

#### 概要

RFC 7807の基本フィールドに `code` や `details` を拡張フィールドとして追加する。

```json
{
  "type": "validation_error",
  "title": "Validation Error",
  "status": 400,
  "detail": "リクエストのバリデーションに失敗しました",
  "code": "VALIDATION_ERROR",
  "details": [...]
}
```

#### メリット

- RFC 7807の骨格を持ちつつ独自ニーズに対応できる

#### デメリット

- RFC 7807準拠とも独自フォーマットとも言えない中途半端な状態になりやすい
- `type` と `code` の役割が重複してチーム内の混乱を招く
- フィールド数が増え、実装・テストの複雑さがどちらの選択肢より高くなる

## 決定理由

- **シンプルさと拡張性のバランス**: `code`・`message`・`details` の3フィールドは、現在のプロジェクト規模に対して必要十分。RFC 7807のURIベース `type` を維持する運用コストなしに、機械可読なエラー識別（`code`）・人間可読な説明（`message`）・フィールドレベル詳細（`details`）の3要件を満たせる
- **Clean Architectureとの整合**: Adapter層のエラーハンドラーでドメインエラーをこのフォーマットに変換する際、フィールドが少ないほどマッピング処理がシンプルになる。RFC 7807のハイブリッドは変換ロジックの複雑さを増やす
- **フロントエンドの実装容易性**: `code` フィールドの文字列でエラー種別を分岐できるため、フロントエンドのエラーハンドリングロジックが明確になる
- **エラーコード体系の自由度**: 独自コードにより、プロジェクト固有のドメインエラー（例: `SCHEDULE_OVERLAP`）を自然に追加できる。RFC 7807のURIはドキュメントページの維持が必要で、内部エラーコードの拡張に向かない

RFC 7807は外部公開APIや複数組織間の相互運用が必要なケースで真価を発揮する。本プロジェクトは特定のフロントエンドとバックエンドが1対1で連携する構成であり、標準化の恩恵よりシンプルさのメリットが上回る。

## 結果

### 良い影響

- バックエンド実装者がエラーレスポンスの構造に迷わなくなり、ハンドラーごとの実装ばらつきを防げる
- フロントエンドが `code` フィールドでエラー種別を判断できるため、エラー処理の実装が一貫する
- `details` 配列によりバリデーションエラーのフィールドレベル情報を単一レスポンスで表現できる
- セキュリティ制約の明文化により、内部実装の詳細やDB情報がレスポンスに漏洩するリスクを防げる
- `message` のバックエンド制御により、フロントエンドがメッセージをそのまま表示できる前提が成立する

### 悪い影響

- RFC 7807非準拠のため、APIの外部公開時に標準ツール（OpenAPI等のProblem Details対応クライアント）との互換性がない
- エラーコード（`code`の値）の追加・廃止はプロジェクト内で管理する必要があり、一覧の維持が必要
