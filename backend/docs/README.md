# Backend Documentation

This document describes the Go backend, its HTTP API, configuration, database schema, and local development workflow.

## Overview

The backend is a Go HTTP service built with Fiber. It creates short links, redirects visitors to their original URLs, and provides authenticated link management and per-link analytics. PostgreSQL stores link and analytics records, GORM handles database operations, and Redis buffers click counters and caches redirects.

## Architecture

The application follows a small layered structure:

1. `cmd/web` loads configuration and initializes the logger, database, validator, and Fiber application.
2. `internal/config` constructs infrastructure and registers middleware and routes.
3. HTTP routes dispatch to controllers.
4. Controllers parse requests and serialize responses.
5. Use cases validate input and implement cookie-based authentication, link management, redirect, click counting, analytics, and retention.
6. Repositories perform database operations through GORM.
7. Entities represent persisted records; models and converters define HTTP request and response data.

The redirect path buffers click counts in Redis and records an analytics event. If analytics storage fails, the error is logged and the redirect still proceeds.

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
To enable country analytics in Compose, place a MaxMind-compatible country database under `backend/data/` and set `GEOIP_DB_PATH` to its path relative to `backend` (for example, `./data/GeoLite2-Country.mmdb`). Compose mounts that directory read-only into the backend. The database is not distributed with this project.

## Configuration

Configuration is read from environment variables and an `.env` file. The loader checks `.env` and then `../.env` relative to the process working directory. Environment variables take precedence over file values and defaults.

| Variable                   | Default / example                         | Description                                                                                                                          |
| -------------------------- | ----------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| `APP_NAME`                 | `go-astro-url-shortener`                  | Fiber application name.                                                                                                              |
| `APP_BASE_URL`             | `http://localhost:3000` in `.env.example` | Base URL used to construct the `short_url` response. Trailing slashes are removed. Set this to the externally reachable backend URL. |
| `WEB_PORT`                 | `3000`                                    | HTTP listen port.                                                                                                                    |
| `WEB_PREFORK`              | `false`                                   | Fiber prefork setting.                                                                                                               |
| `WEB_CORS_ORIGIN`          | `http://localhost:4321`                   | Allowed browser origin for CORS. Set the exact frontend origin.                                                                      |
| `LOG_LEVEL`                | `5`                                       | Numeric Logrus level: `0` panic, `1` fatal, `2` error, `3` warn, `4` info, `5` debug, `6` trace.                                     |
| `DB_HOST`                  | Set in `.env`                             | PostgreSQL hostname. Use `localhost` for a local backend process or `postgres` from the Compose backend container.                   |
| `DB_PORT`                  | `5432`                                    | PostgreSQL port inside the backend configuration.                                                                                    |
| `DB_USER`                  | Set in `.env`                             | PostgreSQL username.                                                                                                                 |
| `DB_PASSWORD`              | Set in `.env`                             | PostgreSQL password. Do not commit real credentials.                                                                                 |
| `DB_NAME`                  | Set in `.env`                             | PostgreSQL database name.                                                                                                            |
| `DB_POOL_IDLE`             | `10`                                      | Maximum idle connections.                                                                                                            |
| `DB_POOL_MAX`              | `100`                                     | Maximum open connections.                                                                                                            |
| `DB_POOL_LIFETIME`         | `300`                                     | Connection lifetime in seconds.                                                                                                      |
| `JWT_SECRET`               | Set in `.env`                             | Signing secret; must contain at least 32 characters.                                                                                 |
| `COOKIE_SECURE`            | `false`                                   | Mark authentication and analytics cookies as Secure; enable when serving over HTTPS.                                                 |
| `REDIS_ADDR`               | `localhost:6379`                          | Redis address; use `redis:6379` from the Compose backend container.                                                                  |
| `CACHE_TTL_SECONDS`        | `600`                                     | Redirect cache lifetime.                                                                                                             |
| `CLICK_FLUSH_SECONDS`      | `15`                                      | Interval for flushing buffered click counts to PostgreSQL.                                                                           |
| `ANALYTICS_RETENTION_DAYS` | `90`                                      | Retention period for detailed analytics events. Expired events are removed daily.                                                    |
| `GEOIP_DB_PATH`            | Empty                                     | Optional local `.mmdb` path for country lookup. When unset, country analytics are unavailable.                                       |

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

Create a short link. Requests are limited by the Fiber limiter (20 creations per minute by default).

Request:

```json
{
  "url": "https://example.com/docs",
  "expires_at": "2030-01-01T00:00:00Z",
  "custom_code": "project-docs",
  "tags": ["docs", "launch"]
}
```

- `url` is required, must be a valid URL, and may be at most 2,048 characters.
- Only `http` and `https` schemes are accepted.
- `expires_at` is optional, must be a valid RFC 3339 timestamp, and must be in the future.
- `custom_code` is optional. It must be 3-16 letters, numbers, hyphens, or underscores, with a letter or number at each end. It must be unique; a collision returns `409 Conflict`.
- `tags` is optional, supports up to 10 values of at most 32 letters, numbers, hyphens, or underscores, and requires authentication.
- If the request has a valid authenticated session, the link is owned by that account; otherwise it is created anonymously.
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

### Authentication

