# study-coach

A simple Go-based hackathon project: an AI Study Coach that times tasks, records student performance locally, and generates study plans.

Run:

  go run ./...

API:

- POST /api/v1/session
  - body: CreateSessionRequest
  - returns: Session with estimated timers

- POST /api/v1/submit_results
  - body: Submission
  - returns: StudyPlan

- GET /api/v1/study_plan/{student_id}
  - returns: last saved StudyPlan

Notes:
- Storage is local JSON files under ./data
- Subject whitelist is minimal; extend as needed
