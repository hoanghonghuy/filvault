import type { User } from '@/api/types'

export type TrashSettings = {
  trashAutoDeleteEnabled: boolean
  trashRetentionDays: number
}

export type PreviewSettings = {
  imageThumbnailsEnabled: boolean
  videoThumbnailsEnabled: boolean
}

export type ExplicitSettingsSection = 'trash' | 'preview'

export function trashSettingsFromUser(user: User): TrashSettings {
  return {
    trashAutoDeleteEnabled: user.trashAutoDeleteEnabled,
    trashRetentionDays: user.trashRetentionDays,
  }
}

export function previewSettingsFromUser(user: User): PreviewSettings {
  return {
    imageThumbnailsEnabled: user.imageThumbnailsEnabled,
    videoThumbnailsEnabled: user.videoThumbnailsEnabled,
  }
}

export function isTrashSettingsDirty(current: TrashSettings, saved: TrashSettings): boolean {
  return (
    current.trashAutoDeleteEnabled !== saved.trashAutoDeleteEnabled ||
    current.trashRetentionDays !== saved.trashRetentionDays
  )
}

export function isPreviewSettingsDirty(current: PreviewSettings, saved: PreviewSettings): boolean {
  return (
    current.imageThumbnailsEnabled !== saved.imageThumbnailsEnabled ||
    current.videoThumbnailsEnabled !== saved.videoThumbnailsEnabled
  )
}

export function hasUnsavedExplicitSettings(
  trashCurrent: TrashSettings,
  trashSaved: TrashSettings,
  previewCurrent: PreviewSettings,
  previewSaved: PreviewSettings,
): boolean {
  return isTrashSettingsDirty(trashCurrent, trashSaved) || isPreviewSettingsDirty(previewCurrent, previewSaved)
}

export function buildTrashPatchBody(current: TrashSettings): Pick<User, 'trashAutoDeleteEnabled' | 'trashRetentionDays'> {
  return {
    trashAutoDeleteEnabled: current.trashAutoDeleteEnabled,
    trashRetentionDays: Number(current.trashRetentionDays),
  }
}

export function buildPreviewPatchBody(
  current: PreviewSettings,
): Pick<User, 'imageThumbnailsEnabled' | 'videoThumbnailsEnabled'> {
  return {
    imageThumbnailsEnabled: current.imageThumbnailsEnabled,
    videoThumbnailsEnabled: current.videoThumbnailsEnabled,
  }
}

/** Keep unsaved section values in auth state while applying the server response for the saved section. */
export function mergeUserAfterSectionSave(
  updated: User,
  section: ExplicitSettingsSection,
  localTrash: TrashSettings,
  localPreview: PreviewSettings,
): User {
  if (section === 'trash') {
    return {
      ...updated,
      imageThumbnailsEnabled: localPreview.imageThumbnailsEnabled,
      videoThumbnailsEnabled: localPreview.videoThumbnailsEnabled,
    }
  }
  return {
    ...updated,
    trashAutoDeleteEnabled: localTrash.trashAutoDeleteEnabled,
    trashRetentionDays: localTrash.trashRetentionDays,
  }
}
