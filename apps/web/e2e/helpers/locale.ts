import type { Page } from '@playwright/test'

export async function useEnglishLocale(page: Page): Promise<void> {
  await page.addInitScript(() => {
    window.localStorage.setItem('filvault.locale', 'en')
  })
}