Register and sign in with `POST /api/auth/register` and `POST /api/auth/login`. Registration accepts `name`, `email`, and `password`; login accepts `email` and `password`. Both endpoints set an HTTP-only `token` session cookie on success and are rate-limited (10 requests per minute by default).

`GET /api/auth/me` returns the current user and requires the session cookie. `POST /api/auth/logout` clears the cookie. Browser clients must send requests with credentials enabled and use the exact frontend origin configured by `WEB_CORS_ORIGIN`.

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

Redirects set a first-party, HTTP-only visitor cookie when one is not already present. The database stores only an HMAC hash of its random token, the referrer hostname (not its path or query), a device category, and an optional country code. The request IP is used transiently for local GeoIP lookup and is not included in request logs or stored in analytics.

### `PUT /api/links/:code`

Update the destination URL, expiration, and tags for a link owned by the authenticated user. Requires the `token` session cookie. Send `expires_at: null` to remove its expiration and `tags: []` to clear its tags. A link owned by another user returns `404 Not Found`.

### `DELETE /api/links/:code`

Delete a link owned by the authenticated user. Its analytics events and tags are removed by database cascades. Returns `204 No Content`; a link owned by another user returns `404 Not Found`.

### `GET /api/me/links`

Lists links owned by the authenticated user. Requires the `token` session cookie. Supports `page` (default `1`), `size` (default `20`, maximum `100`), `q` (matches code or destination URL), `status` (`active` or `expired`), and `tag` (exact match, case-insensitive).

### `GET /api/me/links/:code/analytics?days=30`

Returns per-link click totals, daily click counts, estimated unique visitors, and top referrer, country, and device breakdowns. Requires authentication and link ownership; a link owned by another user is returned as `404 Not Found`. `days` may be `7`, `30`, or `90` and defaults to `30`.

The analytics cookie is browser-based, so unique visitors are estimates rather than verified people. Country breakdowns require `GEOIP_DB_PATH`; without it, `location_available` is false and the countries list is empty. The dashboard can export the selected range and its breakdowns as CSV.

The dashboard also provides QR-code SVG downloads, link editing, custom aliases, tag assignment, server-side search, status/tag filters, and pagination.

### CORS

The API allows credentialed browser requests from the single configured `WEB_CORS_ORIGIN` and supports `GET`, `POST`, `PUT`, `DELETE`, and `OPTIONS`. Allowed request headers are `Origin`, `Content-Type`, and `Accept`. `PUT` is required for dashboard edits; `DELETE` is required for link removal.

### `GET /api/health`

Returns a simple response for checking that the HTTP service is reachable:

```json
{ "status": "ok" }
```

This endpoint does not verify PostgreSQL or Redis connectivity.

## Database and Migrations

Migrations live in `db/migrations` and are applied in order by `make migrate-up`:

| Migration | Purpose                                           |
| --------- | ------------------------------------------------- |
| `000001`  | Create short links.                               |
| `000002`  | Create user accounts.                             |
| `000003`  | Associate links with an optional owner account.   |
| `000004`  | Create per-redirect analytics events and indexes. |
| `000005`  | Create per-link tags and indexes.                 |

The `links` table has the following fields:

| Column         | Type          | Notes                                                              |
| -------------- | ------------- | ------------------------------------------------------------------ |
| `id`           | `BIGSERIAL`   | Primary key.                                                       |
| `code`         | `VARCHAR(16)` | Unique short code; generated codes are currently seven characters. |
| `original_url` | `TEXT`        | Destination URL.                                                   |
| `user_id`      | `BIGINT`      | Nullable owner account; deleting the user deletes owned links.     |
| `clicks`       | `BIGINT`      | Non-null; starts at zero.                                          |
| `expires_at`   | `TIMESTAMPTZ` | Nullable expiration timestamp.                                     |
| `created_at`   | `TIMESTAMPTZ` | Defaults to the database current time.                             |
| `updated_at`   | `TIMESTAMPTZ` | Updated by GORM on entity writes.                                  |

`link_analytics_events` stores each tracked redirect with a link foreign key, an HMAC visitor identifier, a referrer hostname, device category, optional country code, and occurrence time. Raw visitor tokens, IP addresses, and user-agent strings are not stored. `link_tags` stores up to 10 normalized tags per link (32 characters maximum per tag), with a unique link/name pair. Tags cascade when a link is deleted. Deleting a link also cascades to its analytics events.

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

The Go test package is an integration suite and connects to a separate database named `<DB_NAME>_test`. It truncates link, user, and analytics-event tables between tests. Never point the test database configuration at production data.

Create the test database and apply its migrations:

```bash
make test-setup
```

Run the tests:

```bash
make test
```

## Middleware, Errors, and Operational Notes

- Request logging records the HTTP method, path, status, and duration as JSON; it omits client IP addresses.
- Fiber's recovery middleware is enabled.
- CORS allows credentialed `GET`, `POST`, `PUT`, `DELETE`, and `OPTIONS` requests from `WEB_CORS_ORIGIN`, and allows the `Origin`, `Content-Type`, and `Accept` headers.
- CORS is a browser policy, not authentication. Authenticated link management and analytics require a valid session cookie.
- The create rate limiter is process-local; multiple backend replicas do not share limiter state.
- Internal failures are logged. The HTTP error handler returns JSON with an `errors` field.
- Link analytics are scoped to the authenticated owner. Deleting a link also deletes its analytics events.

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
