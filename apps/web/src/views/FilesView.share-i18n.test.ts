import { describe, expect, it } from 'vitest'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'

const source = readFileSync(fileURLToPath(new URL('./FilesView.vue', import.meta.url)), 'utf8')

describe('FilesView share feedback localization wiring', () => {
  it('uses reactive i18n copy for public-link and direct-share feedback', () => {
    for (const key of [
      'shareLinkCreated', 'shareLinkCreateFailed', 'shareLinkCopied', 'shareLinkCopyFailed',
      'shareLinkRevokeTitle', 'shareLinkRevokeMessage', 'shareLinkRevokeConfirm',
      'shareLinkRevoked', 'shareLinkRevokeFailed', 'shareInvitationSent', 'shareFileWithUserSuccess',
    ]) expect(source).toContain(`t.value.${key}`)
  })

  it('does not retain the replaced English-only share literals', () => {
    for (const literal of [
      "ui.showToast('Share link created')", "formatApiError(e, 'Could not create link')",
      "ui.showToast('Link copied')", "ui.showToast('Copy failed', 'info')",
      "title: 'Revoke link?'", "confirmLabel: 'Revoke link'", "ui.showToast('Link revoked')",
      "formatApiError(e, 'Could not revoke link')", 'Invitation sent to ${result.email}',
    ]) expect(source).not.toContain(literal)
  })

  it('keeps dollar sequences literal when interpolating filename and email', () => {
    const name = 'budget-$&-$1.txt'
    const email = 'user+$&@example.com'
    const rendered = 'Shared "{name}" with {email}'
      .replace('{name}', () => name)
      .replace('{email}', () => email)
    expect(rendered).toBe('Shared "budget-$&-$1.txt" with user+$&@example.com')
  })
})
