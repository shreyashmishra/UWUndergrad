# PlanAhead Go API

Go GraphQL backend for the Waterloo roadmap planner.

## Responsibilities

- exposes the GraphQL API on `http://localhost:8000/graphql`
- syncs the official University of Waterloo undergraduate major list
- expands a selected Waterloo program into roadmap sections and required courses
- stores course progress and elective selections in local MySQL

## Stack

- `chi` for HTTP routing
- `graph-gophers/graphql-go` for schema-first GraphQL
- `database/sql` with MySQL

## Local run

```bash
mysql -h 127.0.0.1 -P 3306 -u root -e 'CREATE DATABASE IF NOT EXISTS degree_tracker;'
cd backend/planner-api
go run ./cmd/server
```

## Defaults

- port: `8000`
- DSN: `root@tcp(127.0.0.1:3306)/degree_tracker?parseTime=true`
- allowed origin: `http://localhost:3000`
- mock user key: `local-demo-user`
- rate limit: `10` requests per `1m` per client IP
- rate limit cleanup window: `5m`

## Docker

From the repo root:

```bash
docker compose up --build backend
```
