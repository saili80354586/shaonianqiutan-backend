# CLAUDE.md

Behavioral guidelines for 少年球探 backend development.
Merge with project-specific instructions as needed.

**Tradeoff:** These guidelines bias toward caution over speed.
For trivial tasks, use judgment.

## Project Overview

少年球探 (Youth Scout) - 青少年足球成长服务平台后端 API 服务
新官网 Go 后端

## Tech Stack

- Language: Go 1.25
- Web Framework: Gin 1.12
- ORM: GORM 1.31
- Database: SQLite (gorm.io/driver/sqlite)
- Auth: JWT (golang-jwt/jwt/v5)
- Validation: Gin binding + go-playground/validator
- Config: godotenv
- Crypto: golang.org/x/crypto (bcrypt)
- WebSocket: gorilla/websocket
- CORS: gin-contrib/cors

## 1. Think Before Coding

**Don't assume. Don't hide confusion. Surface tradeoffs.**

Before implementing:
- State your assumptions explicitly. If uncertain, ask.
- If multiple interpretations exist, present them - don't pick silently.
- If a simpler approach exists, say so. Push back when warranted.
- If something is unclear, stop. Name what's confusing. Ask.

## 2. Simplicity First

**Minimum code that solves the problem. Nothing speculative.**

- No features beyond what was asked.
- No abstractions for single-use code.
- No "flexibility" or "configurability" that wasn't requested.
- Use Go standard library when possible before adding dependencies.
- If you write 200 lines and it could be 50, rewrite it.

Ask yourself: "Would a senior Go engineer say this is overcomplicated?"
If yes, simplify.

## 3. Surgical Changes

**Touch only what you must. Clean up only your own mess.**

When editing existing code:
- Don't "improve" adjacent code, comments, or formatting.
- Don't refactor things that aren't broken.
- Match existing style, even if you'd do it differently.
- Follow Go idioms and `gofmt` formatting.
- If you notice unrelated dead code, mention it - don't delete it.
- Don't change database schema (migrations) without explicit approval.

When your changes create orphans:
- Remove imports/variables/functions that YOUR changes made unused.
- Run `go mod tidy` to clean up dependencies.
- Don't remove pre-existing dead code unless asked.

The test: Every changed line should trace directly to the user's request.

## 4. Goal-Driven Execution

**Define success criteria. Loop until verified.**

Transform tasks into verifiable goals:
- "Add validation" → "Write unit tests for invalid inputs, then make them pass"
- "Fix the bug" → "Write a test that reproduces it, then make it pass"
- "Refactor X" → "Ensure `go test ./...` passes before and after"

For multi-step tasks, state a brief plan:
```
1. [Step] → verify: [check]
2. [Step] → verify: [check]
3. [Step] → verify: [check]
```

Strong success criteria let you loop independently.
Weak criteria ("make it work") require constant clarification.

## Go Idioms & Style

- Follow standard Go formatting (`gofmt`, `goimports`)
- Use `camelCase` for unexported, `PascalCase` for exported
- Error handling: check errors immediately, wrap with context using `fmt.Errorf("...: %w", err)`
- Prefer explicit over implicit
- Use structs with JSON tags for serialization
- Context propagation: pass `context.Context` through call chains

## Project Structure Rules

Existing structure (follow it):
```
cmd/               → Application entry point
config/            → Configuration loading
controllers/       → HTTP handlers (Gin)
middleware/        → Gin middleware (auth, CORS, logging)
models/            → GORM models / domain entities
repositories/      → Data access layer (GORM queries)
routes/            → Route definitions
services/          → Business logic layer
utils/             → Shared utilities
wshub/             → WebSocket hub
```

- **Controllers** only handle HTTP concerns (bind JSON, call service, return response)
- **Services** contain business logic, no HTTP or DB directly
- **Repositories** handle all GORM queries, no business logic
- **Models** define GORM structs with JSON tags

## API Design Rules

- RESTful conventions: GET /resource, POST /resource, PUT /resource/:id, DELETE /resource/:id
- Consistent response format: `gin.H{"success": bool, "data": any, "message": string}`
- Use Gin binding tags for input validation (`binding:"required"`)
- Return appropriate HTTP status codes (200, 201, 400, 401, 404, 500)
- Group routes logically

## Database Rules (GORM + SQLite)

- Define models with GORM tags and JSON tags
- Use `AutoMigrate` for schema changes, but backup data first
- Use `db.Where().First()` for single record, `db.Find()` for lists
- Eager loading with `Preload()` when needed
- Transactional operations: use `db.Transaction()`
- Don't use raw SQL unless necessary for performance
- SQLite specific: respect WAL mode and busy timeout settings

## Security Rules

- Never log sensitive data (passwords, tokens, payment info)
- Hash passwords with `bcrypt` before storage
- Validate all user inputs with Gin binding + custom validators
- JWT middleware on protected routes
- CORS configured appropriately
- SQL injection prevention: always use GORM parameterized queries

## WebSocket Rules (wshub)

- Keep hub logic isolated in wshub package
- Handle connection lifecycle (connect, message, disconnect, heartbeat)
- Don't block the hub with long-running operations
- Use goroutines for per-connection handlers

## Testing Checklist

Before declaring a task complete:
- [ ] `go build ./...` compiles without errors
- [ ] `go vet ./...` clean
- [ ] `gofmt -l .` shows no unformatted files
- [ ] API endpoints return expected response format
- [ ] GORM operations work correctly
- [ ] No sensitive data in logs or responses
- [ ] JWT protected routes reject unauthorized requests
- [ ] WebSocket connections handle edge cases gracefully
