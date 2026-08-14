# Agent Instructions for Chirpy

## Overview

**Chirpy** is a Go web server for a birds chirping application. It provides REST APIs for user management, authentication, and chirps (messages), with a PostgreSQL backend and JWT authentication.

- **Language**: Go 1.26.5
- **Database**: PostgreSQL (via `lib/pq`)
- **Auth**: JWT tokens + Argon2id password hashing
- **Code Generation**: sqlc for type-safe SQL

## Build & Run

```bash
# Build binary
go build -o out && ./out

# Run tests
go test ./...

# Generate database code from SQL
sqlc generate
```

**Environment Variables** (required in `.env`):
- `DB_URL` - PostgreSQL connection string (e.g., `postgres://user:pass@localhost/chirpy`)
- `PLATFORM` - deployment environment (e.g., `dev`, `prod`)
- `JWT_SECRET` - secret for signing JWT tokens
- `POLKA_KEY` - API key for Polka webhook service

## Project Structure

```
handler_*.go              # HTTP request handlers (routes)
main.go                   # Server setup, routing
json.go                   # JSON response helpers (respondWithJSON, respondWithError)
metrics.go                # Metrics/analytics handlers
readiness.go              # Health check handler
reset.go                  # Admin reset handler (dev only)
internal/
  auth/auth.go           # JWT + password hashing utilities
  auth/auth_test.go      # Auth tests
  database/              # sqlc-generated code (DO NOT EDIT)
    *.sql.go            # Type-safe database functions
    db.go               # Database connection setup
    models.go           # Query parameter types
sql/
  schema/               # Database migrations (*.sql)
  queries/              # SQL query definitions (*.sql)
```

## Key Patterns

### HTTP Handlers

All handlers follow this pattern:
```go
func (cfg *apiConfig) handlerName(w http.ResponseWriter, r *http.Request) {
    // 1. Decode request
    // 2. Validate
    // 3. Call database
    // 4. Respond with JSON or error
}
```

**Response helpers** in `json.go`:
- `respondWithJSON(w http.ResponseWriter, code int, payload interface{})`
- `respondWithError(w http.ResponseWriter, code int, message string, err error)`

### Routes & Middleware

Routes registered in `main.go`:
- `POST /api/users` - Create user
- `GET /api/healthz` - Health check
- `POST /api/login` - Authenticate
- `POST /api/chirps` - Create chirp
- `GET /api/chirps` - List chirps
- `DELETE /api/chirps/{chirpID}` - Delete (owner only)
- Admin: `POST /admin/reset`, `GET /admin/metrics`

Middleware:
- `middlewareMetricsInc` - Tracks file server hits

### Database

**Code generation via sqlc**:
1. Define queries in `sql/queries/*.sql`
2. Define schemas in `sql/schema/*.sql`
3. Run: `sqlc generate`
4. Generated code in `internal/database/*.sql.go` (DO NOT EDIT)

**Query pattern**:
```sql
-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;
```

### Authentication

- Password: Hash with `auth.HashPassword()`, verify with `auth.CheckPasswordHash()`
- JWT: Create with `auth.MakeJWT()`, validate with `auth.ValidateJWT()`
- Token type: `chirpy-access` (defined in `auth/auth.go`)

### Testing

- Auth tests in `internal/auth/auth_test.go`
- Follow table-driven test pattern
- Use `testing.T` for test functions

## Common Development Tasks

### Add a New API Endpoint

1. **Create handler** in `handler_*.go`:
   ```go
   func (cfg *apiConfig) handlerName(w http.ResponseWriter, r *http.Request) { ... }
   ```
2. **Add route** in `main.go`: `mux.HandleFunc("METHOD /api/path", cfg.handlerName)`
3. **Add database query** if needed in `sql/queries/*.sql`
4. **Run**: `sqlc generate` → `go build`

### Add Database Migration

1. Create new file in `sql/schema/` numbered sequentially (e.g., `006_table_name.sql`)
2. Write migration SQL
3. Ensure PostgreSQL syntax is correct
4. Restart server to apply

### Modify Database Query

1. Edit `sql/queries/*.sql`
2. Run: `sqlc generate`
3. Update handler code to use new generated function signature
4. Test the change

### Authentication Flow

1. **Sign up**: POST `/api/users` with `email` + `password` → creates user with hashed password
2. **Login**: POST `/api/login` with credentials → returns JWT token
3. **Protected routes**: Extract JWT from `Authorization: Bearer <token>` header
4. **Validate**: Use `auth.ValidateJWT()` before processing request

## Conventions & Tips

- **Handler naming**: `handler{Resource}{Action}` (e.g., `handlerChirpsCreate`, `handlerUsersUpdate`)
- **File naming**: lowercase snake_case for Go files (e.g., `handler_chirps_get.go`)
- **Status codes**: Use standard HTTP codes (200, 201, 400, 401, 403, 500)
- **Error handling**: Return meaningful error messages and appropriate status codes
- **JSON tags**: Use `json:"-"` to hide sensitive fields (e.g., passwords)
- **UUID**: Use `google/uuid` for ID generation
- **Timestamps**: Store in UTC, use `time.Time` fields

## Common Pitfalls

- **DB_URL not set**: Server exits with "DB_URL must be set" — check `.env` file
- **Stale generated code**: After adding SQL queries, remember to run `sqlc generate`
- **Password exposure**: Ensure `Password` field has `json:"-"` tag in response structs
- **JWT validation**: Always validate tokens from `Authorization` header before trusting claims
- **CORS**: If adding frontend, configure CORS middleware for cross-origin requests

## Useful Files to Know

- `internal/auth/auth.go` - All authentication logic
- `json.go` - Response formatting (study these for consistency)
- `sql/queries/users.sql`, `chirps.sql`, `refresh_tokens.sql` - Core business queries
- `handler_login.go` - Example of full auth flow
