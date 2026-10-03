package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/google/uuid"

	"github.com/Haya372/ai-trial/backend/infrastructure/db"
	query "github.com/Haya372/ai-trial/backend/infrastructure/db/generated"
)

// toPgUUID wraps id as a valid pgtype.UUID for generated query parameters.
func toPgUUID(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

// checkNotFound inspects err from a single-row lookup and reports
// (isNotFound, wasNoRows, mappedErr):
//   - err == nil: (false, false, nil) — the caller proceeds to use the row.
//   - err is pgx.ErrNoRows: (true, true, notFoundErr) — notFoundErr may be
//     nil, matching repositories that report "not found" as (zero value,
//     nil) rather than a dedicated sentinel error.
//   - any other error: (true, false, err) unchanged, so the caller can
//     still apply its own logging/wrapping before returning.
//
// wasNoRows lets a caller that needs different handling for a genuine
// failure (e.g. logging) branch on it directly, without re-inspecting err
// with its own errors.Is(err, pgx.ErrNoRows) check.
func checkNotFound(err error, notFoundErr error) (bool, bool, error) {
	switch {
	case err == nil:
		return false, false, nil
	case errors.Is(err, pgx.ErrNoRows):
		return true, true, notFoundErr
	default:
		return true, false, err
	}
}

const tracerName = "github.com/Haya372/ai-trial/backend/infrastructure/repository"

type baseRepository struct {
	pool   *pgxpool.Pool
	tracer trace.Tracer
}

func (r *baseRepository) querier(ctx context.Context) *query.Queries {
	if tx, ok := db.GetTx(ctx); ok {
		return query.New(tx)
	}
	return query.New(r.pool)
}

// startDBSpan starts a "db.query" span per the observability guidelines'
// naming and attribute rules (db.system/db.operation/db.table; SQL text is
// never attached). The caller must end the span via endDBSpan.
//
//nolint:spancheck // ended by the caller via endDBSpan, not here
func (r *baseRepository) startDBSpan(ctx context.Context, operation, table string) (context.Context, trace.Span) {
	return r.tracer.Start(ctx, "db.query", trace.WithAttributes(
		attribute.String("db.system", "postgresql"),
		attribute.String("db.operation", operation),
		attribute.String("db.table", table),
	))
}

// endDBSpan records err on the span (if non-nil) and ends it. Call
// immediately after the traced query returns.
func endDBSpan(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}

// endDBSpanNotFound behaves like endDBSpan, except pgx.ErrNoRows is treated
// as a successful outcome rather than a span error: "no row matched" is an
// expected result for an ID lookup, not a query failure, and marking it as
// an error would pollute error-rate metrics/traces for routine lookups
// (e.g. an expired session, an unregistered email at login).
func endDBSpanNotFound(span trace.Span, err error) {
	if errors.Is(err, pgx.ErrNoRows) {
		span.End()
		return
	}
	endDBSpan(span, err)
}
