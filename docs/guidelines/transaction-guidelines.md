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
| Repository | Contextからトランザクション接続を取り出して使用する。なければプール接続を使う（複数テーブルにまたがる集約の保存は例外。下記参照） |
| Infrastructure | `TransactionManager` の実装・コンテキストへの接続格納を担う |

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

1つの集約を複数テーブルに保存するメソッド（例: `events` と `event_tags` に書き込む Event の `Create` / `Update`）は、Contextにトランザクションがあればそれを使い、なければ自前でトランザクションを張って集約を丸ごと保存する（`baseRepository.inTx`）。呼び出し側のUseCaseがトランザクションを張り忘れても、集約が中途半端に保存されないようにするため。

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
