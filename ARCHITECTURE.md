# Architecture

## Request flow

```
React (frontend)
      │  fetch, JSON over HTTP
      ▼
routes/routes.go        — registers paths, applies middleware (CORS, logging)
      │
      ▼
handler/task_handler.go — decodes JSON, calls the service, encodes the response
      │
      ▼
service/task_service.go — validation + business rules (e.g. defaulting priority)
      │
      ▼
repository/task_repository.go — the only layer that writes SQL
      │
      ▼
MySQL
```

Each arrow is a one-way, strict boundary:

- **Handlers never touch SQL.** They only know about `http.Request` /
  `http.ResponseWriter` and the `TaskService` interface.
- **The repository never knows about HTTP.** It only knows about
  `models.Task` and `database/sql`.
- **The service layer is where business rules live** — required fields,
  enum validation, defaulting an omitted priority to `medium`, normalizing
  a blank color to the default accent, etc. Handlers and the repository
  stay "dumb" on purpose.

This is why swapping MySQL for Postgres would only touch
`internal/repository`, and why the service layer can be unit tested with
an in-memory fake `TaskRepository` without spinning up a database.

## Why an interface for the repository?

`TaskRepository` is defined as a Go interface, with `mysqlTaskRepository`
as its only implementation today. `NewTaskService` takes the interface,
not the concrete type. That's the whole trick: tests can pass in a fake
that satisfies the same interface, and `main.go` is the only place that
ever mentions the word "mysql".

## Why Go 1.22's `http.ServeMux` instead of a router library?

Go 1.22 added method-specific patterns (`"GET /api/tasks"`) and path
parameters (`"{id}"`) to the standard library's mux. For an API this size,
that's enough — it avoids a dependency whose only job is routing. If the
route table grows much larger (versioning, nested resource groups), a
router library would earn its place; for ~15 routes it isn't worth it.

## Data model

```
tasks
  id            BIGINT UNSIGNED  PK, auto-increment
  title         VARCHAR(150)     NOT NULL
  description   VARCHAR(2000)    NOT NULL DEFAULT ''
  due_date      DATE             NOT NULL
  status        ENUM(...)        NOT NULL DEFAULT 'pending'
  priority      ENUM(...)        NOT NULL DEFAULT 'medium'
  category      VARCHAR(60)      NOT NULL DEFAULT ''
  color         VARCHAR(9)       NOT NULL DEFAULT '#0e6b5c'
  favorite      BOOLEAN          NOT NULL DEFAULT FALSE
  archived      BOOLEAN          NOT NULL DEFAULT FALSE
  created_at    DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP
  updated_at    DATETIME         NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
```

Indexes exist on `status`, `due_date`, `priority`, `category`, and
`archived` — the columns the list endpoint filters and sorts on. `schema.sql`
is a one-shot convenience script for local setup; `migrations/up` /
`migrations/down` are the source of truth for schema changes going forward
(each migration is additive and backward compatible — see
`migrations/README.md`).

`archived` is a soft-delete flag, not a real delete: `ListTasks` excludes
archived rows by default (`WHERE archived = FALSE`), and the dashboard's
stats query does the same, so archiving a task removes it from view
without destroying data.

## Dashboard stats

`GET /api/tasks/stats` computes its counts with a single aggregate SQL
query (`SUM()` over boolean expressions), not by fetching every row into
Go and counting in a loop. That keeps the endpoint's cost roughly constant
as the table grows, instead of scaling with the number of tasks.

## Frontend structure

```
src/
  api/          fetch client (timeout + envelope unwrapping) + typed task API calls
  types/        shared TypeScript types, mirroring the Go models
  hooks/        useTasks (list + CRUD + bulk ops), useStats, useCategories,
                useKeyboardShortcuts — each owns one slice of server state
  components/   presentational + mildly-stateful UI pieces
  App.tsx       the only place that wires hooks together into a screen
```

The frontend has no routing library — `view` is a single piece of state in
`App.tsx` (`"dashboard" | "tasks"`), which is all two screens need. It also
has no external chart or icon library: the dashboard's status/priority
breakdown is built from styled `<div>`s sized by percentage, and sidebar
icons are inline SVG. That's a deliberate choice, not an oversight — it
keeps the dependency list at exactly `react` + `react-dom`, and both are
simple enough that a library would add more weight than it saves.

State updates are optimistic where the cost of being wrong is low
(favoriting a task flips instantly, then reconciles with the server
response — or rolls back on failure). Destructive actions (delete, bulk
delete) still go through a confirmation dialog before firing.

## What would need to change to scale this further

- **Repository tests**: swap in a fake `TaskRepository` and unit test
  `task_service.go`'s validation branches directly.
- **Pagination**: `ListTasks` currently returns the full filtered set.
  Fine at hundreds of rows; would need `LIMIT`/`OFFSET` or keyset pagination
  before tens of thousands.
- **Auth**: there's no user/session concept yet — every task is global.
  Adding a `user_id` column and an auth middleware would be the natural
  next step before this became multi-tenant.
