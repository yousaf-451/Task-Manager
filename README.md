# Task Manager — Granet Technologies Assignment

A full-stack Task Management application:

- **Frontend:** React + TypeScript (Vite)
- **Backend:** Go (Golang) REST API, layered architecture
- **Database:** MySQL

---

## 1. Project structure

```
task-manager/
├── backend/
│   ├── cmd/api/main.go              # entrypoint: wires config → db → repo → service → handler → router
│   ├── internal/
│   │   ├── config/                  # env var loading
│   │   ├── docs/                    # embedded OpenAPI spec + Swagger UI handler
│   │   ├── models/                  # Task entity + request DTOs + validation
│   │   ├── repository/              # MySQL data-access layer (repository pattern)
│   │   ├── service/                 # business logic / use cases
│   │   ├── handler/                 # HTTP handlers (request/response only)
│   │   ├── routes/                  # route registration
│   │   └── middleware/              # CORS, request logging
│   ├── migrations/                  # numbered up/down SQL migrations (source of truth for schema)
│   ├── schema.sql                   # one-shot local setup: db + user + schema + seed data
│   ├── Dockerfile
│   ├── .env.example
│   └── go.mod
├── frontend/
│   ├── src/
│   │   ├── api/                     # fetch client (with timeout) + task API calls
│   │   ├── types/                   # shared TypeScript types
│   │   ├── hooks/                   # useTasks, useStats, useCategories, useKeyboardShortcuts
│   │   ├── components/              # Sidebar, Header, Dashboard, TaskList, TaskForm, SearchFilterBar,
│   │   │                            # BulkActionBar, PriorityBadge, CategoryTag, StatusBadge, Modal,
│   │   │                            # ConfirmDialog, EmptyState, SkeletonList, Toast, ErrorBoundary
│   │   ├── constants.ts             # centralized timing/length constants
│   │   ├── App.tsx
│   │   └── main.tsx
│   ├── index.html
│   ├── Dockerfile
│   ├── nginx.conf
│   ├── package.json
│   └── .env.example
├── docker-compose.yml               # mysql + backend + frontend, one command
├── Makefile                         # common dev/build/docker shortcuts
├── API.md                           # quick endpoint index
├── API_DOCUMENTATION.md             # full API reference with curl examples
├── ARCHITECTURE.md                  # layered design + key decisions
├── CHANGELOG.md                     # what changed in the premium overhaul
├── INTERVIEW_GUIDE.md               # anticipated interview Q&A
├── AI_PROMPTS.md
└── README.md
```

---

## 2. Prerequisites

- Go 1.22+
- Node.js 18+ and npm
- MySQL 8+ running locally (or via Docker)

---

## 3. Database setup

```bash
mysql -u root -p < backend/schema.sql
```

This creates the `task_manager` database, a `task_user` application user, the `tasks` table, and a few seed rows.
Adjust the username/password in `schema.sql` if you need different credentials.

---

## 4. Backend setup

```bash
cd backend
cp .env.example .env      # edit DB credentials if needed
go mod tidy                # downloads github.com/go-sql-driver/mysql
go run ./cmd/api
```

The API starts on `http://localhost:8080` by default. Verify it's up:

```bash
curl http://localhost:8080/health
```

### Backend environment variables (`backend/.env`)

| Variable | Description | Default |
|---|---|---|
| `SERVER_PORT` | HTTP port | `8080` |
| `SERVER_HOST` | Bind address | `0.0.0.0` |
| `CORS_ALLOWED_ORIGINS` | Comma-separated list of allowed frontend origins | `http://localhost:5173` |
| `DB_HOST` / `DB_PORT` | MySQL host/port | `127.0.0.1` / `3306` |
| `DB_USER` / `DB_PASSWORD` | MySQL credentials | `task_user` / `task_password` |
| `DB_NAME` | Database name | `task_manager` |
| `DB_MAX_OPEN_CONNS` | Connection pool size | `25` |
| `DB_MAX_IDLE_CONNS` | Idle connections kept open | `10` |
| `DB_CONN_MAX_LIFETIME_MIN` | Max connection lifetime (minutes) | `5` |

---

## 5. Frontend setup

```bash
cd frontend
cp .env.example .env       # points VITE_API_BASE_URL at the backend
npm install
npm run dev
```

Open `http://localhost:5173`.

---

## 6. Running both together

Two terminals:

```bash
# Terminal 1
cd backend && go run ./cmd/api

# Terminal 2
cd frontend && npm run dev
```

Or, with `make` (see `Makefile` for the full list of targets):

```bash
make backend-run     # terminal 1
make frontend-dev     # terminal 2
```

---

## 7. Running everything with Docker Compose (no local Go/Node/MySQL needed)

```bash
docker compose up --build
# or: make docker-up
```

This builds and starts three containers:

| Service | Container port | Host port |
|---|---|---|
| `mysql` | 3306 | 3306 |
| `backend` (Go API) | 8080 | 8080 |
| `frontend` (React, served by nginx) | 80 | 3000 |

On first boot, MySQL automatically runs the SQL files in
`backend/migrations/up/` (schema + seed data) via
`docker-entrypoint-initdb.d`. Open `http://localhost:3000` for the app,
`http://localhost:8080/health` for the API health check, and
`http://localhost:8080/docs` for interactive API documentation.

