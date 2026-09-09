import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const readSrc = (relativePath: string) =>
  readFileSync(fileURLToPath(new URL(relativePath, import.meta.url)), 'utf-8')

describe('premium presentation contract (no fake entitlements)', () => {
  const settings = readSrc('./SettingsView.vue')
  const themeView = readSrc('./ThemeView.vue')
  const design = readFileSync(fileURLToPath(new URL('../../../../DESIGN.md', import.meta.url)), 'utf-8')

  it('does not show a PRO badge on the settings profile header', () => {
    const profileCard = settings.match(/profile-header-card[\s\S]*?<\/section>/)?.[0] ?? ''
    expect(profileCard).not.toContain('pro-badge')
    expect(profileCard).not.toContain('>PRO<')
  })

  it('does not render crown or premium badges in Theme Center', () => {
    expect(themeView).not.toContain('isPro')
    expect(themeView).not.toContain('swatch-pro-badge')
    expect(themeView).not.toContain('seasonal-crown-badge')
    expect(themeView).not.toContain('name="crown"')
  })

  it('records the no-premium policy in DESIGN.md', () => {
    expect(design).toContain('## 11. Entitlements & premium presentation (current policy)')
    expect(design).toContain('Chưa có mô hình premium / subscription')
    expect(design).toContain('**Không** hiển thị badge PRO')
  })
})
