# Suggested Git Commit Sequence

If you want your repository history to show incremental development instead
of a single commit, initialize the repo and commit in this order:

```bash
git init
git add .gitignore README.md
git commit -m "chore: initial project scaffolding and gitignore"

git add backend/schema.sql
git commit -m "db: add MySQL schema for tasks table with seed data"

git add backend/go.mod backend/internal/config
git commit -m "backend: add configuration loader (env vars + .env support)"

git add backend/internal/models
git commit -m "backend: add Task model and request validation"

git add backend/internal/repository
git commit -m "backend: add MySQL repository layer (CRUD, repository pattern)"

git add backend/internal/service
git commit -m "backend: add service layer with business validation"

git add backend/internal/handler
git commit -m "backend: add HTTP handlers and JSON response envelope"

git add backend/internal/middleware backend/internal/routes
git commit -m "backend: add CORS/logging middleware and route registration"

git add backend/cmd
git commit -m "backend: wire application entrypoint with graceful shutdown"

git add backend/.env.example
git commit -m "backend: add example environment configuration"

git add frontend/package.json frontend/tsconfig*.json frontend/vite.config.ts frontend/index.html frontend/public
git commit -m "frontend: scaffold Vite + React + TypeScript project"

git add frontend/src/types frontend/src/api
git commit -m "frontend: add task types and API client"

git add frontend/src/hooks
git commit -m "frontend: add useTasks hook for data fetching and CRUD state"

git add frontend/src/components/StatusBadge.tsx frontend/src/components/TaskItem.tsx frontend/src/components/TaskList.tsx frontend/src/components/EmptyState.tsx
git commit -m "frontend: add task list, task card, and status badge components"

git add frontend/src/components/TaskForm.tsx frontend/src/components/Modal.tsx frontend/src/components/ConfirmDialog.tsx
git commit -m "frontend: add add/edit task form and delete confirmation modal"

git add frontend/src/components/SearchFilterBar.tsx frontend/src/components/Header.tsx frontend/src/components/Toast.tsx
git commit -m "frontend: add search/filter/sort toolbar, header, and toast notifications"

git add frontend/src/App.tsx frontend/src/main.tsx frontend/src/index.css frontend/src/vite-env.d.ts
git commit -m "frontend: wire dashboard UI and apply visual design system"

git add frontend/.env.example
git commit -m "frontend: add example environment configuration"

git add API_DOCUMENTATION.md AI_PROMPTS.md GIT_COMMITS.md
git commit -m "docs: add API documentation, AI usage notes, and commit guide"
```

Push to GitHub:

```bash
git branch -M main
git remote add origin <your-repo-url>
git push -u origin main
```

---

## Addendum: professional hardening pass

If these were applied as their own follow-up commits:

```bash
git add backend/internal/docs backend/internal/routes/routes.go
git commit -m "backend: embed OpenAPI spec, serve Swagger UI at /docs"

git add backend/migrations backend/schema.sql
git commit -m "db: split schema into numbered up/down migrations, keep schema.sql for one-shot setup"

git add backend/Dockerfile frontend/Dockerfile frontend/nginx.conf docker-compose.yml .gitignore
git commit -m "infra: add Dockerfiles and docker-compose for mysql+backend+frontend"

git add Makefile
git commit -m "infra: add Makefile with dev/build/docker shortcuts"

git add frontend/src/constants.ts frontend/src/App.tsx frontend/src/components/TaskForm.tsx frontend/src/api/client.ts
git commit -m "frontend: centralize constants, add client-side request timeout"

git add frontend/src/components/SkeletonList.tsx frontend/src/components/TaskList.tsx frontend/src/index.css
git commit -m "frontend: replace loading text with skeleton loaders"

git add frontend/src/components/ErrorBoundary.tsx frontend/src/main.tsx
git commit -m "frontend: add app-level error boundary with friendly fallback page"

git add frontend/package.json frontend/.eslintrc.cjs
git commit -m "frontend: fix broken lint script by adding missing eslint devDependencies"

git add API_DOCUMENTATION.md README.md AI_PROMPTS.md GIT_COMMITS.md
git commit -m "docs: document Docker/Makefile/Swagger/migrations additions"
```
