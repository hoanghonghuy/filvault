import { expect, type Page } from '@playwright/test'
import { e2eEmail, e2ePassword } from './env'
import { useEnglishLocale } from './locale'

export async function login(page: Page): Promise<void> {
  await useEnglishLocale(page)
  await page.goto('/login')
  await page.locator('#login-email').fill(e2eEmail)
  await page.locator('#login-password').fill(e2ePassword)
  await page.getByRole('button', { name: /^Sign in$/ }).click()
  await expect(page).not.toHaveURL(/\/login/)
}

export async function logout(page: Page): Promise<void> {
  await page.goto('/settings')
  await page.getByRole('button', { name: /^Log out$/ }).click()
  await page.getByRole('button', { name: /^Log out$/ }).last().click()
  await expect(page).toHaveURL(/\/login/)
}
