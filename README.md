# VitalTag — API

Backend for **VitalTag**, a system for sharing clinical information in workplace emergency
situations. Workers register their clinical information and receive a QR Code containing a link to
a password-protected public page. The password is printed alongside the code.

Project for the **Integrated Project Development Course** — Systems Analysis and Development.

Frontend: **vitaltag-web** (separate repository).

## Stack

- **Go 1.26+** — REST API
- **chi** — routing and middleware
- **pgx** — PostgreSQL driver
- **PostgreSQL 16** — database
- **golang-jwt** — session tokens · **bcrypt** — password hashing
- **skip2/go-qrcode** — QR Code generation
- **golang-migrate** — schema versioning

## Architecture

```
cmd/api               entry point
internal/handler      HTTP layer: routing, parsing, and input validation
internal/service      business rules (separate passwords, lockout, cascading deletion)
internal/repository   PostgreSQL access
internal/model        domain entities
internal/middleware   JWT authentication, rate limiting, and logging
migrations            versioned schema scripts
```

## Running locally (Phase 2)

```bash
cp .env.example .env
docker compose up -d db
migrate -path migrations -database "$DATABASE_URL" up
go run ./cmd/api
```

## Planned endpoints

| Method | Route | Description |
|---|---|---|
| POST | `/api/v1/auth/registro` | creates an account (RF01) |
| POST | `/api/v1/auth/login` | authenticates and issues a token (RF02) |
| GET | `/api/v1/perfil` | returns the authenticated user's profile (RF11) |
| POST | `/api/v1/perfil` | creates or updates clinical information (RF04, RF05, RF12) |
| GET | `/api/v1/perfil/qrcode` | returns the QR Code as PNG/SVG (RF06) |
| DELETE | `/api/v1/perfil` | deletes the complete record (RF10) |
| POST | `/api/v1/publico/{slug}` | validates the public password and returns the data (RF08, RF09) |

## Documentation

The complete design (requirements, user stories, work breakdown structure, roadmap, prototype,
and data model) is included in the Phase 1 course deliverable.
