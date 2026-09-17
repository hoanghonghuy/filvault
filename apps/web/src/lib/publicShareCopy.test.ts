import { describe, expect, it } from 'vitest'
import { publicShareCopy } from './publicShareCopy'

describe('publicShareCopy', () => {
  it('provides Vietnamese public-share labels', () => {
    expect(publicShareCopy('vi')).toEqual({
      loading: 'Đang tải liên kết…',
      unavailable: 'Liên kết này không còn khả dụng',
      temporary: 'Tạm thời không thể tải liên kết này.',
      retry: 'Thử lại',
      expires: 'Hết hạn',
      preview: 'Xem trước',
      previewing: 'Đang chuẩn bị xem trước…',
      previewFailed:
        'Không thể chuẩn bị bản xem trước. Liên kết vẫn còn hiệu lực; hãy thử lại.',
      download: 'Tải xuống',
      downloading: 'Đang chuẩn bị tải xuống…',
      downloadFailed:
        'Không thể chuẩn bị tệp tải xuống. Liên kết vẫn còn hiệu lực; hãy thử lại.',
      sharedVia: 'Được chia sẻ qua Filvault',
    })
  })

  it('provides English public-share labels', () => {
    expect(publicShareCopy('en')).toEqual({
      loading: 'Loading link…',
      unavailable: 'This link is not available',
      temporary: 'This link could not be loaded right now.',
      retry: 'Retry',
      expires: 'Expires',
      preview: 'Preview',
      previewing: 'Preparing preview…',
      previewFailed: 'Could not prepare the preview. The link is still available; try again.',
      download: 'Download',
      downloading: 'Preparing download…',
      downloadFailed: 'Could not prepare the download. The link is still available; try again.',
      sharedVia: 'Shared via Filvault',
    })
  })
})
