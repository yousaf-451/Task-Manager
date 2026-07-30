# AI Usage Notes

This project was built with AI assistance (Claude). Below are the key prompts
used and a summary of what the AI produced vs. what a developer would
typically review, adjust, or verify by hand before submitting.

> Fill in / adjust this file with your own real working notes — this is a
> template based on how the project was actually generated, written so you
> can submit it as-is or edit it to match your own process.

---

## Key prompts used

1. **Initial scaffolding prompt**
   > "Build a Task Management app: React + TypeScript frontend, Go REST API
   > with layered architecture (routes/handlers/services/repository/models),
   > MySQL persistence. Task fields: title, description, due date, status.
   > Full CRUD. Include validation, error handling, CORS, env config,
   > responsive UI, search/filter/sort, README, and API docs."

2. **Backend architecture follow-up**
   > "Keep handlers thin — no SQL in handlers, no HTTP types in the service
   > or repository layer. Use the repository pattern with an interface so
   > the service layer is unit-testable. Return a consistent JSON envelope
   > with success/data/error."

3. **Frontend follow-up**
   > "Add a debounced search box, status filter dropdown, and due-date sort.
   > Add a modal form for add/edit with client-side validation mirroring the
   > backend rules, a delete confirmation dialog, loading and empty states,
   > and toast notifications for success/error."

4. **Design pass**
   > "Give the UI a distinctive, considered visual identity — not a default
   > template look. Pick a real palette and type system and apply it
   > consistently."

---

## What AI generated

- The full layered backend (config, models, repository, service, handler,
  routes, middleware) and `cmd/api/main.go` wiring.
- The SQL schema and seed data.
- The React/TypeScript frontend: components, hooks, API client, types, and
  styling.
- README, API documentation, and this file.

## What a developer should verify / commonly adjusts by hand

- **Run `go mod tidy`** to fetch `github.com/go-sql-driver/mysql` and
  regenerate `go.sum` — this couldn't be done in the generation environment
  since it had no network access to the Go module proxy.
- **Run `npm install`** for the same reason (no `node_modules` / lockfile is
  committed).
- **Database credentials** in `.env` — the defaults in `.env.example` are
  placeholders and should be changed for anything beyond local development.
- **CORS origins** — confirm `CORS_ALLOWED_ORIGINS` matches wherever the
  frontend is actually deployed.
- **Error messages returned to the client** — reviewed to make sure no
  internal detail (SQL errors, stack traces) leaks; the service/handler
  split enforces this by mapping errors to generic messages for anything
  that isn't a known validation/not-found case.
- **Manual testing** of each CRUD flow end-to-end (create → appears in list
  → edit → filter/search/sort still finds it → delete → confirmation →
  removed from list) against a real MySQL instance, since the generation
  environment could not run `go run` or `npm run dev` against a live
  database.

---

## Addendum: professional hardening pass

A follow-up review pass was run against the completed project with the
prompt below, to bring it to interview-ready, production-adjacent quality
without changing the existing architecture or breaking any assignment
requirement.

> "Audit this complete Task Manager project end-to-end. Don't rebuild it —
> improve it. Add: Dockerfiles + docker-compose, a Makefile, an OpenAPI/
> Swagger spec served by the API, numbered SQL migrations, a client-side
> request timeout, skeleton loaders instead of a plain loading message, an
> app-level error boundary, and centralized constants instead of scattered
> magic numbers. Fix any real code smells (e.g. a lint script with no
> matching devDependencies installed). Keep every existing feature working."

What that pass added, concretely:

- `backend/Dockerfile`, `frontend/Dockerfile`, `frontend/nginx.conf`,
  `docker-compose.yml`, `Makefile`.
- `backend/internal/docs/openapi.yaml` + a small `docs` package that embeds
  it and serves both the raw spec and a Swagger UI at `/docs`.
- `backend/migrations/up|down/*.sql`, splitting schema from seed data and
  documenting rollback, while keeping `schema.sql` as the one-shot local
  setup script.
- `frontend/src/constants.ts` (debounce delay, toast duration, request
  timeout, title/description length limits) — previously these were magic
  numbers duplicated across `App.tsx` and `TaskForm.tsx`.
- `AbortController`-based timeout in `api/client.ts`, so a hung request
  fails with a clear message instead of leaving the UI stuck indefinitely.
- `SkeletonList.tsx`, replacing the plain "Loading tasks…" text with
  pulsing placeholder cards (the text is preserved for screen readers via
  an `sr-only` live region).
- `ErrorBoundary.tsx`, wrapping the app in `main.tsx` so an unexpected
  render-time bug shows a friendly full-page fallback instead of a blank
  screen — distinct from the existing inline error banner, which already
  handled expected API failures.
- Added the missing `eslint` + `@typescript-eslint/*` devDependencies and
  an `.eslintrc.cjs` — the `lint` script existed in `package.json` before
  this pass but had no matching packages installed, so it would have
  failed on a clean `npm install`.

Everything above was reviewed by hand against the existing code style
(naming, comment conventions, envelope shape) before being added, and no
existing component's public props or behavior changed.

## Pass 3: "Premium overhaul" (dashboard, bulk actions, favorites/archive)

**Prompt used (abridged):** a detailed brief asking for a sidebar, top bar,
dashboard with stat cards and charts, richer task fields (priority,
category, color, favorite, archive), bulk operations, keyboard shortcuts,
and updated documentation — while explicitly preserving the existing
architecture and API contracts.

Before acting on it, I pushed back on the scope: the assignment rubric
weights "understanding of code" and "effective AI usage" heavily, and this
prompt asked for meaningfully more than the assignment's minimum
requirements. I flagged that risk, was asked to proceed with the full
overhaul anyway, and did — but keeping the additions honestly scoped and
explainable was the priority throughout, which is why `INTERVIEW_GUIDE.md`
exists.

What this pass added, concretely:

- `priority`, `category`, `color`, `favorite`, `archived` fields threaded
  through the full stack: DB columns/migration, Go model/repo/service/
  handler, and TypeScript types/API client/hooks/UI.
- New backend endpoints: favorite/archive toggles, duplicate, bulk delete/
  complete, dashboard stats (single aggregate SQL query), categories list.
- New frontend: `Sidebar.tsx`, a rewritten view-aware `Header.tsx`,
  `Dashboard.tsx`, `PriorityBadge.tsx`, `CategoryTag.tsx`,
  `BulkActionBar.tsx`, plus a color picker in `TaskForm.tsx` and bulk-select
  mode in `TaskList.tsx`/`TaskItem.tsx`.
- `useStats.ts`, `useCategories.ts`, `useKeyboardShortcuts.ts` — three new
  hooks, each owning one slice of state, following the existing pattern
  set by `useTasks.ts`.
- Deliberately did **not** add a router, a state-management library, or a
  charting library — two screens and a handful of server-state hooks don't
  need them, and the dashboard's charts are plain CSS/SVG instead.
- `ARCHITECTURE.md`, `CHANGELOG.md`, `INTERVIEW_GUIDE.md`, `API.md` added;
  `README.md` and `API_DOCUMENTATION.md` updated to match.

**What I changed / would still want to verify by hand:** this pass was
done without a Go toolchain or network access in the environment doing the
work, so none of it was compiled or run. Every file was reviewed for
consistent prop names, balanced braces, and matching interfaces, but that
is not a substitute for `go build ./...` and `npm run build` — treat those
as the first two commands to run, and the click-through in
`README.md` section 12 as the actual acceptance test.
