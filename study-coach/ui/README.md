This UI is a demo for the Study Coach backend.

Open it at: http://localhost:8080/ui/index.html

Features:
- Create session from tasks (one per line)
- Start per-task timers and double-click to mark a task complete
- Mark correctness manually and submit results
- Student profile panel with points, streak, badges and weekly recap
- Visual badge assets and toast animations for new achievements

Usage:
1. Enter a Student ID (e.g. stu1) and tasks (one per line).
2. Click Create Session, then Start Timers.
3. For each task you can double-click to mark completion or wait for the timer.
4. Tick correct/incorrect checkboxes manually, then Submit Results.
5. After submission the Study Plan is shown and the profile panel updates with gamified elements.

Notes:
- Badge images live under `ui/assets/badges/*.svg` and are referenced by badge id.
- This UI is intentionally lightweight and uses vanilla JS to avoid large toolchain requirements.
