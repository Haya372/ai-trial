-- name: InsertRecurringSeries :exec
INSERT INTO recurring_event_series (
    id, user_id, pattern_type, monthly_nth, end_type, end_until_date, end_count, timezone
)
VALUES (
    sqlc.arg('id'),
    sqlc.arg('user_id'),
    sqlc.arg('pattern_type'),
    sqlc.arg('monthly_nth'),
    sqlc.arg('end_type'),
    sqlc.arg('end_until_date'),
    sqlc.arg('end_count'),
    sqlc.arg('timezone')
);

-- name: FindRecurringSeriesWithExceptionsByID :many
-- One row per excluded date (or one row with a NULL date when there are
-- none). Reading the series and its excluded dates in one statement gives
-- them the same snapshot, so a concurrent change cannot be seen half applied.
SELECT s.id, s.user_id, s.pattern_type, s.monthly_nth, s.end_type, s.end_until_date,
       s.end_count, s.timezone, x.excluded_date
FROM recurring_event_series s
LEFT JOIN recurring_series_exceptions x ON x.series_id = s.id
WHERE s.id = $1
ORDER BY x.excluded_date;

-- name: FindRecurringSeriesByIDForUpdate :one
SELECT id, user_id, pattern_type, monthly_nth, end_type, end_until_date, end_count, timezone
FROM recurring_event_series
WHERE id = $1
FOR UPDATE;

-- name: ListRecurringSeriesExceptions :many
SELECT excluded_date
FROM recurring_series_exceptions
WHERE series_id = $1
ORDER BY excluded_date;

-- name: UpdateRecurringSeries :exec
UPDATE recurring_event_series
SET pattern_type = sqlc.arg('pattern_type'),
    monthly_nth = sqlc.arg('monthly_nth'),
    end_type = sqlc.arg('end_type'),
    end_until_date = sqlc.arg('end_until_date'),
    end_count = sqlc.arg('end_count'),
    timezone = sqlc.arg('timezone')
WHERE id = sqlc.arg('id');

-- name: DeleteRecurringSeriesExceptions :exec
DELETE FROM recurring_series_exceptions
WHERE series_id = $1;

-- name: InsertRecurringSeriesExceptions :copyfrom
INSERT INTO recurring_series_exceptions (series_id, excluded_date)
VALUES ($1, $2);

-- name: DeleteRecurringSeries :exec
DELETE FROM recurring_event_series
WHERE id = $1;

-- name: ListOccurrencesWithTagsBySeriesID :many
-- One row per tag (or one row with NULL tag columns for an untagged event).
-- id breaks ties between occurrences that start at the same time, so each
-- event's rows stay together.
SELECT e.id, e.user_id, e.title, e.description, e.start_at, e.end_at, e.location, e.url,
       e.series_id, e.created_at, e.updated_at, et.tag_id, et.position
FROM events e
LEFT JOIN event_tags et ON et.event_id = e.id
WHERE e.series_id = $1
ORDER BY e.start_at, e.id, et.position;

-- name: CountOccurrencesBySeriesID :one
SELECT COUNT(*)
FROM events
WHERE series_id = $1;

-- name: InsertOccurrences :copyfrom
INSERT INTO events (id, user_id, title, description, start_at, end_at, location, url, series_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9);

-- name: InsertOccurrenceTags :copyfrom
INSERT INTO event_tags (event_id, tag_id, position)
VALUES ($1, $2, $3);

-- name: DeleteOccurrencesFromBySeriesID :exec
DELETE FROM events
WHERE series_id = sqlc.arg('series_id')
  AND start_at >= sqlc.arg('from_at');

-- name: DeleteOccurrencesBySeriesID :exec
DELETE FROM events
WHERE series_id = $1;
