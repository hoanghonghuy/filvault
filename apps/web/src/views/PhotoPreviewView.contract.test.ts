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

  it('allows only ready image or video media and fails without raw API detail', () => {
    expect(preview).toContain("file.status === 'READY'")
    expect(preview).toContain("file.mimeType.startsWith('image/')")
    expect(preview).toContain("file.mimeType.startsWith('video/')")
    expect(preview).toContain('This photo or video is unavailable.')
    expect(preview).not.toContain('formatApiError')
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
