# Chirpy

REST API server for birds-chirping app. Built as part of [Boot.dev's "Learn HTTP Servers in Go"](https://www.boot.dev/courses/learn-http-servers-in-go) course.

## What's Inside

Production-ready Go server with:

- User auth (JWT + Argon2id password hashing)
- Chirps CRUD + ownership enforcement
- PostgreSQL persistence via sqlc
- Webhook integration (Polka service)
- Metrics tracking
- Type-safe database layer (auto-generated from SQL)

## Requirements

- Go 1.26.5+
- PostgreSQL 12+
- Goose (for database migrations): `go install github.com/pressly/goose/v3/cmd/goose@latest`
- `.env` file with:

  ```env
  DB_URL=postgres://user:password@localhost:5432/chirpy
  PLATFORM=dev
  JWT_SECRET=your-secret-key-here
  POLKA_KEY=your-polka-webhook-key
  ```

## Quick Start

```bash
# Build
go build -o out

# Run
./out

# Server listens on :8080
# Health check: curl http://localhost:8080/api/healthz

# Run tests
go test ./...

# Regenerate DB code after SQL changes
sqlc generate

# Run migrations
goose -dir sql/schema postgres "$DB_URL" up
```

## API Endpoints

### Users

- `POST /api/users` — Create user. Body: `{email, password}`
- `PUT /api/users` — Update user (auth required). Body: `{email, password}`
- `POST /api/login` — Authenticate. Body: `{email, password}`. Returns: JWT token + refresh token
- `POST /api/refresh` — Refresh JWT using refresh token. Header: `Authorization: Bearer <refresh_token>`
- `POST /api/revoke` — Revoke refresh token

### Chirps

- `POST /api/chirps` — Create chirp (auth required). Body: `{body}` (max 140 chars)
- `GET /api/chirps` — List all chirps. Query params: `?author_id=<uuid>&sort=<asc|desc>`
- `GET /api/chirps/{chirpID}` — Fetch one chirp
- `DELETE /api/chirps/{chirpID}` — Delete chirp (owner only, auth required)

### Admin

- `GET /api/healthz` — Health check
- `GET /admin/metrics` — File server hit count
- `POST /admin/reset` — Reset DB (dev only)

### Webhooks

- `POST /api/polka/webhooks` — Polka user upgrade webhook

## Project Structure

```
handler_*.go              HTTP handlers by feature
main.go                   Server setup, routing, middleware
json.go                   Response helpers
metrics.go                Admin metrics handler
readiness.go              Health check
reset.go                  Dev reset handler
internal/
  auth/
    auth.go              JWT + password hashing utilities
    auth_test.go         Auth tests (table-driven)
  database/
    *.sql.go             sqlc-generated (DO NOT EDIT)
    db.go                DB connection pool
    models.go            Query param types
sql/
  schema/                Database migrations (numbered 001-005)
  queries/               SQL query definitions
```

## Development Patterns

### Add New Endpoint

1. Create handler in new file `handler_<resource>_<action>.go`:

   ```go
   func (cfg *apiConfig) handler<Name>(w http.ResponseWriter, r *http.Request) {
       // decode JSON
       // validate
       // call cfg.db.Query()
       // respond
   }
   ```

2. Register route in `main.go`:

   ```go
   mux.HandleFunc("POST /api/path", cfg.handler<Name>)
   ```

3. Add DB query in `sql/queries/*.sql` if needed:

   ```sql
   -- name: QueryName :one
   SELECT * FROM table WHERE id = $1;
   ```

4. Generate: `sqlc generate`

5. Build: `go build`

### Add Database Migration

1. Create file in `sql/schema/` (numbered sequentially: `006_name.sql`)
2. Write PostgreSQL SQL with migration comment header:
   ```sql
   -- +goose Up
   CREATE TABLE example (
     id UUID PRIMARY KEY,
     created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
   );

   -- +goose Down
   DROP TABLE example;
   ```
3. Apply migration: `goose -dir sql/schema postgres "$DB_URL" up`
4. Rollback if needed: `goose -dir sql/schema postgres "$DB_URL" down`

## Authentication Flow

1. **Sign up**: `POST /api/users` → password hashed with Argon2id
2. **Login**: `POST /api/login` → returns JWT access token + refresh token
3. **Protected routes**: Extract JWT from `Authorization: Bearer <token>` header
4. **Validate**: Call `auth.ValidateJWT()` in handler
5. **Refresh**: Use refresh token to get new JWT via `POST /api/refresh`

## Key Technologies

| Tool | Purpose |
| ------ | --------- |
| `goose` | Database migration management |
| `sqlc` | Type-safe DB code generation from SQL |
| `golang-jwt/jwt` | JWT creation + validation |
| `alexedwards/argon2id` | Password hashing |
| `lib/pq` | PostgreSQL driver |
| `google/uuid` | ID generation |
| `godotenv` | Environment variable loading |

## Conventions

- Handlers: `handler{Resource}{Action}` (e.g., `handlerChirpsCreate`)
- Files: lowercase snake_case (e.g., `handler_chirps_get.go`)
- Timestamps: UTC, stored as `time.Time`
- Sensitive fields: Use `json:"-"` tag (e.g., passwords)
- Status codes: Standard HTTP (200, 201, 400, 401, 403, 404, 500)

## Common Pitfalls

| Issue | Fix |
| ------- | ----- |
| Server exits "DB_URL must be set" | Check `.env` file loaded + `DB_URL` exists |
| Handlers don't match new DB functions | Forgot `sqlc generate` after SQL edit |
| Password exposed in JSON response | Add `json:"-"` tag to Password field |
| JWT validation always fails | Check `Authorization: Bearer <token>` format exact |
| Refresh token doesn't work | Verify refresh token sent to correct endpoint, not access token |

## Testing

```bash
# Run all tests
go test ./...

# Run specific package
go test ./internal/auth

# Table-driven test example in auth_test.go
```

## Migration Management

**Goose** handles all DB migrations. Workflow:

```bash
# Apply all pending migrations
goose -dir sql/schema postgres "$DB_URL" up

# Check migration status
goose -dir sql/schema postgres "$DB_URL" status

# Rollback last migration
goose -dir sql/schema postgres "$DB_URL" down

# Create new migration file
goose -dir sql/schema create <name> sql
```

Each migration file needs `-- +goose Up` and `-- +goose Down` markers. Goose tracks applied migrations in DB to prevent re-running.

## Learning Resources

- [Boot.dev HTTP Servers Course](https://www.boot.dev/courses/learn-http-servers-in-go)
- [Goose Documentation](https://github.com/pressly/goose)
- [sqlc Documentation](https://sqlc.dev/)
- [golang-jwt Guide](https://github.com/golang-jwt/jwt)
- [PostgreSQL Docs](https://www.postgresql.org/docs/)