Stop everything with `docker compose down` (or `make docker-down`); add
`-v` to also drop the MySQL data volume and start fresh next time.

---

## 8. API documentation (Swagger / OpenAPI)

The API ships with a machine-readable OpenAPI 3.0 spec, embedded directly
into the Go binary — no separate doc server to run:

- **Interactive UI:** `http://localhost:8080/docs`
- **Raw spec:** `http://localhost:8080/openapi.yaml`
- **Source file:** `backend/internal/docs/openapi.yaml`
- **Quick endpoint index:** `API.md`
- **Narrative reference with curl examples:** `API_DOCUMENTATION.md`

---

## 9. Request flow

```
React component (e.g. TaskForm)
   │  fetch() via src/api/taskApi.ts
   ▼
Go routes (internal/routes) — matches method + path
   │
   ▼
Handler (internal/handler) — decodes JSON, calls service, encodes response
   │
   ▼
Service (internal/service) — validates business rules, orchestrates the use case
   │
   ▼
Repository (internal/repository) — builds and runs SQL against MySQL
   │
   ▼
MySQL — persists/returns rows
```

Each layer only talks to the layer directly below it:

- **Handlers** never write SQL.
- **Services** never touch `net/http` (no `http.Request`/`ResponseWriter`).
- **Repositories** never contain business rules — they just do CRUD against `tasks`.

This makes it straightforward to, for example, swap MySQL for another database (only `repository/` changes), or add a CLI on top of the same service layer.

---

## 10. Features implemented

**Backend**
- Full CRUD REST API (`POST/GET/PUT/DELETE /api/tasks`)
- Layered architecture (routes → handler → service → repository → models)
- Input validation (title required/length, valid date, valid status/priority enum)
- Consistent JSON error envelope with correct HTTP status codes (400/404/500)
- CORS middleware, configurable via env var
- Request logging middleware
- Graceful shutdown on SIGINT/SIGTERM
- Search, status/priority/category/favorite filters, and sort (due date, created date, or priority) supported server-side, with indexed columns for query performance
- Favorite/archive toggles, duplicate, and bulk delete/complete endpoints
- Dashboard aggregate endpoint (`/api/tasks/stats`) computed with a single SQL query, and a `/api/categories` endpoint for filter/autocomplete
- Archived tasks are excluded from listings and stats by default (soft delete)
- Health check endpoint (`/health`) for container orchestration / uptime monitoring
- Embedded OpenAPI 3.0 spec with a Swagger UI at `/docs`
- Numbered SQL migrations (`backend/migrations/up|down`) as the schema source of truth
- Dockerfile (multi-stage, non-root, with a container healthcheck)

**Frontend**
- Responsive sidebar (Dashboard / Tasks) and view-aware top bar
- Dashboard with stat cards, a status breakdown chart, and an upcoming-deadlines list — no external chart library
- Beautiful task cards: color accent, priority badge, category tag, favorite star, overdue highlighting
- Add / Edit task in a modal form, with client-side validation mirroring the backend, plus priority/category/color fields
- Delete with confirmation dialog; bulk-select mode with a bulk action bar (complete / delete many at once)
- Duplicate and archive quick actions per task
- Status badges (Pending / In Progress / Completed) and priority badges
- Search (debounced), filter by status/priority/category/favorite, multiple sort options
- Keyboard shortcuts (`/` to search, `n` to add a task)
- Skeleton loaders while fetching (instead of a bare "Loading…" line), empty state, error banner, and toast notifications for success/failure
- App-level error boundary with a friendly fallback page for unexpected render errors
- Optimistic favorite toggling, with rollback on failure
- Client-side request timeout (10s) so a hung request never leaves the UI stuck
- Centralized constants (`src/constants.ts`) for debounce/timeout/length values instead of scattered magic numbers
- Fully responsive down to mobile widths, with an off-canvas sidebar drawer
- ESLint configured and working (`npm run lint`)

**Infrastructure**
- `docker-compose.yml` to run MySQL + backend + frontend with one command
- `Makefile` with shortcuts for local dev, Docker, and builds

---

## 11. Further reading

- `ARCHITECTURE.md` — layered design, data model, and key decisions explained
- `CHANGELOG.md` — what changed in the premium overhaul vs. the original submission
- `INTERVIEW_GUIDE.md` — anticipated interview questions with answers grounded in this code

## 12. Before you submit: verify the build

This project was extended in an environment without a Go toolchain or
network access, so the additions could not be compiled or run locally as
part of that work. Before treating this as submission-ready, run both
builds yourself:

```bash
cd backend && go build ./...
cd frontend && npm install && npm run build
```

Fix anything either command flags, and give the app a manual click-through
(dashboard stats, add/edit/duplicate/archive/favorite a task, bulk select)
before recording the demo video.

## 13. AI usage

See `AI_PROMPTS.md` for the key prompts used and what was AI-generated vs. manually reviewed/adjusted.

## 14. Suggested Git workflow

See `GIT_COMMITS.md` for a suggested sequence of commits if you want the repository history to reflect incremental development rather than a single commit.
This is a test update for demo purposes.
Demo update line
