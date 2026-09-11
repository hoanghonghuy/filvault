import { describe, expect, it } from 'vitest'

import {
  formatSettingsDate,
  formatSettingsRelativeTime,
  getSettingsText,
  settingsActivityLabel,
} from './settingsLocalization'

describe('settings localization', () => {
  it('localizes destructive share copy in Vietnamese and English', () => {
    expect(getSettingsText('vi').revokeConfirmTitle).toBe('Thu hồi liên kết?')
    expect(getSettingsText('vi').revokeConfirmMessage('photo.jpg')).toContain('photo.jpg')
    expect(getSettingsText('vi').revokeConfirmMessage('photo.jpg')).toContain('không còn được chia sẻ công khai')

    expect(getSettingsText('en').revokeConfirmTitle).toBe('Revoke link?')
    expect(getSettingsText('en').revokeConfirmMessage('photo.jpg')).toContain('will no longer be shared publicly')
  })

  it('localizes activity labels', () => {
    expect(settingsActivityLabel('vi', 'file.uploaded')).toBe('Đã tải lên')
    expect(settingsActivityLabel('en', 'file.uploaded')).toBe('Uploaded')
  })

  it('formats relative time with the active locale', () => {
    const now = Date.parse('2026-09-11T03:00:00.000Z')
    const fiveMinutesAgo = '2026-09-11T02:55:00.000Z'

    expect(formatSettingsRelativeTime('en', fiveMinutesAgo, now)).toContain('5')
    expect(formatSettingsRelativeTime('en', fiveMinutesAgo, now).toLowerCase()).toContain('ago')
    expect(formatSettingsRelativeTime('vi', fiveMinutesAgo, now)).toContain('5')
    expect(formatSettingsRelativeTime('vi', fiveMinutesAgo, now).toLowerCase()).toContain('trước')
  })

  it('formats older dates using the active locale', () => {
    const iso = '2026-01-02T10:00:00.000Z'
    expect(formatSettingsDate('en', iso)).not.toBe(formatSettingsDate('vi', iso))
  })
})
