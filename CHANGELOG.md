# Changelog

## [Unreleased] — Premium overhaul

Builds on top of the original assignment submission below without changing
its API contracts or database engine. Existing endpoints keep working
exactly as before; everything here is additive.

### Added — Backend
- Dashboard aggregate endpoint: `GET /api/tasks/stats`
- Category listing endpoint: `GET /api/categories`
- Favorite toggle: `PATCH /api/tasks/{id}/favorite`
- Archive (soft-delete) toggle: `PATCH /api/tasks/{id}/archive`
- Duplicate: `POST /api/tasks/{id}/duplicate`
- Bulk operations: `POST /api/tasks/bulk/delete`, `POST /api/tasks/bulk/complete`
- `priority`, `category`, `color`, `favorite`, `archived` fields on the Task
  model, request DTOs, and JSON responses
- `List` filtering extended to `priority`, `category`, `favorite`, and
  `includeArchived`; sorting extended with a `priority` option
- `Stats()` and `Categories()` added to the repository interface, backed by
  aggregate SQL rather than fetching every row

### Added — Frontend
- Responsive sidebar (Dashboard / Tasks navigation, collapses to an
  off-canvas drawer on mobile) and a view-aware top bar
- Dashboard page: 8 stat cards (total, completed, pending, in progress,
  high priority, overdue, due this week, favorites), a status breakdown
  bar, and an upcoming-deadlines list — all built with plain CSS/SVG, no
  charting library
- Priority badges, category tags, a favorite star, and a color accent
  stripe on task cards
- Bulk-select mode with a bulk action bar (mark complete / delete selected)
- Duplicate and archive quick actions per task
- Category and priority filters, a "favorites only" toggle, and a
  priority sort option, alongside the existing search/status/sort
- Color picker and category/priority fields in the task form
- Keyboard shortcuts: `/` focuses search, `n` opens the add-task form
- Optimistic favorite toggling (flips instantly, rolls back on failure)

### Changed
- `Header` became a view-aware top bar (`Dashboard` / `Tasks` titles) with
  a mobile menu button
- `TaskItem` / `TaskList` / `SearchFilterBar` / `TaskForm` extended to
  carry the new fields and actions through to the API layer
- CSS extended (sidebar, dashboard, badges, bulk bar, color picker) using
  the existing palette and type scale — no new design language introduced

### Not changed
- Database engine, table name, or existing column definitions
- Existing endpoint request/response shapes (only new optional fields
  added; nothing removed or renamed)
- Layered architecture (routes → handler → service → repository → MySQL)

---

## Original submission

- Full CRUD REST API in Go, layered architecture, MySQL persistence
- React + TypeScript frontend: dashboard-style task list, add/edit modal,
  delete confirmation, search/filter/sort, skeleton loaders, toasts,
  error boundary
- Input validation on both client and server
- CORS + request logging middleware, graceful shutdown, env-based config
- Numbered SQL migrations, Docker Compose setup, embedded OpenAPI/Swagger
  docs at `/docs`
