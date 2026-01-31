import { test, expect } from '@playwright/test';

test('landing page loads and shows login', async ({ page }) => {
  await page.goto('/index.html');
  await expect(page.locator('text=Welcome to Study Coach')).toBeVisible();
});

// more tests can be added to exercise login, create session, OCR upload can be mocked, and session player
