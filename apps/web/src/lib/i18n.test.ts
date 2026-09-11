import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const readSrc = (relativePath: string) =>
  readFileSync(fileURLToPath(new URL(relativePath, import.meta.url)), 'utf-8')

describe('i18n contract', () => {
  const i18n = readSrc('./i18n.ts')
  const shell = readSrc('../components/AppShell.vue')
  const settings = readSrc('../views/SettingsView.vue')
  const settingsLocalization = readSrc('./settingsLocalization.ts')

  it('provides Vietnamese and English dictionaries', () => {
    expect(i18n).toContain('vi:')
    expect(i18n).toContain('en:')
    expect(i18n).toMatch(/function useI18n/)
    expect(i18n).toContain('filvault.locale')
    expect(i18n).toMatch(/localStorage\.getItem/)
  })

  it('keeps logout confirmation copy in both locale dictionaries', () => {
    expect(i18n.match(/logoutConfirmTitle:/g)).toHaveLength(2)
    expect(i18n.match(/logoutConfirmMessage:/g)).toHaveLength(2)
    expect(i18n).toContain("logoutConfirmTitle: 'Đăng xuất?'")
    expect(i18n).toContain("logoutConfirmTitle: 'Log out?'")
    expect(i18n).toMatch(/typeof localStorage === 'undefined'/)
  })

  it('uses translated navigation labels in the app shell', () => {
    expect(shell).toMatch(/useI18n/)
    for (const label of ['Overview', 'Files', 'Photos', 'Trash', 'Settings']) {
      expect(shell).not.toContain(`label: '${label}'`)
    }
  })

  it('exposes a localized language control in Settings', () => {
    expect(settings).toMatch(/setLocale\(/)
    expect(settings).toMatch(/filvault.locale/)
    expect(settings).toContain(':aria-label="settingsText.languageGroup"')
    expect(settingsLocalization).toContain("languageGroup: 'Ngôn ngữ'")
    expect(settingsLocalization).toContain("languageGroup: 'Language'")
  })
})
