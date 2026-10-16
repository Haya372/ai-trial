-- Existing rows hold the plaintext session ID in id (the cookie value), which
-- cannot be carried over to hashed storage; delete them to force re-login (ADR-033).
DELETE FROM sessions;

ALTER TABLE sessions ADD COLUMN token_hash TEXT NOT NULL UNIQUE;
