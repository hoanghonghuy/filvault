import { describe, expect, it } from 'vitest'
import { chatBubblePickerCopy } from '@/lib/chatBubblePickerCopy'

describe('chatBubblePickerCopy', () => {
  it('provides Vietnamese picker chrome and toast copy', () => {
    const copy = chatBubblePickerCopy('vi')

    expect(copy.cancel).toBe('Hủy')
    expect(copy.save).toBe('Lưu')
    expect(copy.title).toBe('Chọn kiểu bong bóng')
    expect(copy.suggestions).toBe('Gợi ý')
    expect(copy.stylesAria).toBe('Kiểu bong bóng')
    expect(copy.savedToast('Heart')).toBe('Đã đổi kiểu bong bóng sang "Heart"')
  })

  it('provides English picker chrome and toast copy', () => {
    const copy = chatBubblePickerCopy('en')

    expect(copy.cancel).toBe('Cancel')
    expect(copy.save).toBe('Save')
    expect(copy.title).toBe('Choose bubble style')
    expect(copy.suggestions).toBe('Suggestions')
    expect(copy.stylesAria).toBe('Bubble styles')
    expect(copy.savedToast('Heart')).toBe('Changed bubble style to "Heart"')
    expect(copy.preview).not.toMatch(/[À-ỹ]/)
    expect(copy.help).not.toMatch(/[À-ỹ]/)
  })
})
