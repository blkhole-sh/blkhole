-- Create user table:
CREATE TABLE IF NOT EXISTS user (
    hash TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL
);

-- Create device table:
CREATE TABLE IF NOT EXISTS device (
    hash TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    os TEXT NOT NULL,
    user_hash TEXT NOT NULL REFERENCES user (hash) ON DELETE CASCADE
);

-- Create domain table:
CREATE TABLE IF NOT EXISTS domain (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

-- Create list table:
CREATE TABLE IF NOT EXISTS list (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    source TEXT,
    user_hash TEXT NOT NULL REFERENCES user (hash) ON DELETE CASCADE
);

-- Create domain - list reference table:
CREATE TABLE IF NOT EXISTS domain_list (
    domain_id INTEGER NOT NULL REFERENCES domain (id) ON DELETE CASCADE,
    list_id INTEGER NOT NULL REFERENCES list (id) ON DELETE CASCADE,
    PRIMARY KEY (domain_id, list_id)
);

-- Create schedule table:
CREATE TABLE IF NOT EXISTS schedule (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    start_time TEXT NOT NULL,
    end_time TEXT NOT NULL,
    user_hash TEXT NOT NULL REFERENCES user (hash) ON DELETE CASCADE,
    days INTEGER NOT NULL CHECK (days >= 0 AND days < 128),
    CHECK (start_time < end_time)
);

-- Create device - schedule reference table:
CREATE TABLE IF NOT EXISTS device_schedule (
    device_hash TEXT NOT NULL REFERENCES device (hash) ON DELETE CASCADE,
    schedule_id INTEGER NOT NULL REFERENCES schedule (id) ON DELETE CASCADE,
    PRIMARY KEY (device_hash, schedule_id)
);

-- Create domain - schedule reference table:
CREATE TABLE IF NOT EXISTS domain_schedule (
    domain_id INTEGER NOT NULL REFERENCES domain (id) ON DELETE CASCADE,
    schedule_id INTEGER NOT NULL REFERENCES schedule (id) ON DELETE CASCADE
);

-- Create list - schedule reference table:
CREATE TABLE IF NOT EXISTS list_schedule (
    list_id INTEGER NOT NULL REFERENCES list (id) ON DELETE CASCADE,
    schedule_id INTEGER NOT NULL REFERENCES schedule (id) ON DELETE CASCADE
);
