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
    expect(source).not.toMatch(/formatApiError\(\s*e,\s*t\.value\.[^,)]+\)/)
    expect(source).not.toMatch(/formatApiError\(\s*err,\s*t\.value\.[^,)]+\)/)
    expect(source).not.toContain("formatApiError(e, 'Failed to load chats')")
    expect(source).not.toContain("formatApiError(e, 'Could not open direct chat')")
    expect(source).not.toContain("formatApiError(e, 'Failed to load messages')")
    expect(source).not.toContain("formatApiError(e, 'Failed to load older messages')")
    expect(source).not.toContain("formatApiError(e, 'Search failed')")
  })
})
