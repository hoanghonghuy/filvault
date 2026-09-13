import type { Locale } from '@/lib/i18n'

export type MediaPickerCopy = {
  title: string
  hint: string
  loadFailed: string
  loadMoreFailed: string
  emptyTitle: string
  emptyDescription: string
}

const copy: Record<Locale, MediaPickerCopy> = {
  vi: {
    title: 'Thêm vào album',
    hint: 'Chạm vào ảnh hoặc video để thêm.',
    loadFailed: 'Không thể tải ảnh và video',
    loadMoreFailed: 'Không thể tải thêm',
    emptyTitle: 'Không có ảnh hoặc video khả dụng',
    emptyDescription: 'Hãy tải ảnh hoặc video lên Tệp của tôi trước.',
  },
  en: {
    title: 'Add to album',
    hint: 'Tap a photo or video to add it.',
    loadFailed: 'Failed to load photos',
    loadMoreFailed: 'Failed to load more',
    emptyTitle: 'No media available',
    emptyDescription: 'Upload photos or videos in My Files first.',
  },
}

export function mediaPickerCopy(locale: Locale): MediaPickerCopy {
  return copy[locale]
}
