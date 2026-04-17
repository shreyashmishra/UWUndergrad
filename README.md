# PlanAhead

A degree planning tool for University of Waterloo students. Search your program, see the official course roadmap with prerequisites and requirements, and track your progress term by term.

**Stack:** Next.js · TypeScript · Go · GraphQL · MySQL · Docker

---

## Quickstart (Docker)

The fastest way to run everything — no local Go or Node required.

```bash
docker compose up --build
```

Then open [http://localhost:3000](http://localhost:3000).

That starts three containers:

| Service  | Port |
|----------|------|
| Next.js  | 3000 |
| Go API   | 8000 |
| MySQL    | 3306 |

To stop:

```bash
docker compose down
```

---

## Local dev (with hot reload)

Use this if you want fast iteration — the Go server and Next.js both reload on save.

**Step 1 — start MySQL only:**

```bash
docker compose -f docker-compose.dev.yml up -d
```

**Step 2 — start the Go backend** (in one terminal):

```bash
cd backend/planner-api
go run ./cmd/server
```

Or with [air](https://github.com/air-verse/air) for hot reload:

```bash
cd backend/planner-api
air
```

**Step 3 — start the frontend** (in another terminal):

```bash
cd frontend
npm install
npm run dev
```

Open [http://localhost:3000](http://localhost:3000).

---

## Prerequisites

| Tool | Version |
|------|---------|
| Docker + Docker Compose | any recent version |
| Go | 1.22+ (local dev only) |
| Node.js | 18+ (local dev only) |

---

## Environment variables

The app works out of the box with defaults. To override, copy `.env.example` and edit:

```bash
cp .env.example .env
```

| Variable | Default | Description |
|----------|---------|-------------|
| `NEXT_PUBLIC_GRAPHQL_API_URL` | `http://localhost:8000/graphql` | GraphQL endpoint the frontend calls |
| `PLANAHEAD_PORT` | `8000` | Port the Go API listens on |
| `PLANAHEAD_ALLOWED_ORIGIN` | `http://localhost:3000` | CORS allowed origin |
| `PLANAHEAD_DB_DSN` | `root@tcp(127.0.0.1:3306)/degree_tracker?parseTime=true` | MySQL connection string |
| `PLANAHEAD_MOCK_USER_ID` | `local-demo-user` | Mock user ID (auth is not yet wired up) |

---

## Project structure

```
PlanAhead/
├── frontend/               # Next.js App Router (TypeScript, Tailwind)
│   └── src/
│       ├── app/            # Page entry points
│       ├── components/     # Shared UI components
│       └── features/       # Feature modules (roadmap, programs, etc.)
├── backend/
│   └── planner-api/        # Go GraphQL API
│       ├── cmd/server/     # Entry point
│       └── internal/
│           ├── catalog/    # In-memory course/program cache
│           ├── graph/      # GraphQL schema and resolvers
│           ├── model/      # Domain types
│           ├── repository/ # MySQL queries
│           ├── service/    # Business logic
│           └── waterloo/   # Waterloo academic calendar sync
├── docker-compose.yml      # Full stack (prod-like)
└── docker-compose.dev.yml  # MySQL only (for local dev)
```

---

## How it works

1. On startup, the backend syncs the official University of Waterloo undergraduate program list from the public academic calendar.
2. When you select a program in the UI, the backend fetches and expands its full requirement definition — required courses, choice groups, elective rules — and caches it in memory.
3. The frontend renders a term-by-term roadmap with prerequisite and corequisite info pulled from Waterloo course data.
4. Course progress is stored locally in MySQL per mock user (real auth is not yet implemented).

---

## Notes

- Auth is mocked locally via `PLANAHEAD_MOCK_USER_ID`. User accounts and login are not yet implemented.
- Data is sourced from the public Waterloo academic calendar — no API key needed.
- The backend caches expanded program definitions in memory so roadmap loads are fast after the first fetch.
