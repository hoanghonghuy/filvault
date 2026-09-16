import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'

const source = readFileSync(fileURLToPath(new URL('./OverviewView.vue', import.meta.url)), 'utf8')

describe('OverviewView runtime API error localization wiring', () => {
  it('uses locale-reactive shared API error copy for formatApiError call sites', () => {
    expect(source).toContain('const copy = computed(() => overviewCopy(locale.value))')
    expect(source).toContain(
      'formatApiError(browserResult.reason, copy.value.loadFilesFailed, copy.value.apiError)',
    )
    expect(source).toContain(
      'formatApiError(timelineResult.reason, copy.value.loadPhotosFailed, copy.value.apiError)',
    )
  })

  it('does not retain two-arg formatApiError calls that can hit English defaults', () => {
    expect(source).not.toMatch(/formatApiError\(\s*[^,]+,\s*copy\.value\.[^,)]+\)/)
  })
})
