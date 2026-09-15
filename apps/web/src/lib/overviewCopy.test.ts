import { describe, expect, it } from 'vitest'
import { overviewCopy } from './overviewCopy'

describe('overviewCopy', () => {
  it('localizes Overview accessibility labels', () => {
    expect(overviewCopy('vi').clearSearchAria).toBe('Xóa tìm kiếm')
    expect(overviewCopy('vi').featureCategoriesAria).toBe('Danh mục tính năng')
    expect(overviewCopy('en').clearSearchAria).toBe('Clear search')
    expect(overviewCopy('en').featureCategoriesAria).toBe('Feature categories')
  })

  it('keeps Overview failure fallbacks locale-consistent', () => {
    const vi = overviewCopy('vi')
    const en = overviewCopy('en')

    expect(vi.loadOverviewFailed).toContain('Không thể')
    expect(vi.loadFilesFailed).toBe('Không thể tải tệp')
    expect(vi.loadPhotosFailed).toBe('Không thể tải ảnh')

    expect(en.loadOverviewFailed).toBe('Could not load overview. Try again later.')
    expect(en.loadFilesFailed).toBe('Could not load files')
    expect(en.loadPhotosFailed).toBe('Could not load photos')
  })
})
