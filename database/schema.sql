CREATE extension IF NOT EXISTS "uuid-ossp";

CREATE SCHEMA device;

CREATE TABLE IF NOT EXISTS device.information (
    device_id uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
    device_name TEXT NOT NULL,
    created TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS device.geolocation (
    device_id uuid REFERENCES device.information NOT NULL,
    event_time TIMESTAMPTZ UNIQUE NOT NULL,
    longitude DECIMAL NOT NULL CHECK(longitude >= -180 AND longitude <= 180),
    latitude DECIMAL NOT NULL CHECK(latitude >= -90 AND latitude <= 90),
    created TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP NOT NULL,
    deleted TIMESTAMPTZ,
    PRIMARY KEY (device_id, event_time)
);
