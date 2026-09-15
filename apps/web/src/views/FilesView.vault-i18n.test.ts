import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'
import { vaultMoveCopy } from '@/lib/vaultMoveCopy'

const source = readFileSync(fileURLToPath(new URL('./FilesView.vue', import.meta.url)), 'utf-8')

describe('FilesView Personal Vault localization wiring', () => {
  it('wires single and batch move dialogs/feedback to reactive vault copy', () => {
    expect(source).toContain('const { locale, t } = useI18n()')
    expect(source).toContain('const vaultCopy = computed(() => vaultMoveCopy(locale.value))')
    expect(source).toContain('title: vaultCopy.value.confirmSingleTitle')
    expect(source).toContain('message: vaultCopy.value.confirmSingleMessage(name)')
    expect(source).toContain('title: vaultCopy.value.confirmBatchTitle(fileIds.length)')
    expect(source).toContain('message: vaultCopy.value.confirmBatchMessage(fileIds.length)')
    expect(source).toContain("ui.showToast(vaultCopy.value.moveInSuccess, 'success')")
    expect(source).toContain('ui.showToast(vaultCopy.value.moveInBatchSuccess(fileIds.length)')
    expect(source.match(/formatApiError\(e, vaultCopy\.value\.moveInFailed\)/g)).toHaveLength(2)
  })

  it('does not leave Vietnamese-only move fallbacks in FilesView', () => {
    expect(source).not.toContain("formatApiError(e, 'Không thể chuyển vào kho cá nhân')")
    expect(source).not.toContain('Đã chuyển ${fileIds.length} tệp vào kho cá nhân')
    expect(vaultMoveCopy('en').moveInFailed).toBe('Could not move to Personal Vault')
    expect(vaultMoveCopy('en').moveInBatchSuccess(1)).toBe('Moved 1 file to Personal Vault')
  })
})
