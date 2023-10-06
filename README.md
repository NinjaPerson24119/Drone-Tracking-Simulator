# Drone Tracker

Simulates drones moving around a map and displays them in realtime.

## Design
- Assets within L2 distance get grouped (relative to zoom level)
  - Zooming in expands the group (if beyond L2 distance)
  - Clicking expands the group
- Realtime updates
  - Use Web sockets
- Large item quantities
  - Group on the backend
  - Frontend can request more details

## Stack
- Frontend
  - React + Next.js
- Backend
  - Golang / Gin
  - Simulate ongoing tracker data
- Persistent Storage
  - PostgreSQL
- Hosting
  - Render.com host

# References
- https://docs.mapbox.com/help/tutorials/use-mapbox-gl-js-with-react/
