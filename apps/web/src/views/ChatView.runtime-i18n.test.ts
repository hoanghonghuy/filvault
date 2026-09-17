import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'

const source = readFileSync(fileURLToPath(new URL('./ChatView.vue', import.meta.url)), 'utf8')

describe('ChatView runtime loading localization wiring', () => {
  it('uses locale-reactive fallbacks and shared API error copy for runtime errors', () => {
    expect(source).toContain('const copy = computed(() => chatRuntimeCopy(locale.value))')
    expect(source).toContain('formatApiError(e, t.value.chatLoadFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, t.value.chatOpenDirectFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, t.value.chatMessagesLoadFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, t.value.chatOlderMessagesLoadFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, t.value.chatSearchFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, t.value.chatMediaLoadFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, t.value.chatMessageSendFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(err, t.value.reactionFailed, copy.value.apiError)')
    expect(source).toContain('formatApiError(e, t.value.personalPhotosLoadError, copy.value.apiError)')
  })

  it('does not retain two-arg formatApiError calls that can hit English defaults', () => {
    // Guard the contract rather than today's catch-variable names. A future refactor to
    // `catch (error)` must not be able to bypass the locale-aware third argument.
    expect(source).not.toMatch(/formatApiError\(\s*[^,\n]+,\s*[^,\n)]+\)/)
  })
})
