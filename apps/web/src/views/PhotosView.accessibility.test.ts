import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const source = readFileSync(fileURLToPath(new URL('./PhotosView.vue', import.meta.url)), 'utf-8')

describe('PhotosView accessibility and primary-action contract', () => {
  it('connects roving tabs to labelled tab panels', () => {
    expect(source).toContain('aria-controls="photos-panel-timeline"')
    expect(source).toContain('aria-controls="photos-panel-albums"')
    expect(source).toContain('aria-labelledby="photos-tab-timeline"')
    expect(source).toContain('aria-labelledby="photos-tab-albums"')
    expect(source).toContain("['ArrowLeft', 'ArrowRight', 'Home', 'End']")
    expect(source).toContain(":tabindex=\"activePhotoTab === 'timeline' ? 0 : -1\"")
    expect(source).toContain(":tabindex=\"activePhotoTab === 'albums' ? 0 : -1\"")
  })

  it('uses direct preview as media primary action while keeping secondary actions explicit', () => {
    expect(source).toContain('@click="openLightbox(item)"')
    expect(source).toContain('class="photo-more-btn"')
    expect(source).toContain('@click="openMediaActions(item)"')
  })

  it('makes album primary actions semantic and exposes refresh state', () => {
    expect(source).toContain('<RouterLink class="album-primary"')
    expect(source).toContain('role="status"')
    expect(source).toContain('aria-live="polite"')
    expect(source).toContain('@media (hover: none), (pointer: coarse)')
  })
})
