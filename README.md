# Drone Tracker

Simulates ingestion of drone geolocation data and displays it on a map in realtime.

[Check it out here!](https://map-project-r2zv.onrender.com)

TODO: screenshot

![Sequence Diagram](sequenceDiagram.png)

## Features
- Displays a map with drone locations in realtime.
- Drone locations are simulated by a backend service that moves them on a linear path defined by a starting point, radius to turn around at, and an angle.
- Zooming out the map will group drones together if they are within a certain distance of each other.
- Clicking on a drone will zoom in until the group is expanded

## Design
- Assets within L2 distance get grouped (relative to zoom level)
  - Zooming in expands the group (if beyond L2 distance)
  - Clicking expands the group
- Realtime updates
  - Use Web sockets

## Stack
- Frontend
  - React + Next.js
- Backend
  - Golang / Gin
  - Simulate ongoing tracker data using a circular bound. Devices just move back and forth along an angle through circle's center.
- Persistent Storage
  - PostgreSQL
- Hosting
  - Render.com host

# References
- https://docs.mapbox.com/help/tutorials/use-mapbox-gl-js-with-react/
- https://github.com/timescale/timescaledb
- https://lwebapp.com/en/post/go-websocket-simple-server
- https://docs.mapbox.com/mapbox-gl-js/example/cluster/
- https://github.com/gorilla/websocket/blob/main/examples/chat/client.go
- https://docs.mapbox.com/mapbox-gl-js/example/cluster/

# Corners cut
- Write service layer
  - I just passed around the repo layer since this is basically CRUD and the service would've just been a relay layer with no domain logic
- IDs are just UUIDs for devices
  - Should've prefixed them like `DEVICE-f0f24ee3-44a3-4b2e-b2a1-07809f94fca1` for validation and readability
- No multicast for notification queue of records inserted
  - As a result, max DB connections ~ max websocket connections

# Optimizations / Scaling considerations
- Simulator is decoupled from the websocket server to allow for testing INSERT load
  - Notifications are triggered by `pg_notify` when geolocations are inserted
- Websocket will initially send all locations, but after that it'll only send updates
  - Updates are batched by a buffer size of minimum send period, whichever occurs first

# Performance Improvements Possible
- Client <-> Server
  - Filter out updates where location is less than L2 distance from last known location
  - Binary encoding of websocket payload
- Server <-> DB
  - Use a timeseries DB like TimescaleDB
  - Assign sets of devices to a client's location, and use same server region for each related resource
    - When sharding the DB across multiple regions, this would likely be necessary since `pg_notify` is local to one DB and would need to be relayed, introducing additional latency
- Device <-> Server
  - The geolocation ingestion should pass through a service layer with a queue, or else we could accidentally DDoS ourselves with too many concurrent requests

# Illusions Possible
- Realtime updates over a network of unknown quality is really hard. We could smooth movement by allowing the frontend to guess where it thinks the device is going to be, and then correct it when the next update comes in.

# Deployment Considerations

The web server itself could probably run on a potato, but the PostgreSQL DB needs to be deployed on something with some decent specs.
If it can't keep up with realtime inserts, then the drone movement will be choppy, or the server might just overwhelmed with a backlog of too many inserts.
