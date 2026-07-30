# API Reference (Quick Index)

Full narrative documentation with request/response bodies and curl examples
lives in [`API_DOCUMENTATION.md`](./API_DOCUMENTATION.md). An interactive
Swagger UI is also served directly by the running backend at
`http://localhost:8080/docs` (raw spec at `/openapi.yaml`). This file is
just a fast lookup table.

Base URL: `http://localhost:8080/api` · all responses use the envelope
`{ "success": bool, "data"?: ..., "error"?: string }`.

| Method | Path | Purpose |
|---|---|---|
| GET | `/health` | Liveness check (outside `/api`, used by Docker/uptime checks) |
| POST | `/api/tasks` | Create a task |
| GET | `/api/tasks` | List tasks (`search`, `status`, `priority`, `category`, `favorite`, `includeArchived`, `sortBy`) |
| GET | `/api/tasks/{id}` | Fetch one task |
| PUT | `/api/tasks/{id}` | Full update of one task |
| DELETE | `/api/tasks/{id}` | Delete one task |
| PATCH | `/api/tasks/{id}/favorite` | Toggle favorite (`{"favorite": bool}`) |
| PATCH | `/api/tasks/{id}/archive` | Toggle archive / soft-delete (`{"archived": bool}`) |
| POST | `/api/tasks/{id}/duplicate` | Clone a task |
| POST | `/api/tasks/bulk/delete` | Delete many (`{"ids": [...]}`) |
| POST | `/api/tasks/bulk/complete` | Mark many as completed (`{"ids": [...]}`) |
| GET | `/api/tasks/stats` | Dashboard aggregate counts |
| GET | `/api/categories` | Distinct categories in use, for filter/autocomplete |

All routes are registered in `backend/internal/routes/routes.go`, in the
same order as this table — that file is the ground truth if this ever
drifts.
