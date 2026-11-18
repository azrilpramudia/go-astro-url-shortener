# Backend Documentation

This document describes the Go backend, its HTTP API, configuration, database schema, and local development workflow.

## Overview

The backend is a Go HTTP service built with Fiber. It creates short links, redirects visitors to their original URLs, and exposes link click statistics. PostgreSQL stores link records and GORM handles database operations.

The application does not currently provide user accounts, authentication, a global link-list endpoint, or a server-side delete endpoint. The frontend keeps the codes for its own list in browser storage.

## Architecture

The application follows a small layered structure:

1. `cmd/web` loads configuration and initializes the logger, database, validator, and Fiber application.
2. `internal/config` constructs infrastructure and registers middleware and routes.
3. HTTP routes dispatch to controllers.
4. Controllers parse requests and serialize responses.
5. Use cases validate input and implement link creation, lookup, redirect, expiration, and click-count behavior.
6. Repositories perform database operations through GORM.
7. Entities represent persisted records; models and converters define HTTP request and response data.

The redirect path increments the click count with a database-side atomic update. If that update fails, the error is logged and the redirect still proceeds.

## Requirements

- Go version matching `go.mod` (`go 1.27.1`)
- PostgreSQL 16 or Docker Compose
- The `migrate` CLI for database migrations
- `air` for the optional hot-reload development command
- Make

## Quick Start: Local Development

Run backend commands from the `backend` directory so `.env` is found:

```bash
cp .env.example .env
```

Edit `.env` as needed, then start PostgreSQL and apply the schema:

```bash
make db-up
make migrate-up
```

Start the API:

```bash
make run
```

For automatic rebuilds during development, use `make dev` instead. It runs Air using `backend/.air.toml`.

The API listens on `WEB_PORT` (3000 by default). The short URLs returned by the API are built from `APP_BASE_URL`; set it to the public base URL users should visit.

## Docker Compose

From `backend`, configure `.env` and run:

```bash
make app-up
```

The `app` Compose profile starts PostgreSQL, runs the migration service after PostgreSQL is healthy, and starts the backend after migrations complete. The frontend is also included in this profile.

Useful commands:

```bash
make app-logs
make app-down
```

`app-down` stops and removes the Compose containers but leaves the named PostgreSQL volume in place. Do not remove that volume unless you intend to delete the stored database data.

The backend container is built from a multi-stage Dockerfile and runs as a non-root user in a distroless image.

## Configuration

Configuration is read from environment variables and an `.env` file. The loader checks `.env` and then `../.env` relative to the process working directory. Environment variables take precedence over file values and defaults.

| Variable           | Default / example                         | Description                                                                                                                          |
| ------------------ | ----------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| `APP_NAME`         | `go-astro-url-shortener`                  | Fiber application name.                                                                                                              |
| `APP_BASE_URL`     | `http://localhost:3000` in `.env.example` | Base URL used to construct the `short_url` response. Trailing slashes are removed. Set this to the externally reachable backend URL. |
| `WEB_PORT`         | `3000`                                    | HTTP listen port.                                                                                                                    |
| `WEB_PREFORK`      | `false`                                   | Fiber prefork setting.                                                                                                               |
| `WEB_CORS_ORIGIN`  | `http://localhost:4321`                   | Allowed browser origin for CORS. Set the exact frontend origin.                                                                      |
| `LOG_LEVEL`        | `5`                                       | Numeric Logrus level: `0` panic, `1` fatal, `2` error, `3` warn, `4` info, `5` debug, `6` trace.                                     |
| `DB_HOST`          | Set in `.env`                             | PostgreSQL hostname. Use `localhost` for a local backend process or `postgres` from the Compose backend container.                   |
| `DB_PORT`          | `5432`                                    | PostgreSQL port inside the backend configuration.                                                                                    |
| `DB_USER`          | Set in `.env`                             | PostgreSQL username.                                                                                                                 |
| `DB_PASSWORD`      | Set in `.env`                             | PostgreSQL password. Do not commit real credentials.                                                                                 |
| `DB_NAME`          | Set in `.env`                             | PostgreSQL database name.                                                                                                            |
| `DB_POOL_IDLE`     | `10`                                      | Maximum idle connections.                                                                                                            |
| `DB_POOL_MAX`      | `100`                                     | Maximum open connections.                                                                                                            |
| `DB_POOL_LIFETIME` | `300`                                     | Connection lifetime in seconds.                                                                                                      |

The PostgreSQL DSN currently sets `sslmode=disable` and `TimeZone=Asia/Jakarta`. Review the database transport settings before exposing the service outside a trusted local or container network.

## HTTP API

All API errors use this JSON shape:

```json
{
  "errors": "Error description"
}
```

Successful link responses are wrapped in a `data` property:

```json
{
  "data": {
    "id": 1,
    "code": "a1B2c3D",
    "short_url": "http://localhost:3000/a1B2c3D",
    "original_url": "https://example.com/docs",
    "clicks": 0,
    "expires_at": "2030-01-01T00:00:00Z",
    "created_at": "2026-11-18T08:00:00Z"
  }
}
```

