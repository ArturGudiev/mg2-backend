# MemoryGuard

Go server for MemoryGuard (memory nodes, cards, card items), structured like `dashboard-go`.

## Stack

- Gin HTTP API
- Ent ORM (Postgres)
- Google Wire DI
- Swagger UI (`/swagger/index.html`) via swaggo
- JWT auth (access + refresh cookies; TTLs via env)

## Roles

Users have a required `role`: `admin` | `user` (default `user` on register).

Only **admins** can create/update the `shared` flag on memory nodes, cards, and card items.

After deploying the role column, set existing users to admin once:

```bash
MIGRATE_USERS_ROLE_ADMIN=true go run .
# then remove the env var / set it false
```

## Entities

| Table | Purpose |
|-------|---------|
| `memory_nodes` | Hierarchy of memory nodes with card refs, aliases, priorities, groups |
| `cards` | Flashcards; `question` / `answer` store **card item IDs** |
| `card_items` | Typed content blocks (TEXT, CODE, IMAGE, FORMULA, …) with their own IDs |

Each of the three can be marked `shared: true`. Shared entities are readable (and practice-updatable for cards) by every authenticated user; only the owner can edit or delete content.

## Setup

From this `server/` directory:

```bash
cp .env.example .env
# edit DATABASE_URL / DB_* as needed

go generate ./ent
cd app && wire
cd ..

# Generate / refresh Swagger docs
go install github.com/swaggo/swag/cmd/swag@v1.16.6
swag init -g main.go -o ./docs

go run .
```

Default port: `3033` (same as the old Node server).

Swagger UI: `http://localhost:3033/swagger/index.html`
