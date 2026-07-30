# API Documentation

> **Interactive docs:** once the backend is running, open
> [`http://localhost:8080/docs`](http://localhost:8080/docs) for a live
> Swagger UI, or fetch the raw spec at `http://localhost:8080/openapi.yaml`.
> The machine-readable source is `backend/internal/docs/openapi.yaml`.
> This file is the human-readable companion with narrative context and
> curl examples.

Base URL: `http://localhost:8080/api`

All responses use this envelope:

```json
{ "success": true,  "data": ... }
{ "success": false, "error": "message describing what went wrong" }
```

Content-Type for all requests/responses: `application/json`.

---

## Task object

```json
{
  "id": 1,
  "title": "Design database schema",
  "description": "Model the tasks table with proper indexes",
  "dueDate": "2026-07-20",
  "status": "in_progress",
  "priority": "high",
  "category": "Backend",
  "color": "#0e6b5c",
  "favorite": true,
  "archived": false,
  "createdAt": "2026-07-17T09:00:00Z",
  "updatedAt": "2026-07-17T09:00:00Z"
}
```

`status` is one of: `pending`, `in_progress`, `completed`.
`priority` is one of: `low`, `medium`, `high` (defaults to `medium` if omitted on create).
`color` is a hex string; defaults to `#0e6b5c` if omitted or blank.
`category` is a free-text label (up to 60 characters); empty string means uncategorized.

---

## GET /health

Health check for uptime monitoring / container orchestration.

**Response `200`**
```json
{ "success": true, "data": { "status": "ok" } }
```

---

## POST /api/tasks

Create a new task.

**Request body**
```json
{
  "title": "Write unit tests",
  "description": "Cover the service layer",
  "dueDate": "2026-07-25",
  "status": "pending",
  "priority": "high",
  "category": "Backend",
  "color": "#0e6b5c"
}
```

| Field | Rules |
|---|---|
| `title` | required, 1–150 characters |
| `description` | optional |
| `dueDate` | required, format `YYYY-MM-DD` |
| `status` | required, one of `pending` / `in_progress` / `completed` |
| `priority` | optional, one of `low` / `medium` / `high` (defaults to `medium`) |
| `category` | optional, up to 60 characters |
| `color` | optional hex string (defaults to `#0e6b5c`) |

**Response `201`** — the created Task object.

**Response `400`** — validation error, e.g.:
```json
{ "success": false, "error": "title is required" }
```

---

## GET /api/tasks

List tasks. Supports optional query parameters:

| Param | Values | Description |
|---|---|---|
| `search` | any string | matches against title or description |
| `status` | `pending` \| `in_progress` \| `completed` | filter by exact status |
| `priority` | `low` \| `medium` \| `high` | filter by exact priority |
| `category` | any string | filter by exact category |
| `favorite` | `true` | when present, only favorited tasks |
| `includeArchived` | `true` | when present, includes archived tasks (excluded by default) |
| `sortBy` | `due_date_asc` \| `due_date_desc` \| `created_at_desc` \| `priority` | sort order (default: `created_at_desc`) |

Example:
```
GET /api/tasks?status=pending&sortBy=due_date_asc&search=schema
```

**Response `200`** — array of Task objects (empty array if none match).

---

## GET /api/tasks/{id}

Fetch a single task by ID.

**Response `200`** — the Task object.
**Response `404`**
```json
{ "success": false, "error": "task not found" }
```

---

## PUT /api/tasks/{id}

Full update of a task. Same body/validation rules as `POST /api/tasks`.

**Response `200`** — the updated Task object.
**Response `400`** — validation error.
**Response `404`** — task not found.

---

## DELETE /api/tasks/{id}

Deletes a task.

**Response `200`**
```json
{ "success": true, "data": { "id": 7 } }
```

**Response `404`** — task not found.

---

## PATCH /api/tasks/{id}/favorite

Toggle a task's favorite flag.

**Request body**
```json
{ "favorite": true }
```

**Response `200`** — the updated Task object.
**Response `404`** — task not found.

---

## PATCH /api/tasks/{id}/archive

Archive or unarchive a task. Archived tasks are excluded from `GET /api/tasks`
by default (pass `includeArchived=true` to see them) and from the dashboard's
stats counts. This is a soft delete — data isn't destroyed.

**Request body**
```json
{ "archived": true }
```

**Response `200`** — the updated Task object.
**Response `404`** — task not found.

---

## POST /api/tasks/{id}/duplicate

Creates an independent copy of a task: same fields, title suffixed with
" (copy)", status reset to `pending`, and `favorite`/`archived` reset to
`false`.

**Response `201`** — the newly created Task object.
**Response `404`** — source task not found.

---

## POST /api/tasks/bulk/delete

Delete multiple tasks in one request.

**Request body**
```json
{ "ids": [1, 2, 3] }
```

**Response `200`**
```json
{ "success": true, "data": { "deleted": 3 } }
```

`deleted` reflects however many of the given IDs actually existed; IDs that
don't exist are silently skipped rather than causing a 404.

---

## POST /api/tasks/bulk/complete

Mark multiple tasks as `completed` in one request. Same body/response shape
as bulk delete, with a `completed` count instead of `deleted`.

---

## GET /api/tasks/stats

Aggregate counts for the dashboard, computed with a single SQL query
(archived tasks excluded, except from the `archived` count itself).

**Response `200`**
```json
{
  "success": true,
  "data": {
    "total": 12,
    "completed": 5,
    "pending": 4,
    "inProgress": 3,
    "highPriority": 2,
    "overdue": 1,
    "upcomingWeek": 3,
    "favorites": 2,
    "archived": 1
  }
}
```

---

## GET /api/categories

Distinct, non-empty category values currently in use (excluding archived
tasks), sorted alphabetically. Used to populate the category filter
dropdown and the task form's autocomplete without hardcoding a list.

**Response `200`**
```json
{ "success": true, "data": ["Backend", "Docs", "Frontend", "Setup"] }
```

---

## Error status codes summary

| Status | Meaning |
|---|---|
| `400` | Invalid input (missing/invalid field, bad ID) |
| `404` | Task not found |
| `500` | Unexpected server error (logged server-side; message is not leaked to the client) |

---

## Example curl session

```bash
# Create
curl -X POST http://localhost:8080/api/tasks \
  -H "Content-Type: application/json" \
  -d '{"title":"Buy milk","description":"","dueDate":"2026-07-20","status":"pending"}'

# List, filtered and sorted
curl "http://localhost:8080/api/tasks?status=pending&sortBy=due_date_asc"

# Get one
curl http://localhost:8080/api/tasks/1

# Update
curl -X PUT http://localhost:8080/api/tasks/1 \
  -H "Content-Type: application/json" \
  -d '{"title":"Buy milk","description":"2%","dueDate":"2026-07-21","status":"completed"}'

# Delete
curl -X DELETE http://localhost:8080/api/tasks/1

# Favorite
curl -X PATCH http://localhost:8080/api/tasks/1/favorite \
  -H "Content-Type: application/json" \
  -d '{"favorite":true}'

# Archive
curl -X PATCH http://localhost:8080/api/tasks/1/archive \
  -H "Content-Type: application/json" \
  -d '{"archived":true}'

# Duplicate
curl -X POST http://localhost:8080/api/tasks/1/duplicate

# Bulk delete
curl -X POST http://localhost:8080/api/tasks/bulk/delete \
  -H "Content-Type: application/json" \
  -d '{"ids":[1,2,3]}'

# Bulk complete
curl -X POST http://localhost:8080/api/tasks/bulk/complete \
  -H "Content-Type: application/json" \
  -d '{"ids":[1,2,3]}'

# Dashboard stats
curl http://localhost:8080/api/tasks/stats

# Categories
curl http://localhost:8080/api/categories
```
