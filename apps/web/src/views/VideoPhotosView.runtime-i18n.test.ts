import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'

const source = readFileSync(fileURLToPath(new URL('./VideoPhotosView.vue', import.meta.url)), 'utf8')

describe('VideoPhotosView runtime API error localization wiring', () => {
  it('uses locale-reactive shared API error copy for formatApiError call sites', () => {
    expect(source).toContain('const photosCopy = computed(() => photosRuntimeCopy(locale.value))')
    expect(source).toContain('formatApiError(e, errorCopy.value.load, photosCopy.value.apiError)')
    expect(source).toContain('formatApiError(e, errorCopy.value.loadMore, photosCopy.value.apiError)')
    expect(source).toContain('formatApiError(e, errorCopy.value.view, photosCopy.value.apiError)')
    expect(source).toContain('formatApiError(e, errorCopy.value.download, photosCopy.value.apiError)')
    expect(source).toContain('formatApiError(e, errorCopy.value.favorite, photosCopy.value.apiError)')
  })

  it('does not retain two-arg formatApiError calls that can hit English defaults', () => {
    expect(source).not.toMatch(/formatApiError\(\s*e,\s*errorCopy\.value\.[^,)]+\)/)
  })
})
