#!/bin/bash
# Cleanup Playwright-related files added earlier
rm -f playwright.config.ts
rm -f package.json package-lock.json
rm -f -r node_modules e2e
git add -A
