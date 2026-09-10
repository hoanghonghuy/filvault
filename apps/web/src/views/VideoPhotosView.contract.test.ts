import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const videoSource = readFileSync(fileURLToPath(new URL('./VideoPhotosView.vue', import.meta.url)), 'utf-8')
const routeSource = readFileSync(fileURLToPath(new URL('./PhotosRouteView.vue', import.meta.url)), 'utf-8')

describe('video-focused Photos route contract', () => {
  it('honors the Overview type=video route intent', () => {
    expect(routeSource).toContain("route.query.type === 'video'")
    expect(routeSource).toContain('<VideoPhotosView v-if="videoOnly" />')
    expect(routeSource).toContain('<PhotosView v-else />')
  })

  it('uses the server-filtered timeline so filtering happens before pagination', () => {
    expect(videoSource).toContain("new URLSearchParams({ type: 'video' })")
    expect(videoSource).toContain("api<Timeline>(`/photos/timeline/filter?${params}`)")
    expect(videoSource).toContain('loadVideoTimeline(nextBefore.value)')
    expect(videoSource).not.toContain('while (matches.length <= PAGE_SIZE)')
    expect(videoSource).not.toContain("item.mimeType.startsWith('video/')")
  })

  it('uses authoritative per-item favorite state without a capped favorites preload', () => {
    expect(videoSource).not.toContain('/files/favorites?limit=100')
    expect(videoSource).toContain('Boolean(item.isFavorite)')
    expect(videoSource).toContain('item.isFavorite = !wasFavorited')
  })

  it('keeps preview, secondary actions, load-more and responsive density', () => {
    expect(videoSource).toContain('@click="openLightbox(item)"')
    expect(videoSource).toContain('@click="openActions(item)"')
    expect(videoSource).toContain('v-if="nextBefore"')
    expect(videoSource).toContain('@media (min-width: 768px)')
    expect(videoSource).toContain('@media (min-width: 1200px)')
  })
})
