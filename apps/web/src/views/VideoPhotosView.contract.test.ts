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

  it('filters across source pages instead of only the first mixed timeline page', () => {
    expect(videoSource).toContain('while (matches.length <= PAGE_SIZE)')
    expect(videoSource).toContain("item.mimeType.startsWith('video/')")
    expect(videoSource).toContain('if (matches.length > PAGE_SIZE || !data.nextBefore) break')
    expect(videoSource).toContain('scanBefore = data.nextBefore')
    expect(videoSource).toContain("matches.length > PAGE_SIZE ? pageItems.at(-1)?.createdAt : undefined")
  })

  it('keeps preview, secondary actions, load-more and responsive density', () => {
    expect(videoSource).toContain('@click="openLightbox(item)"')
    expect(videoSource).toContain('@click="openActions(item)"')
    expect(videoSource).toContain('v-if="nextBefore"')
    expect(videoSource).toContain('@media (min-width: 768px)')
    expect(videoSource).toContain('@media (min-width: 1200px)')
  })
})
