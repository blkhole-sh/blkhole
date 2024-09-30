-- Enable foreign keys in SQLite
PRAGMA foreign_keys = ON;

-- Create user table
CREATE TABLE IF NOT EXISTS User (
    hash TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL
);

-- Create device table
CREATE TABLE IF NOT EXISTS Device (
    hash TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    user_hash TEXT NOT NULL REFERENCES User(hash) ON DELETE CASCADE
);

-- Create domain table
CREATE TABLE IF NOT EXISTS Domain (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL
);

-- Create category table
CREATE TABLE IF NOT EXISTS Category (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT UNIQUE NOT NULL
);

-- Create domain - category reference table
CREATE TABLE IF NOT EXISTS Domain_Category (
    domain_id INTEGER NOT NULL REFERENCES Domain(id) ON DELETE CASCADE,
    category_id INTEGER NOT NULL REFERENCES Category(id) ON DELETE CASCADE,
    PRIMARY KEY (domain_id, category_id)
);

-- Create schedule table
CREATE TABLE IF NOT EXISTS Schedule (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    description TEXT,
    start_time TEXT NOT NULL,
    end_time TEXT NOT NULL,
    monday INTEGER NOT NULL,      -- 0 (false) or 1 (true)
    tuesday INTEGER NOT NULL,     -- 0 (false) or 1 (true)
    wednesday INTEGER NOT NULL,   -- 0 (false) or 1 (true)
    thursday INTEGER NOT NULL,    -- 0 (false) or 1 (true)
    friday INTEGER NOT NULL,      -- 0 (false) or 1 (true)
    saturday INTEGER NOT NULL,    -- 0 (false) or 1 (true)
    sunday INTEGER NOT NULL,      -- 0 (false) or 1 (true)
    CHECK (start_time < end_time)
);

-- Create device - schedule reference table
CREATE TABLE IF NOT EXISTS Device_Schedule (
    device_hash TEXT NOT NULL REFERENCES Device(hash) ON DELETE CASCADE,
    schedule_id INTEGER NOT NULL REFERENCES Schedule(id) ON DELETE CASCADE,
    PRIMARY KEY(device_hash, schedule_id)
);

-- Create domain rule table
CREATE TABLE DomainRule (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    domain_id INTEGER NOT NULL REFERENCES Domain(id) ON DELETE CASCADE,
    schedule_id INTEGER NOT NULL REFERENCES Schedule(id) ON DELETE CASCADE,
    blocked INTEGER NOT NULL  -- 0 (false) or 1 (true)
);

-- Create category rule table
CREATE TABLE CategoryRule (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id INTEGER NOT NULL REFERENCES Category(id) ON DELETE CASCADE,
    schedule_id INTEGER NOT NULL REFERENCES Schedule(id) ON DELETE CASCADE,
    blocked INTEGER NOT NULL  -- 0 (false) or 1 (true)
);

