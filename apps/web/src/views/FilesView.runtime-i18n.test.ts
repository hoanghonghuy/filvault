import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'

const source = readFileSync(fileURLToPath(new URL('./FilesView.vue', import.meta.url)), 'utf8')

describe('FilesView runtime API error localization wiring', () => {
  it('uses locale-reactive shared API error copy for formatApiError call sites', () => {
    expect(source).toContain('const copy = computed(() => filesRuntimeCopy(locale.value))')
    expect(source).toContain('formatApiError(e, t.value.filesLoadFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, t.value.filesSearchFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, t.value.filesPreviewFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, t.value.folderCreateFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, t.value.fileDownloadFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, operationsCopy.value.renameFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, operationsCopy.value.moveFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, operationsCopy.value.trashFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, t.value.favoriteBatchFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, t.value.favoritesLoadFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, t.value.favoriteUpdateFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, vaultCopy.value.moveInFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, t.value.shareLinkCreateFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, t.value.shareLinkRevokeFailed, copy.value.apiError)')
  })

  it('does not retain two-arg formatApiError calls that can hit English defaults', () => {
    expect(source).not.toMatch(/formatApiError\(\s*e,\s*t\.value\.[^,)]+\)/)
    expect(source).not.toMatch(/formatApiError\(\s*e,\s*operationsCopy\.value\.[^,)]+\)/)
    expect(source).not.toMatch(/formatApiError\(\s*e,\s*vaultCopy\.value\.[^,)]+\)/)
  })
})
