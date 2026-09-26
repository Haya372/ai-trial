CREATE TABLE event_shares (
    id         UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id   UUID        NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    token_hash TEXT        NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX event_shares_event_id_idx ON event_shares (event_id);
