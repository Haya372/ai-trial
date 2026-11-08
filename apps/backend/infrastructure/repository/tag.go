package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/trace"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/domain/tag"
	"github.com/Haya372/ai-trial/backend/infrastructure/db"
	query "github.com/Haya372/ai-trial/backend/infrastructure/db/generated"
)

const tagsTable = "tags"

// errNoTransaction is returned by the locking methods when ctx carries no
// transaction: outside one, the row lock would be released as soon as the
// statement ends and silently protect nothing.
var errNoTransaction = errors.New("must be called inside a transaction")

func tagFromRow(row query.Tag) (tag.Tag, error) {
	name, err := tag.NewTagName(row.Name)
	if err != nil {
		return nil, fmt.Errorf("reconstruct tag name: %w", err)
	}
	color, err := tag.NewTagColor(row.Color)
	if err != nil {
		return nil, fmt.Errorf("reconstruct tag color: %w", err)
	}
	return tag.New(uuid.UUID(row.ID.Bytes), uuid.UUID(row.UserID.Bytes), name, color), nil
}

type tagRepository struct {
	baseRepository

	logger *slog.Logger
}

// NewTagRepository returns a tag.Repository backed by PostgreSQL.
func NewTagRepository(pool *pgxpool.Pool, logger *slog.Logger, tp trace.TracerProvider) tag.Repository {
	return &tagRepository{baseRepository{pool: pool, tracer: tp.Tracer(tracerName)}, logger}
}

func (r *tagRepository) Create(ctx context.Context, t tag.Tag) (tag.Tag, error) {
	spanCtx, span := r.startDBSpan(ctx, "INSERT", tagsTable)
	row, err := r.querier(spanCtx).InsertTag(spanCtx, query.InsertTagParams{
		ID:     toPgUUID(t.ID()),
		UserID: toPgUUID(t.UserID()),
		Name:   string(t.Name()),
		Color:  string(t.Color()),
	})
	endDBSpan(span, err)
	if err != nil {
		if isUniqueViolation(err) {
			return nil, tag.ErrTagNameTaken
		}
		r.logger.Error("insert tag query failed", "error", err)
		return nil, fmt.Errorf("insert tag: %w", err)
	}
	return tagFromRow(row)
}

func (r *tagRepository) FindByID(ctx context.Context, id uuid.UUID) (tag.Tag, error) {
	spanCtx, span := r.startDBSpan(ctx, "SELECT", tagsTable)
	row, err := r.querier(spanCtx).FindTagByID(spanCtx, toPgUUID(id))
	endDBSpanNotFound(span, err)
	if notFound, wasNoRows, mappedErr := checkNotFound(err, tag.ErrTagNotFound); notFound {
		if wasNoRows {
			return nil, mappedErr
		}
		r.logger.Error("find tag by id query failed", "error", mappedErr)
		return nil, fmt.Errorf("find tag by id: %w", mappedErr)
	}
	return tagFromRow(row)
}

func (r *tagRepository) Update(ctx context.Context, t tag.Tag) error {
	spanCtx, span := r.startDBSpan(ctx, "UPDATE", tagsTable)
	err := r.querier(spanCtx).UpdateTag(spanCtx, query.UpdateTagParams{
		ID:    toPgUUID(t.ID()),
		Name:  string(t.Name()),
		Color: string(t.Color()),
	})
	endDBSpan(span, err)
	if err != nil {
		if isUniqueViolation(err) {
			return tag.ErrTagNameTaken
		}
		r.logger.Error("update tag query failed", "error", err)
		return fmt.Errorf("update tag: %w", err)
	}
	return nil
}

func (r *tagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	spanCtx, span := r.startDBSpan(ctx, "DELETE", tagsTable)
	err := r.querier(spanCtx).DeleteTag(spanCtx, toPgUUID(id))
	endDBSpan(span, err)
	if err != nil {
		r.logger.Error("delete tag query failed", "error", err)
		return fmt.Errorf("delete tag: %w", err)
	}
	return nil
}

func (r *tagRepository) ExistsByName(
	ctx context.Context, userID uuid.UUID, name tag.TagName, excludeID uuid.UUID,
) (bool, error) {
	spanCtx, span := r.startDBSpan(ctx, "SELECT", tagsTable)
	exists, err := r.querier(spanCtx).ExistsTagByName(spanCtx, query.ExistsTagByNameParams{
		UserID:    toPgUUID(userID),
		Name:      string(name),
		ExcludeID: toPgUUID(excludeID),
	})
	endDBSpan(span, err)
	if err != nil {
		r.logger.Error("exists tag by name query failed", "error", err)
		return false, fmt.Errorf("exists tag by name: %w", err)
	}
	return exists, nil
}

func (r *tagRepository) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	spanCtx, span := r.startDBSpan(ctx, "SELECT", tagsTable)
	count, err := r.querier(spanCtx).CountTagsByUserID(spanCtx, toPgUUID(userID))
	endDBSpan(span, err)
	if err != nil {
		r.logger.Error("count tags query failed", "error", err)
		return 0, fmt.Errorf("count tags: %w", err)
	}
	return int(count), nil
}

func (r *tagRepository) LockOwner(ctx context.Context, userID uuid.UUID) error {
	if _, ok := db.GetTx(ctx); !ok {
		return fmt.Errorf("lock tag owner: %w", errNoTransaction)
	}
	spanCtx, span := r.startDBSpan(ctx, "SELECT", usersTable)
	_, err := r.querier(spanCtx).LockUserForUpdate(spanCtx, toPgUUID(userID))
	endDBSpan(span, err)
	if err != nil {
		r.logger.Error("lock tag owner query failed", "error", err)
		return fmt.Errorf("lock tag owner: %w", err)
	}
	return nil
}

func (r *tagRepository) FindByIDsForShare(ctx context.Context, ids []uuid.UUID) ([]tag.Tag, error) {
	if _, ok := db.GetTx(ctx); !ok {
		return nil, fmt.Errorf("find tags for share: %w", errNoTransaction)
	}
	if len(ids) == 0 {
		return []tag.Tag{}, nil
	}

	pgIDs := make([]pgtype.UUID, len(ids))
	for i, id := range ids {
		pgIDs[i] = toPgUUID(id)
	}
	spanCtx, span := r.startDBSpan(ctx, "SELECT", tagsTable)
	rows, err := r.querier(spanCtx).FindTagsByIDsForShare(spanCtx, pgIDs)
	endDBSpan(span, err)
	if err != nil {
		r.logger.Error("find tags for share query failed", "error", err)
		return nil, fmt.Errorf("find tags for share: %w", err)
	}

	tags := make([]tag.Tag, 0, len(rows))
	for _, row := range rows {
		t, err := tagFromRow(row)
		if err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, nil
}
