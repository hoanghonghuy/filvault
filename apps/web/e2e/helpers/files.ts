import { expect, type Page } from '@playwright/test'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { e2eFolderName, e2eImageFileName, e2eTextFileName } from './env'

const fixturesDir = path.join(path.dirname(fileURLToPath(import.meta.url)), '../fixtures')

export const textFixturePath = path.join(fixturesDir, 'sample.txt')
export const imageFixturePath = path.join(fixturesDir, 'sample.png')

export async function navigateToFiles(page: Page): Promise<void> {
  await page.getByRole('link', { name: /^Files$/ }).click()
  await expect(page).toHaveURL(/\/files/)
}

export async function createFolder(page: Page, folderName: string): Promise<void> {
  const viewport = page.viewportSize()
  const isMobile = (viewport?.width ?? 0) < 768

  if (isMobile) {
    await page.getByRole('button', { name: /^Upload file$/ }).click()
    await page.getByRole('button', { name: /^New folder$/ }).click()
  } else {
    await page.getByRole('button', { name: /^New folder$/ }).click()
  }

  const dialog = page.getByRole('dialog')
  await expect(dialog).toBeVisible()
  await dialog.locator('input[type="text"]').fill(folderName)
  await dialog.getByRole('button', { name: /^Create folder$/ }).click()
  await expect(page.getByRole('button', { name: folderName })).toBeVisible()
}

export async function uploadFile(page: Page, fixturePath: string, targetName: string): Promise<void> {
  const viewport = page.viewportSize()
  const isMobile = (viewport?.width ?? 0) < 768

  if (isMobile) {
    await page.getByRole('button', { name: /^Upload file$/ }).click()
    await page.getByRole('button', { name: /^Upload$/ }).click()
  }

  await page.locator('input[type="file"]:not([webkitdirectory])').setInputFiles(fixturePath)
  await expect(page.getByRole('button', { name: targetName })).toBeVisible({ timeout: 30_000 })
}

export async function openFileActions(page: Page, fileName: string): Promise<void> {
  await page.getByRole('button', { name: fileName, exact: true }).click()
  await expect(page.getByRole('dialog')).toBeVisible()
}

export async function previewFile(page: Page, fileName: string): Promise<void> {
  await openFileActions(page, fileName)
  await page.getByRole('button', { name: /^Preview$/ }).click()
  const previewDialog = page.getByRole('dialog', { name: fileName })
  await expect(previewDialog).toBeVisible()
  await previewDialog.getByRole('button', { name: /^Close preview$/ }).click()
  await expect(previewDialog).toBeHidden()
}

export async function downloadFile(page: Page, fileName: string): Promise<void> {
  await openFileActions(page, fileName)
  const popupPromise = page.waitForEvent('popup')
  await page.getByRole('button', { name: /^Download$/ }).click()
  const popup = await popupPromise
  await popup.waitForLoadState('domcontentloaded')
  await popup.close()
}

export async function moveFileToTrash(page: Page, fileName: string): Promise<void> {
  await openFileActions(page, fileName)
  await page.getByRole('button', { name: /^Move to trash$/ }).click()
  await page.getByRole('button', { name: /^Move to trash$/ }).last().click()
  await expect(page.getByRole('button', { name: fileName, exact: true })).toHaveCount(0)
}

export async function moveFolderToTrash(page: Page, folderName: string): Promise<void> {
  const folderCard = page.getByRole('button', { name: folderName, exact: true })
  await folderCard.getByRole('button', { name: /^Folder actions$/ }).click()
  await page.getByRole('button', { name: /^Move to trash$/ }).click()
  await page.getByRole('button', { name: /^Move to trash$/ }).last().click()
  await expect(folderCard).toHaveCount(0)
}

export async function purgeFromTrash(page: Page, names: string[]): Promise<void> {
  await page.goto('/trash')

  for (const name of names) {
    const card = page.locator('.trash-item-card').filter({ hasText: name })
    if (await card.count() === 0) {
      continue
    }
    await card.first().click()
    await page.getByRole('button', { name: /^Delete forever$/ }).click()
    await page.getByRole('button', { name: /^Delete forever$/ }).last().click()
    await expect(card).toHaveCount(0)
  }
}

export async function cleanupSmokeArtifacts(page: Page): Promise<void> {
  await page.goto('/files')
  const names = [e2eTextFileName, e2eImageFileName, e2eFolderName]

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
