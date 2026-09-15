import type { Locale } from '@/lib/i18n'

export type PhotosRuntimeCopy = {
  loadPhotosFailed: string
  loadMoreFailed: string
  albumCreated: string
  createAlbumFailed: string
  openAlbum: string
  renameAlbum: string
  deleteAlbum: string
  renameFailed: string
  albumDeleted: string
  deleteFailed: string
  addFavorite: (name: string) => string
  removeFavorite: (name: string) => string
  favoriteFailed: string
  viewFailed: string
  downloadFailed: string
  pullToRefreshPhotos: string
  photosViewsAria: string
}

const copy: Record<Locale, PhotosRuntimeCopy> = {
  vi: {
    loadPhotosFailed: 'Không thể tải ảnh',
    loadMoreFailed: 'Không thể tải thêm',
    albumCreated: 'Đã tạo album',
    createAlbumFailed: 'Không thể tạo album',
    openAlbum: 'Mở',
    renameAlbum: 'Đổi tên',
    deleteAlbum: 'Xóa album',
    renameFailed: 'Không thể đổi tên album',
    albumDeleted: 'Đã xóa album',
    deleteFailed: 'Không thể xóa album',
    addFavorite: (name) => `Đã thêm "${name}" vào mục yêu thích`,
    removeFavorite: (name) => `Đã bỏ "${name}" khỏi mục yêu thích`,
    favoriteFailed: 'Không thể cập nhật mục yêu thích',
    viewFailed: 'Không thể mở nội dung',
    downloadFailed: 'Không thể tải xuống',
    pullToRefreshPhotos: 'Kéo để làm mới ảnh',
    photosViewsAria: 'Các chế độ xem ảnh',
  },
  en: {
    loadPhotosFailed: 'Failed to load photos',
    loadMoreFailed: 'Failed to load more',
    albumCreated: 'Album created',
    createAlbumFailed: 'Failed to create album',
    openAlbum: 'Open',
    renameAlbum: 'Rename',
    deleteAlbum: 'Delete album',
    renameFailed: 'Rename failed',
    albumDeleted: 'Album deleted',
    deleteFailed: 'Delete failed',
    addFavorite: (name) => `Added "${name}" to favorites`,
    removeFavorite: (name) => `Removed "${name}" from favorites`,
    favoriteFailed: 'Failed to update favorite',
    viewFailed: 'View failed',
    downloadFailed: 'Download failed',
    pullToRefreshPhotos: 'Pull to refresh photos',
    photosViewsAria: 'Photos views',
  },
}

export function photosRuntimeCopy(locale: Locale): PhotosRuntimeCopy {
  return copy[locale]
}
