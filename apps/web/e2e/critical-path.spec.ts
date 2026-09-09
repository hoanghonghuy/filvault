import { test } from '@playwright/test'
import { login, logout } from './helpers/auth'
import { e2eFolderName, e2eImageFileName, e2eTextFileName } from './helpers/env'
import {
  cleanupSmokeArtifacts,
  createFolder,
  downloadFile,
  moveFileToTrash,
  moveFolderToTrash,
  navigateToFiles,
  previewFile,
  purgeFromTrash,
  uploadFile,
} from './helpers/files'

test.describe('critical path', () => {
  test.beforeEach(async ({ page }) => {
    await login(page)
    await cleanupSmokeArtifacts(page)
  })

  test.afterEach(async ({ page }) => {
    await cleanupSmokeArtifacts(page)
  })

  test('auth → files → folder → upload → preview/download → trash → logout', async ({ page }) => {
    await navigateToFiles(page)
    await createFolder(page, e2eFolderName)

    await uploadFile(page, 'text', e2eTextFileName)
    await uploadFile(page, 'image', e2eImageFileName)

    await previewFile(page, e2eImageFileName)
    await downloadFile(page, e2eTextFileName)

    await moveFileToTrash(page, e2eTextFileName)
    await moveFileToTrash(page, e2eImageFileName)
    await moveFolderToTrash(page, e2eFolderName)

    await purgeFromTrash(page, [e2eTextFileName, e2eImageFileName, e2eFolderName])

    await logout(page)
  })
})
