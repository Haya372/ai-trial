ALTER TABLE recurring_event_series
    ADD CONSTRAINT recurring_event_series_id_user_id_key UNIQUE (id, user_id);

ALTER TABLE events
    DROP CONSTRAINT events_series_id_fkey,
    ADD CONSTRAINT events_series_id_user_id_fkey
        FOREIGN KEY (series_id, user_id) REFERENCES recurring_event_series (id, user_id) ON DELETE CASCADE;

DROP INDEX idx_events_series_id;
CREATE INDEX idx_events_series_id ON events (series_id, start_at);
