#!/usr/bin/env bash
set -euo pipefail
# Run frontend uvu tests from the repo root using npm --prefix so CI and local runs are consistent
npm --prefix frontend/tests test
