DROP INDEX IF EXISTS idx_events_series_id;
ALTER TABLE events DROP COLUMN IF EXISTS series_id;
DROP TABLE IF EXISTS recurring_series_exceptions;
DROP TABLE IF EXISTS recurring_event_series;
