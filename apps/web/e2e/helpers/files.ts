import { expect, type Page } from '@playwright/test'
import { readFileSync } from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { e2eFolderName, e2eImageFileName, e2eTextFileName } from './env'
import {
  clickActionSheetItem,
  closePreviewDialog,
  confirmDialog,
  mobileFab,
  topDialog,
} from './ui'

const fixturesDir = path.join(path.dirname(fileURLToPath(import.meta.url)), '../fixtures')

const textFixtureBuffer = readFileSync(path.join(fixturesDir, 'sample.txt'))
const imageFixtureBuffer = readFileSync(path.join(fixturesDir, 'sample.png'))

export async function navigateToFiles(page: Page): Promise<void> {
  await page.getByRole('link', { name: /^Files$/ }).click()
  await expect(page).toHaveURL(/\/files/)
}

export async function createFolder(page: Page, folderName: string): Promise<void> {
  const viewport = page.viewportSize()
  const isMobile = (viewport?.width ?? 0) < 768

  if (isMobile) {
    await mobileFab(page).click()
    await clickActionSheetItem(page, /^New folder$/)
  } else {
    await page.locator('.files-page .desktop-only').getByRole('button', { name: /^New folder$/ }).click()
  }

  const dialog = topDialog(page)
  await expect(dialog).toBeVisible()
  await dialog.locator('input[type="text"]').fill(folderName)
  await dialog.getByRole('button', { name: /^Create folder$/ }).click()
  await expect(page.getByRole('button', { name: folderName, exact: true })).toBeVisible()
}

export async function uploadFile(page: Page, kind: 'text' | 'image', targetName: string): Promise<void> {
  const viewport = page.viewportSize()
  const isMobile = (viewport?.width ?? 0) < 768

  if (isMobile) {
    await mobileFab(page).click()
    await clickActionSheetItem(page, /^Upload$/)
  }

  const payload =
    kind === 'text'
      ? { name: targetName, mimeType: 'text/plain', buffer: textFixtureBuffer }
      : { name: targetName, mimeType: 'image/png', buffer: imageFixtureBuffer }

  await page.locator('input[type="file"]:not([webkitdirectory])').setInputFiles(payload)
  await expect(page.getByRole('button', { name: targetName, exact: true })).toBeVisible({ timeout: 60_000 })
}

export async function openFileActions(page: Page, fileName: string): Promise<void> {
  const fileCard = page.getByRole('button', { name: fileName, exact: true })
  await expect(fileCard).toBeVisible()
  await fileCard.getByRole('button', { name: /^File actions$/ }).click()
  await expect(topDialog(page)).toBeVisible()
}

export async function previewFile(page: Page, fileName: string): Promise<void> {
  await openFileActions(page, fileName)
  await clickActionSheetItem(page, /^Preview$/)
  const previewDialog = page.getByRole('dialog', { name: fileName })
  await expect(previewDialog).toBeVisible()
  await closePreviewDialog(page, fileName)
}

export async function downloadFile(page: Page, fileName: string): Promise<void> {
  await openFileActions(page, fileName)
  const popupPromise = page.waitForEvent('popup')
  await clickActionSheetItem(page, /^Download$/)
  const popup = await popupPromise
  await popup.waitForLoadState('domcontentloaded')
  await popup.close()
}

export async function moveFileToTrash(page: Page, fileName: string): Promise<void> {
  await openFileActions(page, fileName)
  await clickActionSheetItem(page, /^Move to trash$/)
  await confirmDialog(page, /^Move to trash$/)
  await expect(page.getByRole('button', { name: fileName, exact: true })).toHaveCount(0)
}

export async function moveFolderToTrash(page: Page, folderName: string): Promise<void> {
  const folderCard = page.getByRole('button', { name: folderName, exact: true })
  await folderCard.getByRole('button', { name: /^Folder actions$/ }).click()
  await clickActionSheetItem(page, /^Move to trash$/)
  await confirmDialog(page, /^Move to trash$/)
  await expect(folderCard).toHaveCount(0)
}

export async function purgeFromTrash(page: Page, names: string[]): Promise<void> {
  await page.goto('/trash')

  for (const name of names) {
    const card = page.locator('.trash-item-card').filter({ hasText: name })
    if (await card.count() === 0) {
      continue
    }
    await card.first().getByTitle('Delete forever').click()
    await confirmDialog(page, /^Delete forever$/)
    await expect(card).toHaveCount(0)
  }
}

export async function cleanupSmokeArtifacts(page: Page): Promise<void> {
  const names = [e2eTextFileName, e2eImageFileName, e2eFolderName]

  await purgeFromTrash(page, names)

  await page.goto('/files')
  for (const name of names) {
    const item = page.getByRole('button', { name, exact: true })
    if (await item.count() === 0) {
      continue
    }

    if (name === e2eFolderName) {
      await moveFolderToTrash(page, name)
    } else {
      await moveFileToTrash(page, name)
    }
  }

  await purgeFromTrash(page, names)
}
