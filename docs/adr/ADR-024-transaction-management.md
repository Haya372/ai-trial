# ADR-024: トランザクション管理層

- ステータス: 承認済

## コンテキスト

ADR-001でClean Architectureを採用し、ADR-009でsqlc + pgxをDBアクセスライブラリとして採用した。ADR-022の論理CQRSにより、UseCase層はCommandとQueryに分離している。

複数のリポジトリにまたがる書き込み操作（例: 予定作成時にEventとOccurrenceの両テーブルへ書き込む）では、DBトランザクションで原子性を保証する必要がある。この「どの層でトランザクションを開始・コミット・ロールバックするか」を決定する。

Clean Architectureのルールから、UseCase層はインフラ層（pgx）に直接依存できない。したがって、pgxの `pgx.Tx` をUseCaseが直接扱う設計はDependency Rule（ADR-001）に違反する。

## スコープ外

- クエリ（読み取り専用）操作のトランザクション管理（Queryは副作用を持たないためトランザクションは不要）
- コネクションプールのパラメータチューニング
- 分散トランザクション・サガパターン（外部サービスをまたぐ整合性）
- リードレプリカ・読み取り専用DB接続の管理

## 決定

**UseCase層のCommandがトランザクションを管理する。** ただしpgxへの直接依存を避けるため、インターフェースを介した **TransactionManager パターン** を採用する。

- UseCase層に `TransactionManager` インターフェースを定義する
- Infrastructure層に実装（`pgxTxManager`）を配置し、pgxトランザクションを管理する
- トランザクション中の接続はGoの `context.Context` にキーで格納して伝播させる
- Repositoryは `context.Context` からトランザクション接続を取り出して使用し、なければプールを使う
- Command系のUseCaseのみがトランザクションを扱い、Query系は扱わない（ADR-022と整合）

## 検討した選択肢

### 選択肢1: Repository層でトランザクションを管理する

#### 概要

各Repositoryが内部でトランザクションを開始・コミット・ロールバックする。

#### メリット

- Repositoryが自己完結しており実装がシンプル
- UseCaseがトランザクションを意識しなくてよい

#### デメリット

- 複数Repositoryにまたがる操作（例: EventとOccurrenceを同時に作成）で原子性を保証できない
- 「複数リポジトリをまたぐトランザクション」が必要な業務要件（PRD-009等）に対応できない

### 選択肢2: UseCase層がpgx.Txを直接扱う

#### 概要

UseCase層が `pgx.Tx` または `pgxpool.Pool` を直接受け取り、トランザクションを開始・コミットする。

#### メリット

- 追加の抽象化が不要でコード量が少ない

#### デメリット

- UseCase層がインフラ依存（pgx）を持つため、ADR-001のDependency Ruleに違反する
- Repository実装をモックに差し替えることができず、UseCaseのユニットテストが困難になる

### 選択肢3: UseCase層にTransactionManagerインターフェースを定義し、Context経由で伝播する（採用）

#### 概要

UseCase層に抽象インターフェース `TransactionManager` を定義し、Infrastructure層が実装する。
トランザクション中の `pgx.Tx` は `context.Context` にキーで格納し、Repositoryがコンテキストから取り出して使う。

#### メリット

- UseCase層はインフラ依存を持たず、ADR-001のDependency Ruleを守れる
- Repositoryはトランザクション有無を問わず同じ実装で動作する（コンテキストに接続があればそちらを使い、なければプールを使う）
- `TransactionManager` のモックを注入できるため、UseCaseのユニットテストが容易
- 将来的に実装を差し替えてもUseCase層・Repository層は変更不要

#### デメリット

- インターフェースの定義・実装・コンテキストキー管理の定型コードが増える
- `context.Context` を型アサーションで取り出す部分がInfrastructure実装に閉じるが、コンテキストへの格納パターンへの理解が必要

### 選択肢4: Unit of Workパターン

#### 概要

複数Repositoryをひとつの `UnitOfWork` オブジェクトにまとめ、同一トランザクションを共有させる。

#### メリット

- トランザクションの境界が明示的でコードが読みやすい

#### デメリット

- 機能追加のたびに `UnitOfWork` 構造体にリポジトリを追加する必要があり、結合度が上がる
- Goのコンテキストモデルと相性が悪く、定型コードが多い
- 選択肢3と比べてメリットが少なく、複雑さが増す

## 決定理由

- **Dependency Ruleの遵守**: UseCase層がpgxに直接依存する選択肢2はClean Architectureの根幹に反する。インターフェース経由にすることで依存の方向を逆転させ、ADR-001の方針を維持できる
- **複数リポジトリにまたがる原子性**: 選択肢1（Repository層管理）は単一リポジトリで完結する操作には十分だが、予定作成のようにEventとOccurrenceを同時に書き込む要件では原子性を保証できない。UseCaseがトランザクション境界を決める設計が必要
- **テスタビリティ**: `TransactionManager` をインターフェース化することで、UseCaseのユニットテスト時にモックに差し替えられる。選択肢2はpgxへの直接依存によりモック化が困難
- **CQRS整合**: ADR-022のCommandとQueryの分離に従い、トランザクション管理はCommandのみが担う。Queryは読み取り専用のためトランザクションは不要
- **Unit of Workを選ばない理由**: GoではContextを使ったトランザクション伝播が慣用的で、UnitOfWorkよりもコード量・結合度・保守性の面で選択肢3が優る

## 結果

### 良い影響

- UseCase層がインフラ依存を持たないため、DBやORMを差し替えても上位レイヤーの変更が不要
- `TransactionManager` をモックに差し替えることでUseCaseのユニットテストがDBなしで実行できる
- RepositoryはContextの有無を問わず同じインターフェースで動作し、テスト・本番・トランザクション有無を透過的に扱える
- Command系のUseCaseがトランザクション境界を明示するため、コードを読むだけでどの操作が原子かを把握できる

### 悪い影響

- `TransactionManager` インターフェース・pgx実装・コンテキストキー定義など定型コードが増える
- Repository実装でContextから接続を取り出す処理（`getTx` ヘルパー等）が必要になる
- コンテキストへのトランザクション格納は型アサーションを伴うため、キーの一元管理と命名規則の徹底が必要
