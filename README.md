# Drone Tracker

Simulates drones moving around a map and displays them in realtime.

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

# Corners cut
- Write service layer
  - I just passed around the repo layer since this is basically CRUD and the service would've just been a relay layer with no domain logic
- IDs are just UUIDs for devices
  - Should've prefixed them like `DEVICE-f0f24ee3-44a3-4b2e-b2a1-07809f94fca1` for validation and readability
