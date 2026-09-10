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
    expect(lightbox).toMatch(/v-if="hasCaptions"/)
    expect(lightbox).not.toMatch(/No captions/)
    expect(lightbox).toMatch(/@click="emit\('close'\)"/)
    expect(lightbox).toMatch(/@keydown\.esc/)
  })

  it('loads preview URLs through the existing file download endpoint', () => {
    const vault = readSrc('./VaultView.vue')
    for (const view of [photos, album, vault]) {
      expect(view).toContain('MediaLightbox')
    }
    expect(photos).toMatch(/api<DownloadURL>\(`\/files\/\$\{[^}]+\.id\}\/download`\)/)
    expect(album).toMatch(/api<DownloadURL>\(`\/files\/\$\{[^}]+\.id\}\/download`\)/)
    expect(vault).toMatch(/api<\{ downloadUrl: string \}>\(`\/files\/\$\{file\.id\}\/download`\)/)
  })

  it('uses direct preview for Photos and Album thumbnails with explicit secondary actions', () => {
    expect(photos).toMatch(/@click="openLightbox\(item\)"/)
    expect(photos).toMatch(/@click="openMediaActions\(item\)"/)
    expect(album).toMatch(/@click="openLightbox\(item\)"/)
    expect(album).toMatch(/@click="openItemActions\(item\)"/)
  })

  it('keeps Photos action-sheet View wired to the lightbox', () => {
    expect(photos).toMatch(/async function handleSheetAfterLeave\(\)[\s\S]*?openLightbox\(mediaItem\.value\)/)
  })

  it('keeps album secondary actions contextual and destructive removal last', () => {
    expect(album).not.toMatch(/label: 'View'/)
    expect(album).toMatch(/label: 'Download'/)
    expect(album).toMatch(/label: 'Remove from this album'/)
    expect(album).toMatch(/action === 'set-cover'/)
    expect(album).toMatch(/action === 'unset-cover'/)
    expect(album).toMatch(/action === 'remove'/)
  })
})
