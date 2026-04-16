# PlanAhead Go API

Go GraphQL backend for the PlanAhead roadmap app.

## Uses

- `chi` for HTTP routing
- `graph-gophers/graphql-go` for schema-first GraphQL
- `database/sql` with MySQL

## Features

- lists available universities
- lists programs by university
- evaluates a roadmap by program
- stores course progress and elective selections
- seeds demo Waterloo catalog data into MySQL on startup

## Local run

```bash
mysql -h 127.0.0.1 -P 3306 -u root -e 'CREATE DATABASE IF NOT EXISTS degree_tracker;'
cd backend/planner-api
go run ./cmd/server
```

Defaults:

- port: `8080`
- DSN: `root@tcp(127.0.0.1:3306)/degree_tracker?parseTime=true`
- demo student key: `local-demo-user`
