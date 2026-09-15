import { describe, expect, it } from 'vitest'
import { photosRuntimeCopy } from './photosCopy'

describe('photosRuntimeCopy', () => {
  it('provides Vietnamese runtime, action and accessibility copy', () => {
    const copy = photosRuntimeCopy('vi')

    expect(copy.loadPhotosFailed).toBe('Không thể tải ảnh')
    expect(copy.openAlbum).toBe('Mở')
    expect(copy.deleteAlbum).toBe('Xóa album')
    expect(copy.addFavorite('ảnh.jpg')).toBe('Đã thêm "ảnh.jpg" vào mục yêu thích')
    expect(copy.removeFavorite('ảnh.jpg')).toBe('Đã bỏ "ảnh.jpg" khỏi mục yêu thích')
    expect(copy.pullToRefreshPhotos).toBe('Kéo để làm mới ảnh')
    expect(copy.photosViewsAria).toBe('Các chế độ xem ảnh')
  })

  it('preserves the existing English semantics', () => {
    const copy = photosRuntimeCopy('en')

    expect(copy.loadPhotosFailed).toBe('Failed to load photos')
    expect(copy.openAlbum).toBe('Open')
    expect(copy.renameAlbum).toBe('Rename')
    expect(copy.deleteAlbum).toBe('Delete album')
    expect(copy.addFavorite('photo.jpg')).toBe('Added "photo.jpg" to favorites')
    expect(copy.removeFavorite('photo.jpg')).toBe('Removed "photo.jpg" from favorites')
    expect(copy.pullToRefreshPhotos).toBe('Pull to refresh photos')
    expect(copy.photosViewsAria).toBe('Photos views')
  })
})
