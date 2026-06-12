-- +goose Up
-- Create user table:
CREATE TABLE IF NOT EXISTS user (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL
);

-- Create device table:
CREATE TABLE IF NOT EXISTS device (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    os TEXT NOT NULL,
    hash TEXT NOT NULL,
    user_id INTEGER NOT NULL REFERENCES user (id) ON DELETE CASCADE
);

-- Create domain table:
CREATE TABLE IF NOT EXISTS domain (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

-- Create rule table:
CREATE TABLE IF NOT EXISTS rule (
    id INTEGER PRIMARY KEY,
    allowed INTEGER NOT NULL,
    domain_id INTEGER NOT NULL REFERENCES domain (id) ON DELETE CASCADE,
    UNIQUE(domain_id, allowed)
);

-- Create list table:
CREATE TABLE IF NOT EXISTS list (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    source TEXT,
    user_id INTEGER NOT NULL REFERENCES user (id) ON DELETE CASCADE,
    count INTEGER NOT NULL DEFAULT 0,
    is_default INTEGER NOT NULL DEFAULT 0,
    UNIQUE(name, user_id)
);

-- Create schedule table:
CREATE TABLE IF NOT EXISTS schedule (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    start_time TEXT NOT NULL,
    end_time TEXT NOT NULL,
    days INTEGER NOT NULL CHECK (days >= 0 AND days < 128),
    active INTEGER NOT NULL DEFAULT 1,
    is_default INTEGER NOT NULL DEFAULT 0,
    user_id INTEGER NOT NULL REFERENCES user (id) ON DELETE CASCADE,
    -- Validate that times follow HH:MM format
    CHECK (start_time GLOB '[0-2][0-9]:[0-5][0-9]'),
    CHECK (end_time GLOB '[0-2][0-9]:[0-5][0-9]'),
    -- Validate that times fit 5-minute slots (minutes must be multiple of 5)
    CHECK (CAST(SUBSTR(start_time, 4, 2) AS INTEGER) % 5 = 0),
    CHECK (CAST(SUBSTR(end_time, 4, 2) AS INTEGER) % 5 = 0),
    -- Validate that hours are 00-23
    CHECK (CAST(SUBSTR(start_time, 1, 2) AS INTEGER) >= 0 AND CAST(SUBSTR(start_time, 1, 2) AS INTEGER) <= 23),
    CHECK (CAST(SUBSTR(end_time, 1, 2) AS INTEGER) >= 0 AND CAST(SUBSTR(end_time, 1, 2) AS INTEGER) <= 23)
);

-- Create device - schedule reference table:
CREATE TABLE IF NOT EXISTS device_schedule (
    device_id INTEGER NOT NULL REFERENCES device (id) ON DELETE CASCADE,
    schedule_id INTEGER NOT NULL REFERENCES schedule (id) ON DELETE CASCADE,
    PRIMARY KEY (device_id, schedule_id)
);

-- Create list - schedule reference table:
CREATE TABLE IF NOT EXISTS list_schedule (
    list_id INTEGER NOT NULL REFERENCES list (id) ON DELETE CASCADE,
    schedule_id INTEGER NOT NULL REFERENCES schedule (id) ON DELETE CASCADE,
    PRIMARY KEY (list_id, schedule_id)
);

-- Create list - rule reference table:
CREATE TABLE IF NOT EXISTS list_rule (
    list_id INTEGER NOT NULL REFERENCES list (id) ON DELETE CASCADE,
    rule_id INTEGER NOT NULL REFERENCES rule (id),
    PRIMARY KEY (list_id, rule_id)
);

-- Create schedule - rule reference table:
CREATE TABLE IF NOT EXISTS schedule_rule (
    schedule_id INTEGER NOT NULL REFERENCES schedule (id) ON DELETE CASCADE,
    rule_id INTEGER NOT NULL REFERENCES rule (id),
    PRIMARY KEY (schedule_id, rule_id)
);

-- Create query log table:
CREATE TABLE IF NOT EXISTS query_log (
    id          INTEGER PRIMARY KEY,
    device_hash TEXT NOT NULL,
    domain      TEXT NOT NULL,
    blocked     INTEGER NOT NULL DEFAULT 0,
    timestamp   INTEGER NOT NULL
);

-- Create indexes for performance optimization:
CREATE INDEX IF NOT EXISTS idx_user_email ON user(email);
CREATE INDEX IF NOT EXISTS idx_device_user_id ON device(user_id);
CREATE INDEX IF NOT EXISTS idx_rule_domain_id ON rule(domain_id);
CREATE INDEX IF NOT EXISTS idx_domain_name ON domain(name);
CREATE INDEX IF NOT EXISTS idx_schedule_user_id ON schedule(user_id);
CREATE INDEX IF NOT EXISTS idx_device_schedule_device_id ON device_schedule(device_id);
CREATE INDEX IF NOT EXISTS idx_list_schedule_list_id ON list_schedule(list_id);
CREATE INDEX IF NOT EXISTS idx_list_schedule_schedule_id ON list_schedule(schedule_id);
CREATE INDEX IF NOT EXISTS idx_list_rule_rule_id ON list_rule(rule_id);
CREATE INDEX IF NOT EXISTS idx_schedule_rule_rule_id ON schedule_rule(rule_id);
CREATE INDEX IF NOT EXISTS idx_query_log_device ON query_log(device_hash, timestamp);

-- +goose Down
DROP TABLE IF EXISTS query_log;
DROP TABLE IF EXISTS schedule_rule;
DROP TABLE IF EXISTS list_rule;
DROP TABLE IF EXISTS list_schedule;
DROP TABLE IF EXISTS device_schedule;
DROP TABLE IF EXISTS schedule;
DROP TABLE IF EXISTS list;
DROP TABLE IF EXISTS rule;
DROP TABLE IF EXISTS domain;
DROP TABLE IF EXISTS device;
DROP TABLE IF EXISTS user;
