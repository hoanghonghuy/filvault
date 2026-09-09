import { expect, test } from '@playwright/test'
import { e2eEmail } from './helpers/env'
import { useEnglishLocale } from './helpers/locale'

test('shows a login error for invalid credentials', async ({ page }) => {
  await useEnglishLocale(page)
  await page.goto('/login')
  await page.locator('#login-email').fill(e2eEmail)
  await page.locator('#login-password').fill('definitely-wrong-password')
  await page.getByRole('button', { name: /^Sign in$/ }).click()

  const alert = page.locator('#auth-error')
  await expect(alert).toBeVisible()
  await expect(alert).toHaveText('Incorrect email or password.')
  await expect(page).toHaveURL(/\/login/)
})
