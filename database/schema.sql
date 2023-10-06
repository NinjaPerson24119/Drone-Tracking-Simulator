CREATE extension IF NOT EXISTS "uuid-ossp";

CREATE SCHEMA tracker;

CREATE TABLE IF NOT EXISTS tracker.device (
    tracker_id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    tracker_name TEXT NOT NULL,
    created TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS tracker.geolocation (
    tracker_id TEXT REFERENCES tracker.device NOT NULL,
    event_time TIMESTAMPTZ UNIQUE NOT NULL,
    longitude DECIMAL NOT NULL CHECK(longitude >= -180 AND longitude <= 180),
    latitude DECIMAL PRECISION NOT NULL CHECK(latitude >= -90 AND latitude <= 90),
    created TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted TIMESTAMPTZ
    PRIMARY KEY (tracker_id, event_time)
);