`expires_at` is omitted when the link has no expiration.

### `POST /api/links`

Create a short link. Requests are limited to 20 creations per minute by the Fiber limiter.

Request:

```json
{
  "url": "https://example.com/docs",
  "expires_at": "2030-01-01T00:00:00Z"
}
```

- `url` is required, must be a valid URL, and may be at most 2,048 characters.
- Only `http` and `https` schemes are accepted.
- `expires_at` is optional, must be a valid RFC 3339 timestamp, and must be in the future.
- The generated code is seven characters from a base62 alphabet. The service retries up to five times if a code collides with an existing record.

Example:

```bash
curl -i -X POST http://localhost:3000/api/links \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://example.com/docs"}'
```

Responses:

- `201 Created` — link created; response contains the `data` object.
- `400 Bad Request` — malformed JSON or invalid URL/expiration.
- `429 Too Many Requests` — creation rate limit exceeded.
- `500 Internal Server Error` — unexpected creation or persistence failure.

### `GET /api/links/:code`

Get one link and its current click count. This endpoint does not increment the count.

Example:

```bash
curl -i http://localhost:3000/api/links/a1B2c3D
```

Responses:

- `200 OK` — link returned under `data`.
- `404 Not Found` — code does not exist.

The detail endpoint returns the stored record even if its expiration time has passed. Expiration is enforced by the redirect endpoint.

### `GET /:code`

Resolve a short code and redirect to the original URL.

Example:

```bash
curl -i http://localhost:3000/a1B2c3D
```

Responses:

- `302 Found` — redirects to the original URL and attempts to increment the click count.
- `404 Not Found` — code does not exist.
- `410 Gone` — link has expired.

### `GET /api/hello`

Returns a simple response for checking that the HTTP service is reachable:

```json
{ "status": "ok" }
```

This endpoint does not verify PostgreSQL connectivity. There is currently no `/api/health` route.

## Database and Migrations

Migrations live in `db/migrations`. The initial migration creates the `links` table:

| Column         | Type          | Notes                                                              |
| -------------- | ------------- | ------------------------------------------------------------------ |
| `id`           | `BIGSERIAL`   | Primary key.                                                       |
| `code`         | `VARCHAR(16)` | Unique short code; generated codes are currently seven characters. |
| `original_url` | `TEXT`        | Destination URL.                                                   |
| `clicks`       | `BIGINT`      | Non-null; starts at zero.                                          |
| `expires_at`   | `TIMESTAMPTZ` | Nullable expiration timestamp.                                     |
| `created_at`   | `TIMESTAMPTZ` | Defaults to the database current time.                             |
| `updated_at`   | `TIMESTAMPTZ` | Updated by GORM on entity writes.                                  |

Apply and roll back migrations from `backend`:

```bash
make migrate-up
make migrate-down
```

Create a numbered migration:

```bash
make migrate-create name=add_link_metadata
```

`make migrate-down` rolls back one migration. `make migrate-force version=<n>` changes the migration version without running schema changes; use it only to recover a known migration state.

## Testing

The Go test package is an integration suite and connects to a separate database named `<DB_NAME>_test`. It truncates the `links` table between tests. Never point the test database configuration at production data.

Create the test database and apply its migrations:

```bash
make test-setup
```

Run the tests:

```bash
make test
```

## Middleware, Errors, and Operational Notes

- Request logging records the HTTP method, path, status, duration, and client IP as JSON.
- Fiber's recovery middleware is enabled.
- CORS allows `GET`, `POST`, and `OPTIONS`, and the `Origin`, `Content-Type`, and `Accept` headers. The allowed origin is controlled by `WEB_CORS_ORIGIN`.
- CORS is a browser policy, not authentication. The API currently has no authentication or authorization.
- The create rate limiter is process-local; multiple backend replicas do not share limiter state.
- Internal failures are logged. The HTTP error handler returns JSON with an `errors` field.
- Only create, detail lookup, and redirect operations are exposed. There are no API endpoints to list every link, update a link, or delete a stored link.

## Useful Make Targets

| Target                                  | Purpose                                             |
| --------------------------------------- | --------------------------------------------------- |
| `make run`                              | Run the backend with `go run`.                      |
| `make dev`                              | Run with Air hot reload.                            |
| `make build`                            | Build `bin/go-astro-url-shortener`.                 |
| `make test`                             | Run Go integration tests.                           |
| `make tidy`                             | Run `go mod tidy`.                                  |
| `make db-up` / `make db-down`           | Start or stop the PostgreSQL Compose service.       |
| `make migrate-up` / `make migrate-down` | Apply or roll back one migration.                   |
| `make app-up` / `make app-down`         | Start or stop the full Compose application profile. |
| `make app-logs`                         | Follow logs for the full application profile.       |
