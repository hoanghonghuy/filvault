import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'

const source = readFileSync(fileURLToPath(new URL('./MediaPickerSheet.vue', import.meta.url)), 'utf8')

describe('MediaPickerSheet runtime API error localization wiring', () => {
  it('uses locale-reactive shared API error copy for formatApiError call sites', () => {
    expect(source).toContain('const copy = computed(() => mediaPickerCopy(locale.value))')
    expect(source).toContain('formatApiError(e, copy.value.loadFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, copy.value.loadMoreFailed, copy.value.apiError)')
  })

  it('does not retain two-arg formatApiError calls that can hit English defaults', () => {
    expect(source).not.toMatch(/formatApiError\(\s*e,\s*copy\.value\.[^,)]+\)/)
  })
})
