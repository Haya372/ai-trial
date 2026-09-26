package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/Haya372/ai-trial/backend/infrastructure/db"
	query "github.com/Haya372/ai-trial/backend/infrastructure/db/generated"
)

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
