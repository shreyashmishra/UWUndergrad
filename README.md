# PlanAhead

PlanAhead is a University of Waterloo degree planner with a Next.js frontend, a Go GraphQL API, and a local MySQL database.

## What it does now

- fixes the app to a single university: `University of Waterloo`
- syncs the official Waterloo undergraduate major list from the public academic calendar
- loads official requirement sections for a selected program
- shows required courses, choice groups, and text-only requirement rules
- stores student progress locally in MySQL
- runs on the standard local ports:
  - frontend: `3000`
  - backend: `8000`

## Stack

- frontend: Next.js App Router, TypeScript, Tailwind CSS
- backend: Go, `chi`, `graph-gophers/graphql-go`
- database: MySQL
- local containers: Docker Compose

## Local run

Start MySQL first, then run the backend and frontend in separate terminals.

```bash
mysql -h 127.0.0.1 -P 3306 -u root -e 'CREATE DATABASE IF NOT EXISTS degree_tracker;'
```

```bash
cd backend/planner-api
go run ./cmd/server
```

```bash
cd frontend
npm install
npm run dev
```

Open `http://localhost:3000`.

## Docker

Run the full stack with Docker:

```bash
docker compose up --build
```

That starts:

- MySQL on `3306`
- Go API on `8000`
- Next.js app on `3000`

## Environment defaults

See [`.env.example`](./.env.example).

Current local defaults:

- `NEXT_PUBLIC_GRAPHQL_API_URL=http://localhost:8000/graphql`
- `PLANAHEAD_PORT=8000`
- `PLANAHEAD_ALLOWED_ORIGIN=http://localhost:3000`
- `PLANAHEAD_DB_DSN=root@tcp(127.0.0.1:3306)/degree_tracker?parseTime=true`

## Notes

- The backend caches expanded Waterloo program definitions in memory so roadmap refreshes stay fast after the first load.
- Official prerequisite, corequisite, and antirequisite text is pulled from Waterloo course data and shown in the roadmap.
- Auth0 is deferred for now; the app still uses a local mock user.
