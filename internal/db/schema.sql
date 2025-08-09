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

-- Create list table:
CREATE TABLE IF NOT EXISTS list (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    description TEXT,
    source TEXT,
    user_hash TEXT NOT NULL REFERENCES user (hash) ON DELETE CASCADE
);

-- Create rule table:
CREATE TABLE IF NOT EXISTS rule (
    id INTEGER PRIMARY KEY,
    domain TEXT NOT NULL,
    allowed INTEGER NOT NULL
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

-- Create list - schedule reference table:
CREATE TABLE IF NOT EXISTS list_schedule (
    list_id INTEGER NOT NULL REFERENCES list (id) ON DELETE CASCADE,
    schedule_id INTEGER NOT NULL REFERENCES schedule (id) ON DELETE CASCADE,
    PRIMARY KEY (list_id, schedule_id)
);

-- Create list - rule reference table:
CREATE TABLE IF NOT EXISTS list_rule (
    list_id INTEGER NOT NULL REFERENCES list (id) ON DELETE CASCADE,
    rule_id INTEGER NOT NULL REFERENCES rule (id) ON DELETE CASCADE,
    PRIMARY KEY (list_id, rule_id)
);

-- Create schedule - domain reference table:
CREATE TABLE IF NOT EXISTS schedule_rule (
    schedule_id INTEGER NOT NULL REFERENCES schedule (id) ON DELETE CASCADE,
    rule_id INTEGER NOT NULL REFERENCES rule (id) ON DELETE CASCADE,
    PRIMARY KEY (schedule_id, rule_id)
);

-- Create indexes for performance optimization:
CREATE INDEX IF NOT EXISTS idx_user_email ON user(email);
CREATE INDEX IF NOT EXISTS idx_device_user_hash ON device(user_hash);
CREATE INDEX IF NOT EXISTS idx_rule_domain ON rule(domain);
CREATE INDEX IF NOT EXISTS idx_schedule_user_hash ON schedule(user_hash);
CREATE INDEX IF NOT EXISTS idx_device_schedule_device_hash ON device_schedule(device_hash);
CREATE INDEX IF NOT EXISTS idx_list_schedule_list_id ON list_schedule(list_id);
CREATE INDEX IF NOT EXISTS idx_list_schedule_schedule_id ON list_schedule(schedule_id);
CREATE INDEX IF NOT EXISTS idx_list_rule_rule_id ON list_rule(rule_id);
CREATE INDEX IF NOT EXISTS idx_schedule_rule_rule_id ON schedule_rule(rule_id);
