import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const preview = readFileSync(fileURLToPath(new URL('./PhotoPreviewView.vue', import.meta.url)), 'utf-8')
const overview = readFileSync(fileURLToPath(new URL('./OverviewView.vue', import.meta.url)), 'utf-8')
const router = readFileSync(fileURLToPath(new URL('../router/index.ts', import.meta.url)), 'utf-8')

describe('photo deep-link contract', () => {
  it('resolves a specific owned file independently of timeline pagination', () => {
    expect(preview).toContain('api<OwnedFile>(`/files/${encodeURIComponent(id)}`)')
    expect(preview).toContain('api<DownloadURL>(`/files/${encodeURIComponent(id)}/download`)')
  })

  it('allows only ready image or video media and keeps error state locale-reactive', () => {
    expect(preview).toContain("file.status === 'READY'")
    expect(preview).toContain("file.mimeType.startsWith('image/')")
    expect(preview).toContain("file.mimeType.startsWith('video/')")
    expect(preview).toContain("error.value = 'unavailable'")
    expect(preview).toContain("error.value = 'unsupported'")
    expect(preview).toContain("locale.value === 'vi'")
    expect(preview).toContain("unavailable: 'Ảnh hoặc video này hiện không khả dụng.'")
    expect(preview).toContain("unavailable: 'This photo or video is unavailable.'")
    expect(preview).toContain("unsupported: 'Mục này không thể xem trước trong Ảnh.'")
    expect(preview).toContain("unsupported: 'This item cannot be previewed in Photos.'")
    expect(preview).not.toContain('formatApiError')
  })

  it('localizes loading, recovery and navigation copy in VI and EN', () => {
    expect(preview).toContain("photos: 'Ảnh'")
    expect(preview).toContain("photos: 'Photos'")
    expect(preview).toContain("loading: 'Đang tải bản xem trước…'")
    expect(preview).toContain("loading: 'Loading preview…'")
    expect(preview).toContain("unavailableTitle: 'Không thể xem trước'")
    expect(preview).toContain("unavailableTitle: 'Preview unavailable'")
    expect(preview).toContain("openPhotos: 'Mở Ảnh'")
    expect(preview).toContain("openPhotos: 'Open Photos'")
    expect(preview).toContain('{{ copy.loading }}')
    expect(preview).toContain('{{ errorMessage }}')
  })

  it('returns to a usable Photos surface when preview closes', () => {
    expect(preview).toContain("router.replace('/photos')")
  })

  it('wires Overview photos to the stable preview route', () => {
    expect(overview).toContain('router.push(`/photos/preview/${item.id}`)')
    expect(router).toContain("path: '/photos/preview/:id'")
    expect(router).toContain("name: 'photo-preview'")
  })
})
