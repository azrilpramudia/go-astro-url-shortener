# Frontend

The frontend is a static Astro site with interactive React components and Tailwind CSS. It communicates with the Go API; it does not persist application data locally.

## Features

- Create short links with an optional expiration date and custom alias.
- Create links while signed out; those links remain in the current page session only.
- Register and sign in to save, edit, and delete owned links.
- Search by short code or destination URL; filter by active/expired status or tag, with pagination.
- Add and edit tags for account-owned links.
- View per-link click trends, estimated unique visitors, referrers, countries, and device categories.
- Choose a 7-, 30-, or 90-day analytics range and export the selected report as CSV.
- Generate and download a QR code as SVG.
- Responsive layout for narrow mobile screens through desktop viewports.

Country analytics are shown only when the backend has a GeoIP database configured. Unique visitor counts are estimates based on a first-party browser cookie.

## Requirements

- Node.js `>=22.23.4` (see `package.json`)
- The Go backend running and its database migrations applied

## Local Development

Start the backend first; see the [backend setup guide](../backend/docs/README.md#quick-start-local-development). Then, from `frontend/`:

```bash
cp .env.example .env
npm ci
npm run dev
```

The development server is available at http://localhost:4321.

### Configuration

`PUBLIC_API_URL` is the API base URL used by the browser:

```dotenv
PUBLIC_API_URL=http://localhost:3000
```

It is declared in `astro.config.mjs`, is public in the built client bundle, and is embedded at build time. Set it to the API address reachable from the user's browser. For local development, the backend's `WEB_CORS_ORIGIN` must match the frontend origin (`http://localhost:4321` by default). API requests include credentials for session-cookie authentication.

For Docker builds, `PUBLIC_API_URL` is passed as a build argument by the Compose configuration. Configure backend environment variables in `backend/.env`; see [backend/.env.example](../backend/.env.example).

## Commands

Run these commands from `frontend/`:

| Command           | Purpose                                       |
| ----------------- | --------------------------------------------- |
| `npm ci`          | Install the locked dependencies.              |
| `npm run dev`     | Start the Astro development server.           |
| `npm test`        | Run the Vitest component and utility tests.   |
| `npm run check`   | Run Astro and TypeScript diagnostics.         |
| `npm run build`   | Create the static production site in `dist/`. |
| `npm run preview` | Preview the production build locally.         |

## Project Structure

```text
src/
  components/   React forms, link manager, link list, and analytics UI
  lib/          API client, clipboard helpers, and CSV generation
  pages/        Dashboard, sign-in, and registration pages
  styles/       Global Tailwind styles
  types/        API response and frontend domain types
```

The frontend uses the backend endpoints documented in the [backend HTTP API reference](../backend/docs/README.md#http-api). For full project setup, see the [main README](../README.md).
