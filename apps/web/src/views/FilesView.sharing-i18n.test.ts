import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'

const source = readFileSync(fileURLToPath(new URL('./FilesView.vue', import.meta.url)), 'utf8')

describe('FilesView sharing feedback localization wiring', () => {
  it('uses locale-reactive copy for public-link lifecycle and direct sharing', () => {
    expect(source).toContain('ui.showToast(t.value.shareLinkCreated)')
    expect(source).toContain('formatApiError(e, t.value.shareLinkCreateFailed)')
    expect(source).toContain('ui.showToast(t.value.shareLinkCopied)')
    expect(source).toContain("ui.showToast(t.value.shareLinkCopyFailed, 'info')")
    expect(source).toContain('title: t.value.shareLinkRevokeTitle')
    expect(source).toContain('message: t.value.shareLinkRevokeMessage')
    expect(source).toContain('confirmLabel: t.value.shareLinkRevokeConfirm')
    expect(source).toContain('ui.showToast(t.value.shareLinkRevoked)')
    expect(source).toContain('formatApiError(e, t.value.shareLinkRevokeFailed)')
    expect(source).toContain("t.value.shareInvitationSent.replace('{email}', () => result.email)")
    expect(source).toContain(".replace('{name}', () => name)")
    expect(source).toContain(".replace('{email}', () => result.email)")
  })

  it('preserves dollar sequences in filenames and emails literally during interpolation', () => {
    const name = 'budget-$&-$1.txt'
    const email = 'user+$&-$1@example.com'
    const template = 'Shared "{name}" with {email}'
    expect(template.replace('{name}', () => name).replace('{email}', () => email)).toBe(
      'Shared "budget-$&-$1.txt" with user+$&-$1@example.com',
    )
  })

  it('does not retain the replaced English-only sharing feedback', () => {
    expect(source).not.toContain("ui.showToast('Share link created')")
    expect(source).not.toContain("formatApiError(e, 'Could not create link')")
    expect(source).not.toContain("ui.showToast('Link copied')")
    expect(source).not.toContain("ui.showToast('Copy failed', 'info')")
    expect(source).not.toContain("title: 'Revoke link?'")
    expect(source).not.toContain("message: 'Anyone who has this link will lose access immediately.'")
    expect(source).not.toContain("confirmLabel: 'Revoke link'")
    expect(source).not.toContain("ui.showToast('Link revoked')")
    expect(source).not.toContain("formatApiError(e, 'Could not revoke link')")
    expect(source).not.toContain('Invitation sent to ${result.email}')
    expect(source).not.toContain('Shared "${name}" with ${result.email}')
  })
})
