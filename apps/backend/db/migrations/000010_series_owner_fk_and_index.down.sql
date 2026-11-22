DROP INDEX IF EXISTS idx_events_series_id;
CREATE INDEX idx_events_series_id ON events (series_id);

ALTER TABLE events
    DROP CONSTRAINT IF EXISTS events_series_id_user_id_fkey,
    ADD CONSTRAINT events_series_id_fkey
        FOREIGN KEY (series_id) REFERENCES recurring_event_series (id) ON DELETE CASCADE;

ALTER TABLE recurring_event_series
    DROP CONSTRAINT IF EXISTS recurring_event_series_id_user_id_key;
