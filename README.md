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

# Improvements TODO
- https://pkg.go.dev/github.com/jackc/pgx/v5#RowToStructByName
  - Now that Go has generics, we can use this to make a generic function to convert a row to a struct
- Write service layer. Didn't bother because it basically just duplicates the repo layer (CRUD)
  - Simulator just does a loop-back API call, so it's not necessary there either
- Prefix IDs like `DEVICE-f0f24ee3-44a3-4b2e-b2a1-07809f94fca1` for validation and readability
