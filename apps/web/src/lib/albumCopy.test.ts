import { describe, expect, it } from 'vitest'
import { albumRuntimeCopy } from './albumCopy'

describe('albumRuntimeCopy', () => {
  it('provides Vietnamese album action and runtime copy', () => {
    const copy = albumRuntimeCopy('vi')

    expect(copy.loadFailed).toBe('Không thể tải album')
    expect(copy.addPhotos).toBe('Thêm ảnh')
    expect(copy.setCover).toBe('Đặt làm ảnh bìa')
    expect(copy.removeFromAlbum).toBe('Xóa khỏi album này')
    expect(copy.removeConfirmation('ảnh.jpg')).toBe('"ảnh.jpg" sẽ chỉ bị xóa khỏi album này, tệp gốc vẫn được giữ.')
    expect(copy.downloadFailed).toBe('Không thể tải xuống')
  })

  it('preserves existing English action semantics', () => {
    const copy = albumRuntimeCopy('en')

    expect(copy.addPhotos).toBe('Add photos')
    expect(copy.renameAlbum).toBe('Rename album')
    expect(copy.deleteAlbum).toBe('Delete album')
    expect(copy.setCover).toBe('Set cover')
    expect(copy.removeCover).toBe('Remove cover')
    expect(copy.removeFromAlbum).toBe('Remove from this album')
    expect(copy.removeConfirmation('photo.jpg')).toBe('"photo.jpg" will be removed from this album only.')
  })
})
