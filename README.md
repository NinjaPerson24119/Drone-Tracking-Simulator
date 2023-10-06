# MapProject

- Assets within L2 distance get grouped (relative to zoom level)
  - Zooming in expands the group (if beyond L2 distance)
  - Clicking expands the group

# Stack
- Frontend
  - React + Next.js
- Backend
  - Golang / Gin
- Persistent Storage
  - PostgreSQL
- Hosting
  - Render.com host

# Design Considerations

- Realtime updates
  - Use Web sockets
- Large item quantities
  - Group on the backend
  - Frontend can request more details
