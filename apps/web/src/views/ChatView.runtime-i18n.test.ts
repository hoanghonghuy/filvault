import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'

const source = readFileSync(fileURLToPath(new URL('./ChatView.vue', import.meta.url)), 'utf8')

describe('ChatView runtime loading localization wiring', () => {
  it('uses locale-reactive fallbacks and shared API error copy for conversation, history and search errors', () => {
    expect(source).toContain("import { chatApiErrorCopy } from '@/lib/chatApiErrorCopy'")
    expect(source).toContain('chatApiErrorCopy(locale.value).apiError')
    expect(source).toContain('formatApiError(e, t.value.chatLoadFailed, apiErrorCopy.value)')
    expect(source).toContain('formatApiError(e, t.value.chatOpenDirectFailed, apiErrorCopy.value)')
    expect(source).toContain('formatApiError(e, t.value.chatMessagesLoadFailed, apiErrorCopy.value)')
    expect(source).toContain('formatApiError(e, t.value.chatOlderMessagesLoadFailed, apiErrorCopy.value)')
    expect(source).toContain('formatApiError(e, t.value.chatSearchFailed, apiErrorCopy.value)')
  })

  it('does not retain the replaced English-only runtime fallbacks', () => {
    expect(source).not.toContain("formatApiError(e, 'Failed to load chats')")
    expect(source).not.toContain("formatApiError(e, 'Could not open direct chat')")
    expect(source).not.toContain("formatApiError(e, 'Failed to load messages')")
    expect(source).not.toContain("formatApiError(e, 'Failed to load older messages')")
    expect(source).not.toContain("formatApiError(e, 'Search failed')")
  })
})
