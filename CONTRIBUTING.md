Developer instructions — Study Coach (local dev)

Prerequisites
-------------
- Go 1.20+
- curl (for quick smoke tests)

Local run
---------
1. From the repository root run:

   cd backend
   go mod tidy
   go run .

2. Open the demo UI in a browser:

   http://localhost:8080/ui/index.html

Environment
-----------
- GEMINI_ENABLED=1 enables external LLM calls. Use with caution.
- GEMINI_API_KEY can be provided in environment or in a local `.env.local` file with the single line `GEMINI_API_KEY=...`.

Testing
-------
Run unit tests in the backend:

cd backend
go test ./... -v

Committing
----------
We keep commits local during development. Use `git add`/`git commit` as usual; do not push until changes are reviewed.

Notes
-----
- The backend stores per-student data under `backend/data` as JSON files for simple persistence; replace with a proper DB in production.
- The code expects module path `github.com/kc-clintone/study-coach` so `go mod` operations resolve local packages correctly.
