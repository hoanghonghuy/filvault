export type OverviewLocale = 'vi' | 'en'

export type OverviewCopy = {
  clearSearchAria: string
  featureCategoriesAria: string
  loadOverviewFailed: string
  loadFilesFailed: string
  loadPhotosFailed: string
}

const COPY: Record<OverviewLocale, OverviewCopy> = {
  vi: {
    clearSearchAria: 'Xóa tìm kiếm',
    featureCategoriesAria: 'Danh mục tính năng',
    loadOverviewFailed: 'Không thể tải trang tổng quan. Vui lòng thử lại sau.',
    loadFilesFailed: 'Không thể tải tệp',
    loadPhotosFailed: 'Không thể tải ảnh',
  },
  en: {
    clearSearchAria: 'Clear search',
    featureCategoriesAria: 'Feature categories',
    loadOverviewFailed: 'Could not load overview. Try again later.',
    loadFilesFailed: 'Could not load files',
    loadPhotosFailed: 'Could not load photos',
  },
}

export function overviewCopy(locale: OverviewLocale): OverviewCopy {
  return COPY[locale]
}
