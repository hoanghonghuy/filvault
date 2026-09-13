import { describe, expect, it } from 'vitest'
import {
  buildPreviewPatchBody,
  buildTrashPatchBody,
  hasUnsavedExplicitSettings,
  isPreviewSettingsDirty,
  isTrashSettingsDirty,
  previewSettingsFromUser,
  trashSettingsFromUser,
} from './settingsSections'
import type { User } from '@/api/types'

const baseUser: User = {
  id: 'user-1',
  email: 'user@example.com',
  displayName: 'User',
  emailVerified: true,
  storageUsed: 0,
  storageQuota: 1024,
  imageThumbnailsEnabled: true,
  videoThumbnailsEnabled: true,
  trashAutoDeleteEnabled: false,
  trashRetentionDays: 30,
  createdAt: '2026-01-01T00:00:00.000Z',
}

describe('settingsSections', () => {
  it('detects dirty state per section independently', () => {
    const savedTrash = trashSettingsFromUser(baseUser)
    const savedPreview = previewSettingsFromUser(baseUser)

    const dirtyTrash = { ...savedTrash, trashRetentionDays: 14 }
    const dirtyPreview = { ...savedPreview, imageThumbnailsEnabled: false }

    expect(isTrashSettingsDirty(dirtyTrash, savedTrash)).toBe(true)
    expect(isPreviewSettingsDirty(savedPreview, savedPreview)).toBe(false)
    expect(isPreviewSettingsDirty(dirtyPreview, savedPreview)).toBe(true)
    expect(isTrashSettingsDirty(savedTrash, savedTrash)).toBe(false)
  })

  it('builds section-scoped PATCH bodies', () => {
    const trash = { trashAutoDeleteEnabled: true, trashRetentionDays: 7 }
    const preview = { imageThumbnailsEnabled: false, videoThumbnailsEnabled: true }

    expect(buildTrashPatchBody(trash)).toEqual({
      trashAutoDeleteEnabled: true,
      trashRetentionDays: 7,
    })
    expect(buildPreviewPatchBody(preview)).toEqual({
      imageThumbnailsEnabled: false,
      videoThumbnailsEnabled: true,
    })
    expect(Object.keys(buildTrashPatchBody(trash))).not.toContain('imageThumbnailsEnabled')
    expect(Object.keys(buildPreviewPatchBody(preview))).not.toContain('trashRetentionDays')
  })

  it('tracks unsaved explicit settings across sections', () => {
    const savedTrash = trashSettingsFromUser(baseUser)
    const savedPreview = previewSettingsFromUser(baseUser)

    expect(
      hasUnsavedExplicitSettings(savedTrash, savedTrash, savedPreview, savedPreview),
    ).toBe(false)

    expect(
      hasUnsavedExplicitSettings(
        { ...savedTrash, trashRetentionDays: 10 },
        savedTrash,
        savedPreview,
        savedPreview,
      ),
    ).toBe(true)

    expect(
      hasUnsavedExplicitSettings(
        savedTrash,
        savedTrash,
        { ...savedPreview, videoThumbnailsEnabled: false },
        savedPreview,
      ),
    ).toBe(true)
  })
})
