import { expect, type Locator, type Page } from '@playwright/test'

export function mobileFab(page: Page): Locator {
  return page.locator('button.fab')
}

export function topDialog(page: Page): Locator {
  return page.getByRole('dialog').last()
}

export async function clickActionSheetItem(page: Page, label: RegExp): Promise<void> {
  const sheet = topDialog(page)
  await expect(sheet).toBeVisible()
  await sheet.getByRole('button', { name: label }).click()
}

export async function confirmDialog(page: Page, label: RegExp): Promise<void> {
  const dialog = page.getByRole('dialog').filter({ has: page.getByRole('button', { name: label }) })
  await expect(dialog).toBeVisible()
  await dialog.getByRole('button', { name: label }).click()
}

export async function closePreviewDialog(page: Page, fileName: string): Promise<void> {
  const previewDialog = page.getByRole('dialog', { name: fileName })
  await previewDialog.locator('button.btn.icon-only[title="Close preview"]').click()
  await expect(previewDialog).toBeHidden()
}
