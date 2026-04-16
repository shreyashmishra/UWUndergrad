# PlanAhead

PlanAhead is a University of Waterloo degree-planning app with a Next.js frontend, a Go GraphQL backend, and a local MySQL database.

## Stack

- Frontend: Next.js App Router, TypeScript, Tailwind CSS
- Backend: Go, `chi`, `graph-gophers/graphql-go`
- Database: MySQL

## Current scope

- Waterloo program selection
- multiple seeded programs in the Go backend
  - `CS`
  - `MATH`
- roadmap-by-term view
- prerequisite warnings
- elective selection
- course status tracking
- progress summary

## Repo layout

```text
.
├── .env.example
├── backend
│   └── planner-api
│       ├── cmd/server
│       ├── internal
│       ├── migrations
│       └── seeds
└── frontend
    └── src
```

## Run locally

1. Start MySQL.
2. Create the local database:

```bash
mysql -h 127.0.0.1 -P 3306 -u root -e 'CREATE DATABASE IF NOT EXISTS degree_tracker;'
```

3. Start the Go API:

```bash
cd backend/planner-api
go run ./cmd/server
```

4. Start the frontend in another terminal:

```bash
cd frontend
npm install
npm run dev
```

5. Open:

```text
http://localhost:3000
```

## Environment

Copy the example file if you want explicit local config:

```bash
cp .env.example .env
```

The frontend will automatically prefer the Go API on `http://localhost:8080/graphql` during local development.

## Notes

- The Go backend auto-bootstraps the MySQL schema and demo Waterloo data on startup.
- Auth0 is intentionally deferred.
- The old Python backend has been removed from the active codebase.
