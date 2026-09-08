# ADR-019: ENUM値の管理方針

- ステータス: 承認済

## コンテキスト

DESIGN-001（ユーザー認証）の設計中に、DBのENUM値管理方式について複数の選択肢があることが判明した。
`audit_logs.entity_type` や `audit_logs.operation` 等、文字列の選択肢が限定されるカラムに適用される。

アプリケーション側での型安全性と、DB側での制約の強さ、および値追加時のコスト（マイグレーション要否）のトレードオフを整理し、プロジェクト全体の方針を決定する。

## スコープ外

- 既存カラムへのさかのぼった適用方針は、個別のリファクタリングIssueで決定する
- マスターテーブルが適切なケース（動的に値が増える・ユーザーが管理できる等）はこのADRの対象外とし、個別に判断する

## 決定

**VARCHAR + アプリ側定数管理**を採用する。DBにはENUM制約を持たせず、Goの定数（`const` / `iota`）でENUM値を一元管理する。

## 検討した選択肢

### 選択肢1: VARCHAR + アプリ側定数管理

#### 概要

カラム型は `VARCHAR` とし、DB側に制約を設けない。Goコード内で定数として値を定義し、アプリケーション層で制約を担保する。

```go
type EntityType string

const (
    EntityTypeUser    EntityType = "user"
    EntityTypeSession EntityType = "session"
)
```

#### メリット

- 値の追加がGoコードの変更のみで完結し、マイグレーションが不要
- DB依存がなく、テスト・ローカル開発環境のセットアップが単純
- Goの型システムで不正値の混入を防げる（型エイリアスを使用した場合）
- 値の変更・削除もマイグレーションなしで対応可能

#### デメリット

- DBレベルの制約がないため、アプリを介さない直接INSERT等で不正値が入る可能性がある
- DBのスキーマを見ただけでは許容値が分からない

### 選択肢2: PostgreSQL ENUM型

#### 概要

PostgreSQLの `CREATE TYPE ... AS ENUM` でENUM型を定義し、カラムに適用する。

```sql
CREATE TYPE entity_type AS ENUM ('user', 'session');
ALTER TABLE audit_logs ADD COLUMN entity_type entity_type NOT NULL;
```

#### メリット

- DB側で不正値を完全に排除できる
- スキーマを見るだけで許容値が明確

#### デメリット

- 値の追加には `ALTER TYPE ... ADD VALUE` マイグレーションが必要
- `ALTER TYPE` はトランザクション内で扱いが制限される（PostgreSQL固有の制約）
- golang-migrate で管理する場合、down マイグレーションで値を削除できないケースがある
- 型名がDB全体のグローバル名前空間に存在するため、命名衝突リスクがある

### 選択肢3: CHECK制約

#### 概要

カラム型は `VARCHAR` とし、`CHECK (column IN (...))` でDB制約を設ける。

```sql
ALTER TABLE audit_logs ADD COLUMN entity_type VARCHAR NOT NULL
    CHECK (entity_type IN ('user', 'session'));
```

#### メリット

- DB側で不正値を排除できる
- PostgreSQL ENUM型のトランザクション制限がない
- 値が許容されるかをスキーマから確認できる

#### デメリット

- 値の追加・変更のたびに `ALTER TABLE` マイグレーションが必要
- 許容値リストが長くなるとスキーマの可読性が下がる
- マイグレーションの頻度が上がるほど運用コストが増大する

### 選択肢4: マスターテーブル

#### 概要

ENUM値をマスターテーブルとして管理し、FK制約で参照する。

```sql
CREATE TABLE entity_types (id VARCHAR PRIMARY KEY);
INSERT INTO entity_types VALUES ('user'), ('session');

ALTER TABLE audit_logs ADD COLUMN entity_type VARCHAR NOT NULL
    REFERENCES entity_types(id);
```

#### メリット

- 値の追加がINSERTのみで完結し、スキーマ変更不要
- FK制約でDB側の整合性を担保できる

#### デメリット

- クエリにJOINが増え、複雑になる
- マスターテーブルの初期データ投入・管理が必要
- `audit_logs.entity_type` のような固定的な値には過剰な設計となる

## 決定理由

- **値の追加コストが最小**: アプリケーション開発中は仕様変更により新しい値が頻繁に追加される。PostgreSQL ENUMやCHECK制約はそのたびにマイグレーションが必要となり、開発速度を下げる
- **Goの型システムで十分な安全性を担保できる**: 型エイリアス（`type EntityType string`）を使うことで、コンパイル時に不正値の混入を防げる。直接DBを操作する機会が限られた本プロジェクトでは、DBレベルの制約が必須ではない
- **PostgreSQL ENUMのトランザクション制限を回避**: `ALTER TYPE ... ADD VALUE` はトランザクション内で扱いが困難で、golang-migrate のdown マイグレーションとの相性も悪い
- **CHECK制約はマイグレーション運用コストが高い**: 値を追加するたびに `ALTER TABLE` が必要で、ENUMに比べてメリットが薄い
- **マスターテーブルは過剰設計**: `audit_logs.entity_type` のような静的な分類値にFKとJOINを導入するのは複雑さに見合わない

## 使い分けガイド

新しいカラムを設計するとき、以下のフローで管理方式を選ぶ。

```
値がユーザー・管理者によって動的に増える、または値ごとに属性（表示名・有効フラグ等）が必要か？
  → YES: マスターテーブル
  → NO ↓

仕様で固定された値か（コードリリースなしに変えることがない）？
  → YES: VARCHAR + Goの定数管理（本ADRの決定）
  → NO ↓ （固定値だが変更頻度が高い等、判断が難しい場合）

チームで議論してIssueを立てる
```

### 判断基準の早見表

| 条件 | 推奨 |
|---|---|
| 値が仕様で固定（`entity_type`, `operation` 等） | VARCHAR + Goの定数管理 |
| 値をアプリ外（管理画面・SQLコンソール）から追加したい | マスターテーブル |
| 値ごとに表示名・並び順・有効フラグ等の属性が必要 | マスターテーブル |
| 複数テーブルから同じ値セットを参照する | マスターテーブルを検討 |

PostgreSQL ENUM型とCHECK制約は、マイグレーション運用コストが高いためこのプロジェクトでは原則採用しない。

## 結果

### 良い影響

- ENUM値の追加がGoコードの変更のみで完結し、マイグレーションが不要
- Goの型エイリアスによりコンパイル時に型安全性が担保される
- テスト・ローカル環境のDBセットアップが単純になる
- PostgreSQL ENUM型のトランザクション制限を踏まない

### 悪い影響

- DBスキーマだけを見ても許容値が分からないため、Goコードと合わせて確認する必要がある
- アプリを介さない直接INSERTでは不正値が入る可能性があるため、開発時のDB操作ルールを整備する必要がある
