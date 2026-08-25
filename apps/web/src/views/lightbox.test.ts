import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const readSrc = (relativePath: string) =>
  readFileSync(fileURLToPath(new URL(relativePath, import.meta.url)), 'utf-8')

describe('media lightbox contract', () => {
  const lightbox = readSrc('../components/MediaLightbox.vue')
  const photos = readSrc('./PhotosView.vue')
  const album = readSrc('./AlbumView.vue')

  it('renders a dialog with image/video support and close controls', () => {
    expect(lightbox).toMatch(/<dialog/)
    expect(lightbox).toMatch(/<img[\s\S]*?v-if="isImage"/)
    expect(lightbox).toMatch(/<video[\s\S]*?v-else-if="isVideo"/)
    expect(lightbox).toMatch(/<track[\s\S]*?kind="captions"/)
    expect(lightbox).toMatch(/@click="emit\('close'\)"/)
    expect(lightbox).toMatch(/@keydown\.esc/)
  })

  it('loads preview URLs through the existing file download endpoint', () => {
    for (const view of [photos, album]) {
      expect(view).toContain('MediaLightbox')
      expect(view).toMatch(/api<DownloadURL>\(`\/files\/\$\{[^}]+\.id\}\/download`\)/)
    }
  })

  it('opens lightbox from Photos and Album thumbnails', () => {
    expect(photos).toMatch(/@click="openLightbox\(item\)"/)
    expect(album).toMatch(/@click="openLightbox\(item\)"/)
  })
})
