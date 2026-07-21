# Interview Guide

The assignment says you'll be asked to explain request flow, routing, state
management, database schema, API design, and any AI-generated code. This
doc collects likely questions with straight answers grounded in this
codebase, so you can walk through them without hunting for the right file
mid-interview.

## "Walk me through what happens when I click 'Add task'."

1. `TaskForm` (a modal) collects title/description/due date/status/priority/
   category/color, validates client-side (`validate()` in `TaskForm.tsx`,
   mirroring the backend's rules so the user gets instant feedback).
2. On submit, `App.tsx`'s `handleFormSubmit` calls `createTask` from the
   `useTasks` hook, which calls `taskApi.create()`.
3. `taskApi.create` POSTs JSON to `/api/tasks` via the shared `apiClient`
   (`api/client.ts`), which adds a 10s timeout and unwraps the
   `{success, data, error}` envelope.
4. `routes.go` matches `POST /api/tasks` to `TaskHandler.CreateTask`, which
   decodes the JSON body and calls `TaskService.CreateTask`.
5. The service validates (required title, valid date format, valid status/
   priority enum, length limits) and normalizes optional fields (priority
   defaults to `medium`, color defaults to the accent green), then calls
   `TaskRepository.Create`.
6. The repository runs a parameterized `INSERT`, then re-fetches the row by
   its new ID (`GetByID`) so it returns the full row including DB defaults
   like `created_at`.
7. The handler wraps the returned `models.Task` in a `taskResponse` DTO
   (which formats the due date as a plain string and the timestamps as
   RFC3339) and writes it back as JSON.
8. The frontend prepends the new task to local state and shows a success
   toast — no full-page refetch needed.

## "Why is there a separate `taskResponse` type instead of just marshaling `models.Task` directly?"

`models.Task.DueDate` is a `time.Time`. Marshaled directly, that would come
out as a full RFC3339 timestamp, but the frontend's `<input type="date">`
wants a plain `YYYY-MM-DD` string. `taskResponse` exists so the JSON shape
the frontend receives can differ from the Go-native domain type without
adding date-parsing logic to the frontend.

## "Why an interface for the repository instead of just calling MySQL from the service?"

`TaskService` depends on the `TaskRepository` interface, not the concrete
`mysqlTaskRepository`. That means the service layer's business rules
(required fields, enum validation) can be unit tested against an in-memory
fake repository, with no database involved. It also means swapping MySQL
for another store would only touch `internal/repository`.

## "How does search/filter/sort work end-to-end?"

Query params (`search`, `status`, `priority`, `category`, `favorite`,
`includeArchived`, `sortBy`) are parsed in `TaskHandler.ListTasks`, passed
into a `service.TaskListQuery`, validated there (e.g. an invalid status
value is rejected as a 400, not silently ignored), then turned into a
`repository.TaskFilter`. The repository builds the SQL query by
conditionally appending `AND` clauses and a `ORDER BY` — every value is
still passed as a bound parameter (`?`), never string-concatenated, so this
isn't vulnerable to SQL injection despite building the query dynamically.
On the frontend, `useTasks` re-fetches whenever any filter/sort param
changes (its `useCallback` dependency array), and the search box is
debounced 350ms so typing doesn't fire a request per keystroke.

## "Why no React Router / Redux / charting library?"

The app has exactly two screens (Dashboard, Tasks), so `view` is one piece
of state in `App.tsx` rather than a router. State is server data + a
handful of local UI flags, owned by three small hooks (`useTasks`,
`useStats`, `useCategories`) — there's no cross-cutting client state that
would justify Redux/Zustand. The dashboard's charts are styled `<div>`s
sized by percentage rather than a charting library, because the visuals
needed (a segmented bar, a sorted list) don't need one. This was a
deliberate scope call, not an oversight — the tradeoff is that this
approach wouldn't scale gracefully to a much more complex dashboard.

## "What's the archived flag for, and why not just delete?"

`archived` is a soft-delete: `ListTasks` excludes archived rows by default,
and the stats query does the same. Archiving lets a task disappear from
the active view without destroying its history — useful for tasks you want
out of the way but might reference later. `includeArchived=true` on the
list endpoint surfaces them again.

## "How would you add pagination if the task list grew to 100k rows?"

`ListTasks` currently returns every matching row. I'd add `limit`/`offset`
(or keyset pagination on `id`/`created_at` for large offsets, since
`OFFSET` gets slow at scale) as new query params, threaded through
`TaskFilter` the same way the existing filters are — the repository's
query-building pattern already supports appending clauses conditionally.

## "What would you point to as AI-generated vs. things you changed or would want to double check?"

Be honest and specific here — this is literally 20% of the grade. Concretely
for this codebase:
- The layered architecture, the repository interface pattern, and the
  overall file structure are the kind of thing AI tools produce well and
  consistently.
- The SQL in `Stats()` (aggregate `SUM()` over boolean expressions) is a
  common pattern but worth being able to explain in your own words — see
  `ARCHITECTURE.md`'s "Dashboard stats" section.
- Anything you haven't personally run and clicked through, say so plainly
  rather than claiming full ownership. This submission's build wasn't
  verified against a live Go/Node toolchain in the environment it was
  generated in — see the note in `README.md` — so treat `go build ./...`
  and `npm run build` as your first two commands before you consider it
  submission-ready, and be ready to say that's what you did.

## "Show me the database schema and explain an index choice."

See `ARCHITECTURE.md`'s "Data model" section for the full table. One
concrete example: `idx_tasks_due_date` exists because the default and most
common sort is by due date, and the dashboard's overdue/upcoming-week
counts both filter on `due_date` — without the index those would be full
table scans as the table grows.
