
-- Create user table
CREATE TABLE IF NOT EXISTS User (
    hash VARCHAR(26) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL
);

-- Create device table
CREATE TABLE IF NOT EXISTS Device (
    hash VARCHAR(26) PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    user_hash VARCHAR(26) NOT NULL REFERENCES User(hash) ON DELETE CASCADE
);

-- Create domain table
CREATE TABLE IF NOT EXISTS Domain (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL
);

-- Create category table
CREATE TABLE IF NOT EXISTS Catgeory (
    id SMALLSERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL
);

-- Create domain - category reference table
CREATE TABLE IF NOT EXISTS Domain_Category (
    domain_id SERIAL NOT NULL REFERENCES Domain(id),
    category_id SMALLSERIAL NOT NULL REFERENCES Category(id),
    PRIMARY KEY (domain_id, category_id)
);

-- Create schedule table
CREATE TABLE IF NOT EXISTS Schedule (
    id SERIAL PRIMARY KEY,
    description VARCHAR(255),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    monday BOOLEAN NOT NULL,
    tuesday BOOLEAN NOT NULL,
    wednesday BOOLEAN NOT NULL,
    thursday BOOLEAN NOT NULL,
    friday BOOLEAN NOT NULL,
    saturday BOOLEAN NOT NULL,
    sunday BOOLEAN NOT NULL,
    CHECK (start_time < end_time)
);

-- Create device - schedule reference table
CREATE TABLE IF NOT EXISTS Device_Schedule (
  device_hash VARCHAR(26) NOT NULL REFERENCES Device(hash) ON DELETE CASCADE,
  schedule_id SERIAL NOT NULL REFERENCES Schedule(id) ON DELETE CASCADE,
  PRIMARY KEY(device_hash, schedule_id)
)

-- Create domain rule table
CREATE TABLE DomainRule (
  id SERIAL PRIMARY KEY,
  domain_id SERIAL NOT NULL REFERENCES Domain(id) ON DELETE CASCADE,
  schedule_id SERIAL NOT NULL REFERENCES Schedule(id) ON DELETE CASCADE,
  blocked BOOLEAN NOT NULL,
);

-- Create domain rule table
CREATE TABLE CategoryRule (
  id SERIAL PRIMARY KEY,
  category_id SERIAL NOT NULL REFERENCES Category(id) ON DELETE CASCADE,
  schedule_id SERIAL NOT NULL REFERENCES Schedule(id) ON DELETE CASCADE,
  blocked BOOLEAN NOT NULL,
);





