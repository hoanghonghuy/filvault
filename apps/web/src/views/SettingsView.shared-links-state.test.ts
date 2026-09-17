import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'

const view = readFileSync(new URL('./SettingsView.vue', import.meta.url), 'utf8')
const localization = readFileSync(new URL('../lib/settingsLocalization.ts', import.meta.url), 'utf8')

describe('Settings shared-links load state contract', () => {
  it('keeps load failure distinct from authoritative empty success', () => {
    expect(view).toContain("const linksLoadError = ref(false)")
    expect(view).toContain('linksLoadError.value = false')
    expect(view).toContain('linksLoadError.value = true')
    expect(view).not.toMatch(/catch\s*\{[\s\S]*?shareLinks\.value\s*=\s*\[\][\s\S]*?\}/)
  })

  it('renders an accessible localized error and retry before empty state', () => {
    expect(view).toContain('v-else-if="linksLoadError"')
    expect(view).toContain('role="alert"')
    expect(view).toContain('{{ settingsText.sharedLinksLoadFailed }}')
    expect(view).toContain('@click="loadShareLinks"')
    expect(view).toContain('{{ settingsText.retrySharedLinks }}')
    expect(view.indexOf('v-else-if="linksLoadError"')).toBeLessThan(
      view.indexOf('v-else-if="shareLinks.length === 0"'),
    )
  })

  it('ships VI and EN copy for failure and retry', () => {
    expect(localization).toContain('sharedLinksLoadFailed: string')
    expect(localization).toContain('retrySharedLinks: string')
    expect(localization).toContain("sharedLinksLoadFailed: 'Không thể tải các liên kết đã chia sẻ.'")
    expect(localization).toContain("retrySharedLinks: 'Thử lại'")
    expect(localization).toContain("sharedLinksLoadFailed: 'Could not load shared links.'")
    expect(localization).toContain("retrySharedLinks: 'Retry'")
  })
})
