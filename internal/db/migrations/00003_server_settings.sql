-- +goose Up
CREATE TABLE server_setting (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- +goose Down
DROP TABLE server_setting;
