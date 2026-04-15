# PlanAhead

PlanAhead is a production-style MVP for a university degree requirement tracker. This version starts with the University of Waterloo, exposes a GraphQL API over FastAPI, and ships a Next.js portal that persists student progress in `localStorage` while keeping the backend ready for real authenticated persistence later.

## Architecture

- Frontend: Next.js App Router, TypeScript, Tailwind CSS, local-first persistence via storage services.
- Backend: FastAPI + Strawberry GraphQL with thin resolvers and service/domain modules.
- Persistence: MySQL via SQLAlchemy 2.0, Alembic migrations, seeded Waterloo Computer Science roadmap data.
- Domain split: shared requirement evaluation lives under `backend/app/modules/universities/common`, while Waterloo-specific program definitions and engine hooks live under `backend/app/modules/universities/waterloo`.
- API-first flow: the frontend stores progress locally, sends it to GraphQL for roadmap evaluation, and renders a semester-by-semester plan with prerequisite feedback and remaining requirement counts.

## Folder Structure

```text
.
├── .env.example
├── backend
│   ├── api
│   │   └── index.py
│   ├── alembic
│   │   ├── env.py
│   │   ├── script.py.mako
│   │   └── versions
│   │       └── 20260412_0001_initial_schema.py
│   ├── app
│   │   ├── core
│   │   │   └── config.py
│   │   ├── db
│   │   │   ├── models
│   │   │   ├── repositories
│   │   │   ├── seed.py
│   │   │   └── session.py
│   │   ├── graphql
│   │   │   ├── context.py
│   │   │   ├── mutations
│   │   │   ├── queries
│   │   │   ├── schema
│   │   │   └── types
│   │   ├── main.py
│   │   └── modules
│   │       ├── programs
│   │       ├── students
│   │       └── universities
│   │           ├── common
│   │           └── waterloo
│   ├── requirements.txt
│   ├── tests
│   │   └── test_requirement_engine.py
│   └── vercel.json
├── frontend
│   ├── package.json
│   ├── next.config.ts
│   ├── tailwind.config.ts
│   └── src
│       ├── app
│       ├── components
│       ├── features
│       │   ├── programs
│       │   ├── progress
│       │   ├── roadmap
│       │   └── storage
│       ├── lib
│       │   ├── graphql
│       │   ├── storage
│       │   └── utils
│       └── types
└── README.md
```

## Backend Highlights

- SQLAlchemy models cover `University`, `Program`, `Course`, `Term`, `RequirementGroup`, `ElectiveGroup`, `ProgramPlanTemplate`, `ProgramRequirement`, `PrerequisiteRule`, `Student`, `StudentCourseProgress`, and `ElectiveSelection`.
- Alembic ships with an initial migration in `backend/alembic/versions/20260412_0001_initial_schema.py`.
- GraphQL queries:
  - `availableUniversities`
  - `programsByUniversity`
  - `roadmapByProgram`
  - `studentProgress`
  - `requirementSummary`
- GraphQL mutations:
  - `updateCourseStatus`
  - `selectElective`
  - `clearElectiveSelection`
- Waterloo seed data includes a realistic multi-year Computer Science roadmap, elective groups, prerequisites, and demo student progress.

## Frontend Highlights

- The root page routes directly into the student portal.
- University and program selectors persist through `ProgramSelectionStorageService`.
- Course and elective progress persist through `ProgressStorageService`.
- The roadmap UI renders by year and semester with:
  - required course cards
  - elective choice groups
  - prerequisite warnings
  - progress summary and remaining requirements

## Environment Setup

Create a single repo-root `.env` from `.env.example`.

```bash
cp .env.example .env
```

The backend reads that root `.env` directly. The frontend also reads the root `.env` via `frontend/next.config.ts`, so `NEXT_PUBLIC_*` values stay in one place during local development.

Important variables:

- `NEXT_PUBLIC_APP_URL`
- `NEXT_PUBLIC_GRAPHQL_API_URL`
- `APP_ENV`
- `DATABASE_URL`
- `MYSQL_*`
- placeholder `AUTH0_*` values for future integration

Do not commit the real `.env`.

## Local Setup

### 1. Backend

```bash
python3 -m venv .venv
source .venv/bin/activate
pip install -r backend/requirements.txt
cd backend
alembic upgrade head
python -m app.db.seed
uvicorn app.main:app --reload
```

The GraphQL endpoint will be available at `http://localhost:8000/graphql`.

Notes:

- Do not separately install a newer `pydantic` version on top of `backend/requirements.txt`. The current Strawberry pin in this MVP expects `pydantic<2.11`.
- `alembic upgrade head` and `python -m app.db.seed` require a running MySQL server reachable at the `DATABASE_URL` in your repo-root `.env`.
- If `uvicorn` reports `Address already in use`, another process is already bound to port `8000`. Stop that process or run `uvicorn app.main:app --reload --port 8001`.

### 2. Frontend

In a separate terminal:

```bash
cd frontend
npm install
npm run dev
```

The app will be available at `http://localhost:3000`.

## MySQL Notes

- Create a local MySQL database that matches `MYSQL_DATABASE`.
- Ensure the MySQL user in `DATABASE_URL` can create tables and indexes.
- If you change the schema, add a new Alembic revision rather than editing the initial migration after real usage begins.
- If you see `Can't connect to MySQL server on 'localhost' ([Errno 61] Connection refused)`, MySQL is not running yet or `DATABASE_URL` points at the wrong host/port.

## Seed Data

The seed currently loads:

- University of Waterloo
- Computer Science program
- semester-by-semester roadmap from Year 1 Fall through Year 4 Winter
- core CS, math, stats, breadth, and upper-year elective options
- prerequisite edges for major CS progression points
- a demo local student with mixed completed, in-progress, and planned courses

## Deployment on Vercel

Deploy the frontend and backend as separate Vercel projects.

### Frontend deployment

1. Create a Vercel project with root directory `frontend`.
2. Add `NEXT_PUBLIC_APP_URL` and `NEXT_PUBLIC_GRAPHQL_API_URL`.
3. Build command: default Next.js build.
4. Output: default Next.js output.

### Backend deployment

1. Create a second Vercel project with root directory `backend`.
2. Vercel will use `api/index.py` as the Python entrypoint.
3. `backend/vercel.json` routes all traffic to the FastAPI app.
4. Add environment variables for `APP_ENV`, `DATABASE_URL`, and any future values you need.
5. Point the frontend `NEXT_PUBLIC_GRAPHQL_API_URL` at the deployed backend `/graphql` URL.

## Future Extension Points

- Replace `localStorage` services with authenticated API persistence without changing most UI components.
- Add more Waterloo programs under `backend/app/modules/universities/waterloo/programs`.
- Add new universities by introducing a new package beside `waterloo` and reusing the shared requirement engine.
- Extend prerequisite modeling to anti-requisites, transfer credit, and calendar-year variants.
