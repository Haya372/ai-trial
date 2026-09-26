CREATE TABLE event_subscriptions (
    id         UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    event_id   UUID        NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (event_id, user_id)
);

CREATE INDEX idx_event_subscriptions_user_id ON event_subscriptions (user_id);
