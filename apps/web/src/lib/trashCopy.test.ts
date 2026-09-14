import { describe, expect, it } from 'vitest'
import { formatTrashItemDate, trashCopy } from './trashCopy'

describe('trashCopy', () => {
  it('formats result counts with locale-aware grammar', () => {
    expect(trashCopy('en').resultCount(1)).toBe('1 result')
    expect(trashCopy('en').resultCount(2)).toBe('2 results')
    expect(trashCopy('vi').resultCount(1)).toBe('1 kết quả')
    expect(trashCopy('vi').resultCount(2)).toBe('2 kết quả')
  })

  it('keeps destructive confirmation explicit in English and Vietnamese', () => {
    const english = trashCopy('en')
    const vietnamese = trashCopy('vi')

    expect(english.deleteConfirmMessage('archive.zip')).toContain('cannot be undone')
    expect(english.emptyConfirmMessage).toContain('cannot be undone')
    expect(vietnamese.deleteConfirmMessage('archive.zip')).toContain('không thể hoàn tác')
    expect(vietnamese.emptyConfirmMessage).toContain('không thể hoàn tác')
  })

  it('localizes recovery and progress copy', () => {
    expect(trashCopy('en').deleting(2, 4)).toBe('Deleting 2/4…')
    expect(trashCopy('vi').deleting(2, 4)).toBe('Đang xóa 2/4…')
    expect(trashCopy('en').restoreFailed).toBe('Could not restore item')
    expect(trashCopy('vi').restoreFailed).toBe('Không thể khôi phục mục')
  })

  it('formats deleted dates with the active locale and rejects invalid input', () => {
    const iso = '2026-09-10T12:00:00Z'
    const en = formatTrashItemDate(iso, 'en')
    const vi = formatTrashItemDate(iso, 'vi')

    expect(en).not.toBe(vi)
    expect(en).toMatch(/Sep/)
    expect(vi).toMatch(/thg 9|tháng 9|9/)
    expect(formatTrashItemDate('invalid', 'en')).toBe('')
  })
})
