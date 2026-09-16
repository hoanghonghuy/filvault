import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'

const source = readFileSync(fileURLToPath(new URL('./SharedWithMeView.vue', import.meta.url)), 'utf8')

describe('SharedWithMeView runtime API error localization wiring', () => {
  it('uses locale-reactive shared API error copy for formatApiError call sites', () => {
    expect(source).toContain('const copy = computed(() => sharedCopy(locale.value))')
    expect(source).toContain('formatApiError(e, copy.value.loadIncomingFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, copy.value.loadLinksFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, copy.value.openFolderFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, copy.value.downloadFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, copy.value.revokeFailed, copy.value.apiError)')
  })

  it('does not retain two-arg formatApiError calls that can hit English defaults', () => {
    expect(source).not.toMatch(/formatApiError\(\s*e,\s*copy\.value\.[^,)]+\)/)
  })
})
