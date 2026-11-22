CREATE TABLE recurring_event_series (
    id             UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id        UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    pattern_type   VARCHAR     NOT NULL,
    monthly_nth    VARCHAR,
    end_type       VARCHAR     NOT NULL,
    end_until_date DATE,
    end_count      SMALLINT,
    timezone       TEXT        NOT NULL,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_recurring_event_series_user_id ON recurring_event_series (user_id);

CREATE TRIGGER update_recurring_event_series_updated_at
    BEFORE UPDATE ON recurring_event_series
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE recurring_series_exceptions (
    series_id     UUID        NOT NULL REFERENCES recurring_event_series(id) ON DELETE CASCADE,
    excluded_date DATE        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (series_id, excluded_date)
);

ALTER TABLE events
    ADD COLUMN series_id UUID REFERENCES recurring_event_series(id) ON DELETE CASCADE;

CREATE INDEX idx_events_series_id ON events (series_id);
