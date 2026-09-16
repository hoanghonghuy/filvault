import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'

const source = readFileSync(fileURLToPath(new URL('./SettingsView.vue', import.meta.url)), 'utf8')

describe('SettingsView runtime API error localization wiring', () => {
  it('uses locale-reactive shared API error copy for formatApiError call sites', () => {
    expect(source).toContain('const settingsText = computed(() => getSettingsText(locale.value))')
    expect(source).toContain(
      'formatApiError(e, t.value.trashSaveFailed, settingsText.value.apiError)',
    )
    expect(source).toContain(
      'formatApiError(e, t.value.previewSaveFailed, settingsText.value.apiError)',
    )
    expect(source).toContain(
      'formatApiError(e, settingsText.value.revokeFailed, settingsText.value.apiError)',
    )
  })

  it('does not retain two-arg formatApiError calls that can hit English defaults', () => {
    expect(source).not.toMatch(/formatApiError\(\s*e,\s*t\.value\.(?:trashSaveFailed|previewSaveFailed)\)/)
    expect(source).not.toMatch(/formatApiError\(\s*e,\s*settingsText\.value\.revokeFailed\)/)
  })
})
