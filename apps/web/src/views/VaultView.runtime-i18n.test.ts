import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'

const source = readFileSync(fileURLToPath(new URL('./VaultView.vue', import.meta.url)), 'utf8')

describe('VaultView runtime API error localization wiring', () => {
  it('uses locale-reactive shared API error copy for formatApiError call sites', () => {
    expect(source).toContain('const copy = computed(() => vaultRuntimeCopy(locale.value))')
    expect(source).toContain('formatApiError(e, t.value.vaultSetupFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, t.value.vaultUnlockFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, t.value.vaultChangePinFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, t.value.vaultResetFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, t.value.vaultDownloadFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, t.value.vaultPreviewFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, t.value.vaultMoveOutFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, t.value.vaultDeleteFailed, copy.value.apiError)')
  })

  it('does not retain two-arg formatApiError calls that can hit English defaults', () => {
    expect(source).not.toMatch(/formatApiError\(\s*e,\s*t\.value\.[^,)]+\)/)
  })
})
