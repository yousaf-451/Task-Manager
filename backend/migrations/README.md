# Database migrations

Plain numbered SQL files, applied in order. No migration framework/binary is
required for this project's scope, but the structure follows the standard
`NNNN_description.up.sql` / `.down.sql` convention so it can be dropped
straight into a tool like `golang-migrate` later without renaming anything.

```
migrations/
├── up/     # forward migrations - auto-applied by docker-compose on first boot
│   ├── 0001_init_schema.sql
│   ├── 0002_seed_data.sql
│   ├── 0003_add_task_metadata.sql
│   ├── 0004_add_users_and_ownership.sql
│   └── 0005_add_auth.sql
└── down/   # rollback scripts - run manually, never auto-applied
    ├── 0001_init_schema.sql
    ├── 0002_seed_data.sql
    ├── 0003_add_task_metadata.sql
    ├── 0004_add_users_and_ownership.sql
    └── 0005_add_auth.sql
```

Migration `0004` adds a `users` table and a `user_id` foreign key on `tasks`,
so tasks are scoped to (and can never exist without) a valid user — this is
the referential-integrity fix. It is backward compatible: existing tasks are
backfilled to a default user before the column is made `NOT NULL`.

Migration `0005` adds real authentication: a `password_hash` column on
`users` and a `sessions` table. Signup/login/logout now work with actual
credentials and a session-cookie flow (see `internal/service/auth_service.go`
and `internal/middleware/auth.go`) instead of the old self-reported
`X-User-Id` header. The three demo users seeded in `0002`/`0004`
(Yousaf/Arham/Aaqib) get an empty password hash and can't log in until a
real password is set for them directly in the database — every user created
through the app from now on goes through `POST /api/auth/signup` instead.

## Applying manually (without Docker)

```bash
mysql -u task_user -p task_manager < migrations/up/0001_init_schema.sql
mysql -u task_user -p task_manager < migrations/up/0002_seed_data.sql   # optional seed data
```

## Rolling back

```bash
mysql -u task_user -p task_manager < migrations/down/0002_seed_data.sql
mysql -u task_user -p task_manager < migrations/down/0001_init_schema.sql
```

## With Docker Compose

`docker-compose.yml` mounts `migrations/up` into MySQL's
`/docker-entrypoint-initdb.d`, so both files run automatically the **first**
time the `mysql` container starts with an empty data volume. Files under
`down/` are intentionally not mounted, so they can never run automatically.

> For a brand-new local (non-Docker) setup, `../schema.sql` is the fastest
> path — it creates the database, application user, table, and seed data in
> one shot. This `migrations/` folder is the source of truth for schema
> changes going forward.
