# go-astro-url-shortener

A URL shortener with a Go backend (clean architecture) and an Astro frontend.

## Features

- Create short links from HTTP/HTTPS URLs using random Base62 codes
- Redirect with HTTP 302 and an atomic hit counter
- Optionally set an expiration date for each link (returns 410 Gone after expiration)
- Choose a custom short alias
- Rate-limit link creation per IP address
- View per-link click trends, estimated unique visitors, referrers, countries, and device categories
- Edit, search, filter, tag, and paginate account-owned links in the dashboard
- Generate downloadable QR codes and export analytics to CSV
- Optionally enable country analytics with a locally supplied GeoIP database

## Tech Stack

- **Backend**: Go, Fiber, GORM, PostgreSQL, Viper, Logrus, golang-migrate
- **Frontend**: Astro (static) + React island, Tailwind CSS
- **Infra**: Docker Compose, nginx

## Project Structure

Complete backend documentation: [backend/docs/README.md](backend/docs/README.md).

```
backend/    # Go, clean architecture (delivery -> usecase -> repository)
frontend/   # Astro
docker-compose.yml
```

## Run with Docker

```bash
cd backend
cp .env.example .env
make app-up
```

- Frontend: http://localhost:4321
- API: http://localhost:3000

Stop the services with `make app-down`.

## Local Development

Prerequisites: Go, Node.js, Docker, [air](https://github.com/air-verse/air), and [golang-migrate](https://github.com/golang-migrate/migrate).

```bash
cd backend
cp .env.example .env
make db-up
make migrate-up
make dev

cd frontend
cp .env.example .env
npm install
npm run dev
```

Run backend tests (using a separate `<DB_NAME>_test` database):

```bash
cd backend
make test-setup
make test
```

Run frontend tests (no database required):

```bash
cd frontend
npm test
```

## API

| Method | Path               | Description                                                                                          |
| ------ | ------------------ | ---------------------------------------------------------------------------------------------------- |
| POST   | `/api/links`       | Create a link; optional `custom_code`, `expires_at`, and owner-only `tags`. |
| GET    | `/api/links/:code` | Get link details and statistics. Returns 200 or 404.                                                 |
| PUT    | `/api/links/:code` | Update the authenticated user's destination, expiry, and tags.                                        |
| GET    | `/api/me/links`    | List/search/filter the authenticated user's links with pagination.                                     |
| GET    | `/api/me/links/:code/analytics?days=30` | Get owner-only analytics for 7, 30, or 90 days. |
| DELETE | `/api/links/:code` | Delete an authenticated user's link.                                                                  |
| GET    | `/:code`           | Redirect to the original URL. Returns 302, 404, or 410.                                              |
| GET    | `/api/health`      | Check whether the service is reachable.                                                              |

## Configuration

See `backend/.env.example` and `frontend/.env.example`. The frontend's `PUBLIC_API_URL` is embedded at build time.
Analytics event retention defaults to 90 days. Country analytics require a MaxMind-compatible `.mmdb` file configured with `GEOIP_DB_PATH`; the database is not included in this repository.
