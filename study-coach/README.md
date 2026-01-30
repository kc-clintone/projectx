# Study Coach — AI Agents Hackathon Project

Problem statement
-----------------
Study Coach helps students improve learning efficiency by timing practice tasks, measuring speed and correctness, and generating personalized study plans that highlight weak topics and adjust future task timers.

SDG alignment
-------------
- SDG 4: Quality Education — improves equitable access to personalized learning and study support.

Architecture & design
---------------------
- Language: Go for the HTTP backend and local storage.
- UI: lightweight single-page UI served from `/ui/` (vanilla JS, no build step).
- Components:
  - API server (`main.go`, `handlers.go`) — endpoints for session creation, submission, and profile retrieval.
  - Logic (`logic.go`) — complexity inference, timer estimation, topic extraction (NLP heuristics), study plan generation, badge catalog.
  - Storage (`storage.go`) — local JSON file persistence (`./data`).
  - UI (`ui/index.html`) — modern, responsive demo UI with badge visuals and gamification.

AI usage
--------
- Lightweight NLP heuristics: tokenization, stopword removal, simple stemming to extract topics from task prompts.
- Rules-based agent: combines rules, stored performance data, and heuristics to produce recommendations and timers. No external LLM is required; the design allows swapping in LLMs later for richer analysis.

Trade-offs & limitations
------------------------
- Local persistence: student data stored as JSON; suitable for hackathon/demo but should be migrated to a proper DB for production.
- Naive NLP: the topic extraction uses simple heuristics (stopwords + suffix stripping), not full linguistic models — acceptable for demo but limited in accuracy.
- Gamification: achievements and badges are simple rules; further UX design and balance required for real adoption.

Run & demo
----------
1. Install dependencies: `go mod tidy`.
2. Run: `go run ./...`.
3. Open demo UI: `http://localhost:8080/ui/index.html`.

Submission checklist
--------------------
- Agent name and description: Study Coach — a timed practice and study-plan agent.
- Problem solved and target user: helps students and educators create paced practice sessions and personalized study plans.
- Input/output schema: documented in `model.go` (Session, Task, Submission, StudyPlan, ProfileResponse).
- Working demo: UI at `/ui/index.html`.
- Deployment/usage instructions: locally via `go run`, extend for deploy.
