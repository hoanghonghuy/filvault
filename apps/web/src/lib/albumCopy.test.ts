import { describe, expect, it } from 'vitest'
import { albumRuntimeCopy } from './albumCopy'

describe('albumRuntimeCopy', () => {
  it('provides Vietnamese album action and shared API error copy', () => {
    const copy = albumRuntimeCopy('vi')

    expect(copy.loadFailed).toBe('Không thể tải album')
    expect(copy.addPhotos).toBe('Thêm ảnh')
    expect(copy.setCover).toBe('Đặt làm ảnh bìa')
    expect(copy.removeFromAlbum).toBe('Xóa khỏi album này')
    expect(copy.removeConfirmation('ảnh.jpg')).toBe('"ảnh.jpg" sẽ chỉ bị xóa khỏi album này, tệp gốc vẫn được giữ.')
    expect(copy.downloadFailed).toBe('Không thể tải xuống')
    expect(copy.apiError.network).toBe('Không thể kết nối. Hãy kiểm tra mạng và thử lại.')
    expect(copy.apiError.codes?.unauthorized).toBe('Phiên đăng nhập đã hết hạn. Vui lòng đăng nhập lại.')
    expect(copy.apiError.codes?.forbidden).toBe('Bạn không có quyền thực hiện thao tác này.')
    expect(copy.apiError.codes?.not_found).toBe('Không tìm thấy nội dung này hoặc nội dung đã bị xóa.')
    expect(copy.apiError.codes?.conflict).toBe('Nội dung đã thay đổi. Hãy tải lại và thử lại.')
    expect(copy.apiError.codes?.rate_limited).toBe('Bạn thao tác quá nhanh. Vui lòng thử lại sau.')
  })

  it('preserves English action semantics and shared formatter defaults', () => {
    const copy = albumRuntimeCopy('en')

    expect(copy.addPhotos).toBe('Add photos')
    expect(copy.renameAlbum).toBe('Rename album')
    expect(copy.deleteAlbum).toBe('Delete album')
    expect(copy.setCover).toBe('Set cover')
    expect(copy.removeCover).toBe('Remove cover')
    expect(copy.removeFromAlbum).toBe('Remove from this album')
    expect(copy.removeConfirmation('photo.jpg')).toBe('"photo.jpg" will be removed from this album only. The original file will remain in your library.')
    expect(copy.apiError).toEqual({})
  })
})
