export type EmojiPickerLocale = 'vi' | 'en'

export type EmojiPickerCopy = {
  regionAria: string
  typeAria: string
  emojiType: string
  stickerType: string
  closeAria: string
  emojiCategoriesAria: string
  stickerCategoriesAria: string
}

const COPY: Record<EmojiPickerLocale, EmojiPickerCopy> = {
  vi: {
    regionAria: 'Bộ chọn biểu tượng cảm xúc và nhãn dán',
    typeAria: 'Loại biểu tượng',
    emojiType: 'Biểu tượng',
    stickerType: 'Nhãn dán',
    closeAria: 'Đóng',
    emojiCategoriesAria: 'Danh mục biểu tượng',
    stickerCategoriesAria: 'Danh mục nhãn dán',
  },
  en: {
    regionAria: 'Emoji and sticker picker',
    typeAria: 'Picker type',
    emojiType: 'Emoji',
    stickerType: 'Stickers',
    closeAria: 'Close',
    emojiCategoriesAria: 'Emoji categories',
    stickerCategoriesAria: 'Sticker categories',
  },
}

export function emojiPickerCopy(locale: EmojiPickerLocale): EmojiPickerCopy {
  return COPY[locale]
}
