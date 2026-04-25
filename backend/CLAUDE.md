# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 限制條件
- **不要**修改或讀取 `*_gen.go`。
- 在未經許可下，**不要**自動執行任何 `git push` 命令。
- **不要**使用 AI 生成的虛假資料來測試功能。
- 

## Commands

```bash
# Run the server
go run ./cmd/main.go

# Build
go build ./...

# Run all tests
go test ./...

# Run a single package's tests
go test ./internal/services/...

# Run the enum code generator (after modifying enum types)
go generate ./internal/enums/...

# Or run directly
go run ./cmd/enumx-gen
```

## Architecture

This is a **Go event-sourcing accounting system** using SQLite, `sqlx`, and `golang-migrate`. The server runs on `:8080` using the standard `net/http` package.

### Event Flow (the core pattern)

Every write goes through this chain:

1. **`handler/`** — decodes HTTP request into `cmd.AppendCmd`
2. **`services/EventStoreService.Append`** — the single write entry point
3. **`pipelines/`** — validates and enriches the event payload before writing; each `EventType` maps to a registered `PipelineRunner` in `factory/base.go`
4. **`UnitOfWork.Do`** — wraps all writes in a single SQLite transaction:
   - Optimistic version lock (`VersionRepository.UpdateIfVersionMatch`)
   - Event insert
   - All `Projection.Apply` calls (read-model updates)
   - Snapshot every 50 versions

### Key Abstractions

**`enumx` package** (`internal/pkg/enumx/`): A type-safe enum system backed by a global registry. Enum types are declared with `//enumx:enum` comments and code-generated via `cmd/enumx-gen`. Always run `go generate ./internal/enums/...` after modifying enum constants.

**Projections** (`internal/services/projection/`): Read-model builders. Each projection implements the `Projection` interface (`Name()`, `Apply()`). `NewProjection()` in `base.go` is the authoritative list — add new projections there. They run inside the same transaction as the event write.

**Pipelines** (`internal/services/pipelines/`): Pre-write business logic. A `TypedPipeline[S, P]` combines a state type `S` (fetched from DB) and a payload type `P` (from JSON). Pipelines validate domain rules before the event is persisted. Each event type is registered in `factory/base.go`.

**UnitOfWork** (`internal/repos/unit_of_work/event_store/`): `EventStoreRepositories` bundles all transactional repos (Event, Version, Snap, Check, Projection) and is only valid inside `UnitOfWork.Do`. Services never hold a direct DB reference — they receive repos through the UoW callback.

**`query.Repo`** (`internal/repos/query/`): Read-only query layer used outside transactions (e.g., for pipeline state lookups and report queries).

### Configuration

All config is environment-variable driven (see `internal/bootstrap/config.go`). Key defaults:
- `DB_DRIVER=sqlite`, `DATA_SOURCE=../output/akatengu/db.db`
- `JWT_SECRET=this is a secret JWT`
- `APP_ENV=local`

### Database

- Migrations live in `internal/database/migrations/` (embedded via `//go:embed`) and run automatically on startup.
- Seeds in `internal/database/seeds/` — `accounts.sql` runs on every startup (idempotent).
- SQLite is configured for WAL mode with a single connection (`MaxOpenConns=1`).

### Cache

`internal/pkg/cache/` provides a two-layer cache (local in-memory + optional distributed). `cache_service/account.go` wraps it for account lookups used by pipelines.