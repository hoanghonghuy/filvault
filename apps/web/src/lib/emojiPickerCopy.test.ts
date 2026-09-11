import { describe, expect, it } from 'vitest'
import { emojiPickerCopy } from './emojiPickerCopy'

describe('emojiPickerCopy', () => {
  it('localizes picker controls and accessibility chrome', () => {
    expect(emojiPickerCopy('vi')).toMatchObject({
      regionAria: 'Bộ chọn biểu tượng cảm xúc và nhãn dán',
      emojiType: 'Biểu tượng',
      stickerType: 'Nhãn dán',
      closeAria: 'Đóng',
    })
    expect(emojiPickerCopy('en')).toMatchObject({
      regionAria: 'Emoji and sticker picker',
      emojiType: 'Emoji',
      stickerType: 'Stickers',
      closeAria: 'Close',
    })
  })
})
