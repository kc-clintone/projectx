# Study Coach — AI Agents Hackathon Submission

Agent name and description
--------------------------
Study Coach — a timed practice and personalized study-plan agent. It accepts a student's grade/level, a subject and a list of tasks (questions/assignments), estimates timers per task based on complexity, records actual completion times and correctness, and generates an adaptive study plan that adjusts future timers and recommendations.

Problem solved and target user
------------------------------
Target users: students (K-12, learners preparing for exams) and educators who want a lightweight tool to pace practice sessions, measure speed and accuracy, and produce actionable study plans.

SDG alignment
-------------
- SDG 4: Quality Education — personalized pacing and study recommendations help improve learning outcomes and access to effective self-study tools.

Input / Output schema
---------------------
The canonical request/response types are defined in `backend/model/model.go`. Key types:
- CreateSessionRequest: { student_id, student_level, subject, tasks[] }
- Task: { id, prompt, complexity }
- Submission: { session_id, results[] { actual_seconds, correct } }
- StudyPlan: { student_id, subject, created_at, focus: map[topic]recommendation, next_timers: [] }

Working demo
------------
A lightweight demo UI is served from the backend under the `/ui/` path and the demo entry is `/ui/index.html`. Start the backend and open:

http://localhost:8080/ui/index.html

System architecture
-------------------
- Backend (Go): HTTP API, core logic, local JSON storage (./backend/*)
  - Handlers: request validation, session creation, submission processing, profile/plan retrieval
  - Logic: complexity inference, estimated timers, study plan generation, achievements/badges
  - Storage: local JSON files under `backend/data` (easy to replace with a DB later)
  - AI integration: optional Gemini-compatible helper for short summaries (gated by env var)
- Frontend: static single-page UI served from `/ui/` (demo assets included)

AI usage
--------
- Core behavior uses deterministic heuristics and simple NLP-style topic extraction.
- Optional LLM enrichment: a Gemini-compatible endpoint can produce concise summaries or extended recommendations when enabled (GEMINI_ENABLED=1).

Trade-offs and limitations
--------------------------
- Local file storage is used for simplicity (not production-ready).
- Topic extraction and complexity inference are heuristic-based (acceptable for a hackathon demo but limited precision).
- Gemini API usage is optional and gated behind environment flags to avoid accidental external calls.

Documentation and source layout
-------------------------------
- Backend code: `backend/`
- Models: `backend/model/` (JSON schema and Go structs)
- Frontend UI: `frontend/` and `frontend/ui/`
- Developer instructions and run steps: see `CONTRIBUTING.md` (detailed dev/build/run and tests).

Submission checklist
--------------------
- Agent name and description: included above.
- Problem solved and target user: included above.
- Input/output schema: referenced `backend/model/model.go`.
- Working demo: `/ui/index.html` served by backend.
- Deployment/usage instructions: short run instructions above; more detailed developer steps in `CONTRIBUTING.md`.

License & safety
----------------
This repository is a hackathon/demo project. No production data should be committed. Keep API keys and secrets local (see `.env.local` example).
