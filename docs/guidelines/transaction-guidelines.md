# トランザクション管理ガイドライン

意思決定の背景は [ADR-024](../adr/ADR-024-transaction-management.md) を参照。

## 基本方針

- トランザクションは **Command系のUseCase** が管理する（ADR-022のCQRS方針と整合）
- UseCase層は `TransactionManager` インターフェース経由でトランザクションを制御し、pgxに直接依存しない
- トランザクション中の接続は `context.Context` で伝播させる
- Query系のUseCase（読み取り専用）はトランザクションを使用しない

## レイヤー別の責務

| レイヤー | 役割 |
|---|---|
| UseCase（Command） | `TransactionManager.RunInTx` を呼び出してトランザクション境界を定義する |
| Repository | Contextからトランザクション接続を取り出して使用する。なければプール接続を使う |
| Infrastructure | `TransactionManager` の実装・コンテキストへの接続格納を担う |

## インターフェース定義

`TransactionManager` インターフェースはUseCase層（またはドメイン層）に定義する。

```go
// internal/usecase/transaction.go
package usecase

import "context"

type TransactionManager interface {
    RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}
```

## Infrastructure層の実装

`pgxpool.Pool` を使ってトランザクションを開始し、接続をContextに格納する。

```go
// internal/infrastructure/db/transaction.go
package db

import (
    "context"
    "fmt"

    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
)

type txKey struct{}

type PgxTxManager struct {
    pool *pgxpool.Pool
}

func NewPgxTxManager(pool *pgxpool.Pool) *PgxTxManager {
    return &PgxTxManager{pool: pool}
}

func (m *PgxTxManager) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
    tx, err := m.pool.Begin(ctx)
    if err != nil {
        return fmt.Errorf("begin transaction: %w", err)
    }

    txCtx := context.WithValue(ctx, txKey{}, tx)
    if err := fn(txCtx); err != nil {
        _ = tx.Rollback(ctx)
        return err
    }

    if err := tx.Commit(ctx); err != nil {
        return fmt.Errorf("commit transaction: %w", err)
    }
    return nil
}

// GetTx はContextからトランザクション接続を取り出す。
// Repositoryのヘルパー関数として使用する。
func GetTx(ctx context.Context) (pgx.Tx, bool) {
    tx, ok := ctx.Value(txKey{}).(pgx.Tx)
    return tx, ok
}
```

## Repository実装

Contextにトランザクションがあればそちらを使い、なければプール接続を使う。

```go
// internal/infrastructure/repository/event_repository.go
package repository

import (
    "context"

    "github.com/jackc/pgx/v5/pgxpool"
    "yourproject/internal/infrastructure/db"
    "yourproject/internal/infrastructure/db/sqlc"
)

type EventRepository struct {
    pool    *pgxpool.Pool
    queries *sqlc.Queries
}

func NewEventRepository(pool *pgxpool.Pool) *EventRepository {
    return &EventRepository{
        pool:    pool,
        queries: sqlc.New(pool),
    }
}

// getQuerier はContextのトランザクションを優先して返す。
func (r *EventRepository) getQuerier(ctx context.Context) sqlc.DBTX {
    if tx, ok := db.GetTx(ctx); ok {
        return tx
    }
    return r.pool
}

func (r *EventRepository) Create(ctx context.Context, params sqlc.CreateEventParams) (*sqlc.Event, error) {
    q := sqlc.New(r.getQuerier(ctx))
    event, err := q.CreateEvent(ctx, params)
    if err != nil {
        return nil, err
    }
    return &event, nil
}
```

## UseCase（Command）の実装

`TransactionManager` をDIで受け取り、`RunInTx` でトランザクション境界を明示する。

```go
// internal/usecase/command/create_event.go
package command

import (
    "context"

    "yourproject/internal/domain/repository"
    "yourproject/internal/usecase"
)

type CreateEventCommand struct {
    // ...
}

type CreateEventUseCase struct {
    txManager       usecase.TransactionManager
    eventRepo       repository.EventRepository
    occurrenceRepo  repository.OccurrenceRepository
}

func NewCreateEventUseCase(
    txManager usecase.TransactionManager,
    eventRepo repository.EventRepository,
    occurrenceRepo repository.OccurrenceRepository,
) *CreateEventUseCase {
    return &CreateEventUseCase{
        txManager:      txManager,
        eventRepo:      eventRepo,
        occurrenceRepo: occurrenceRepo,
    }
}

func (uc *CreateEventUseCase) Execute(ctx context.Context, cmd CreateEventCommand) error {
    return uc.txManager.RunInTx(ctx, func(ctx context.Context) error {
        event, err := uc.eventRepo.Create(ctx, toEventParams(cmd))
        if err != nil {
            return err
        }

        for _, occ := range toOccurrenceParams(cmd, event.ID) {
            if _, err := uc.occurrenceRepo.Create(ctx, occ); err != nil {
                return err
            }
        }

        return nil
    })
}
```

## DIへの登録

`TransactionManager` の実装をDIコンテナに登録する。

```go
// internal/infrastructure/di/container.go
container.Provide(db.NewPgxTxManager)

// usecase.TransactionManager として解決されるようにする
container.Provide(func(m *db.PgxTxManager) usecase.TransactionManager {
    return m
})
```

## テストのモック

`TransactionManager` はインターフェースのため、UseCaseのユニットテストでモックに差し替えられる。

```go
// internal/usecase/command/create_event_test.go
type noopTxManager struct{}

func (m *noopTxManager) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
    return fn(ctx)
}

func TestCreateEventUseCase(t *testing.T) {
    uc := NewCreateEventUseCase(
        &noopTxManager{},
        &mockEventRepo{},
        &mockOccurrenceRepo{},
    )
    // ...
}
```

## チェックリスト

実装時に以下を確認する。

- [ ] Command系のUseCaseのみが `RunInTx` を呼び出しているか
- [ ] Query系のUseCaseが `RunInTx` を呼び出していないか
- [ ] Repository実装で `getQuerier`（またはそれに相当するヘルパー）を使ってContextのトランザクションを優先しているか
- [ ] `txKey` の型がexportされておらず、パッケージ外からContextに直接格納できない構造になっているか
- [ ] UseCaseのユニットテストで `noopTxManager`（または同等のモック）を使っているか
- [ ] エラー発生時にロールバックされるか（`RunInTx` 内でエラーを返せばロールバックされる）
