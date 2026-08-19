-- +goose Up
CREATE TABLE browser_pairing (
    id INTEGER PRIMARY KEY,
    device_id INTEGER NOT NULL REFERENCES device (id) ON DELETE CASCADE,
    token_hash TEXT UNIQUE NOT NULL,
    created_at INTEGER NOT NULL,
    expires_at INTEGER NOT NULL,
    consumed_at INTEGER
);

CREATE TABLE browser_client (
    id INTEGER PRIMARY KEY,
    device_id INTEGER NOT NULL REFERENCES device (id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    browser TEXT NOT NULL DEFAULT '',
    token_hash TEXT UNIQUE NOT NULL,
    created_at INTEGER NOT NULL,
    last_used_at INTEGER NOT NULL,
    revoked_at INTEGER
);

CREATE INDEX idx_browser_pairing_token_hash ON browser_pairing(token_hash);
CREATE INDEX idx_browser_client_device_id ON browser_client(device_id);
CREATE INDEX idx_browser_client_token_hash ON browser_client(token_hash);

-- +goose Down
DROP TABLE IF EXISTS browser_client;
DROP TABLE IF EXISTS browser_pairing;
