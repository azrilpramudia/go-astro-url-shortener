# Go Astro URL Shortener

A URL shortener with a Go API and an Astro + React frontend. Authenticated users can manage links and view detailed analytics; visitors can create and use short links without an account.

## Features

- Create short links for HTTP and HTTPS URLs, with optional expiration and custom aliases.
- Redirect with HTTP 302; expired links return 410 Gone.
- Register and sign in to manage owned links.
- Edit destinations, expiration dates, and tags; search, filter, and paginate links.
- View click trends, estimated unique visitors, referrers, countries, and device categories.
- Download QR codes as SVG and export analytics as CSV.
- Optionally enable country analytics using a locally supplied GeoIP database.
- Responsive dashboard layout for mobile, tablet, and desktop screens.

## Tech Stack

- **Backend:** Go, Fiber, GORM, PostgreSQL, Redis, golang-migrate
- **Frontend:** Astro, React, Tailwind CSS
- **Infrastructure:** Docker Compose, nginx

## Project Structure

```text
backend/             Go API, migrations, tests, and backend documentation
frontend/            Astro/React application and frontend documentation
docker-compose.yml   Local infrastructure and full application profile
```

- Backend setup, API, configuration, and database details: [backend/docs/README.md](backend/docs/README.md)
- Frontend setup, configuration, and available scripts: [frontend/README.md](frontend/README.md)

## Run the Full App with Docker

Requires Docker Compose. From the repository root:

```bash
cd backend
cp .env.example .env
```

Set `DB_PASSWORD` and a `JWT_SECRET` of at least 32 characters in `backend/.env`, then run:

```bash
make app-up
```

- Frontend: http://localhost:4321
- API: http://localhost:3000

The Compose profile starts PostgreSQL and Redis, applies migrations, then starts the API and frontend. Stop the services with `make app-down`; this leaves the named database volumes intact.

## Local Development

Prerequisites: Go (version in `backend/go.mod`), Node.js 22.23.4 or newer, Docker, `air` (optional, for Go hot reload), and the `migrate` CLI.

1. Configure and start the backend dependencies:

   ```bash
   cd backend
   cp .env.example .env
   ```

   Set `DB_PASSWORD` and a `JWT_SECRET` of at least 32 characters in `.env`, then:

   ```bash
   make db-up
   make migrate-up
   ```

2. In a backend terminal, start the API:

   ```bash
   cd backend
   make dev
   ```

   Use `make run` instead if `air` is not installed.

3. In a frontend terminal, configure and start the UI:

   ```bash
   cd frontend
   cp .env.example .env
   npm ci
   npm run dev
   ```

   For local development, `PUBLIC_API_URL` should point to `http://localhost:3000`, and backend `WEB_CORS_ORIGIN` should match the frontend origin, normally `http://localhost:4321`.

### Tests

Backend integration tests use a separate `<DB_NAME>_test` database. Create it and apply its migrations before running the suite:

```bash
cd backend
make test-setup
make test
```

Frontend tests do not require a database:

```bash
cd frontend
npm test
npm run check
npm run build
```

## API Overview

Authenticated routes use the session cookie set by login/register. Link ownership is enforced on update, deletion, listing, and analytics.

| Method | Path                                    | Description                                                                             |
| ------ | --------------------------------------- | --------------------------------------------------------------------------------------- |
| POST   | `/api/auth/register`                    | Register an account.                                                                    |
| POST   | `/api/auth/login`                       | Sign in and start a session.                                                            |
| POST   | `/api/auth/logout`                      | End the current session.                                                                |
| GET    | `/api/auth/me`                          | Get the current authenticated user.                                                     |
| POST   | `/api/links`                            | Create a link; supports optional `custom_code`, `expires_at`, and authenticated `tags`. |
| GET    | `/api/links/:code`                      | Get public link details and click count.                                                |
| PUT    | `/api/links/:code`                      | Update the authenticated owner's destination, expiration, and tags.                     |
| DELETE | `/api/links/:code`                      | Delete the authenticated owner's link.                                                  |
| GET    | `/api/me/links`                         | List, search, filter, and paginate the authenticated user's links.                      |
| GET    | `/api/me/links/:code/analytics?days=30` | Get owner-only analytics for 7, 30, or 90 days.                                         |
| GET    | `/:code`                                | Redirect to the destination; returns 302, 404, or 410.                                  |
| GET    | `/api/health`                           | Check whether the HTTP service is reachable.                                            |

See the [backend API reference](backend/docs/README.md#http-api) for request fields, validation, and response details.

## Configuration Notes

- Backend settings are documented in [backend/.env.example](backend/.env.example). `WEB_CORS_ORIGIN` must match the browser-facing frontend origin.
- Frontend `PUBLIC_API_URL` is embedded at build time; set it to the API URL reachable by the user's browser.
- Analytics events are retained for 90 days by default. Country analytics require a MaxMind-compatible `.mmdb` file configured with `GEOIP_DB_PATH`; the database is not included.
