import type { Locale } from '@/lib/i18n'

export type ChatBubblePickerCopy = {
  cancel: string
  save: string
  title: string
  preview: string
  suggestions: string
  help: string
  stylesAria: string
  savedToast: (styleName: string) => string
}

export function chatBubblePickerCopy(locale: Locale): ChatBubblePickerCopy {
  if (locale === 'vi') {
    return {
      cancel: 'Hủy',
      save: 'Lưu',
      title: 'Chọn kiểu bong bóng',
      preview: 'Nay bạn có thể thay đổi kiểu bong bóng và cuộc trò chuyện sẽ có giao diện mới. Quá ngầu!',
      suggestions: 'Gợi ý',
      help: 'Bong bóng này áp dụng cho tất cả cuộc trò chuyện. Bong bóng này chỉ ảnh hưởng đến các tin nhắn bạn gửi sau khi lưu.',
      stylesAria: 'Kiểu bong bóng',
      savedToast: (styleName) => `Đã đổi kiểu bong bóng sang "${styleName}"`,
    }
  }

  return {
    cancel: 'Cancel',
    save: 'Save',
    title: 'Choose bubble style',
    preview: 'You can now change the bubble style and give your conversations a new look. Nice!',
    suggestions: 'Suggestions',
    help: 'This bubble style applies to all conversations. It only affects messages you send after saving.',
    stylesAria: 'Bubble styles',
    savedToast: (styleName) => `Changed bubble style to "${styleName}"`,
  }
}
