CREATE TABLE tags (
    id         UUID        PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       TEXT        NOT NULL,
    color      VARCHAR     NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_tags_user_id_lower_name ON tags (user_id, LOWER(name));

CREATE TRIGGER update_tags_updated_at
    BEFORE UPDATE ON tags
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TABLE event_tags (
    event_id   UUID        NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    tag_id     UUID        NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    position   SMALLINT    NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (event_id, tag_id),
    UNIQUE (event_id, position)
);

CREATE INDEX idx_event_tags_tag_id ON event_tags (tag_id);
