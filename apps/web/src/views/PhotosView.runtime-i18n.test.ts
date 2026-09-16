import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'

const source = readFileSync(fileURLToPath(new URL('./PhotosView.vue', import.meta.url)), 'utf8')

describe('PhotosView runtime API error localization wiring', () => {
  it('uses locale-reactive shared API error copy for formatApiError call sites', () => {
    expect(source).toContain('const copy = computed(() => photosRuntimeCopy(locale.value))')
    expect(source).toContain('formatApiError(e, copy.value.loadPhotosFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, copy.value.loadMoreFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, copy.value.createAlbumFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, copy.value.renameFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, copy.value.deleteFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, copy.value.favoriteFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, copy.value.viewFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, copy.value.downloadFailed, copy.value.apiError)')
  })

  it('does not retain two-arg formatApiError calls that can hit English defaults', () => {
    expect(source).not.toMatch(/formatApiError\(\s*e,\s*copy\.value\.[^,)]+\)/)
  })
})
