import { describe, expect, it } from 'vitest'
import { mediaPickerCopy } from './mediaPickerCopy'

describe('mediaPickerCopy', () => {
  it('provides the full Vietnamese album media-picker contract', () => {
    expect(mediaPickerCopy('vi')).toEqual({
      title: 'Thêm vào album',
      hint: 'Chạm vào ảnh hoặc video để thêm.',
      loadFailed: 'Không thể tải ảnh và video',
      loadMoreFailed: 'Không thể tải thêm',
      emptyTitle: 'Không có ảnh hoặc video khả dụng',
      emptyDescription: 'Hãy tải ảnh hoặc video lên Tệp của tôi trước.',
    })
  })

  it('provides the full English album media-picker contract', () => {
    expect(mediaPickerCopy('en')).toEqual({
      title: 'Add to album',
      hint: 'Tap a photo or video to add it.',
      loadFailed: 'Failed to load photos',
      loadMoreFailed: 'Failed to load more',
      emptyTitle: 'No media available',
      emptyDescription: 'Upload photos or videos in My Files first.',
    })
  })
})
