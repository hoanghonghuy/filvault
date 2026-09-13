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

  it('localizes runtime fallback errors for VI and EN without bypassing API error formatting', () => {
    expect(videoSource).toContain("const { t, locale } = useI18n()")
    expect(videoSource).toContain("load: 'Không thể tải video'")
    expect(videoSource).toContain("loadMore: 'Không thể tải thêm video'")
    expect(videoSource).toContain("view: 'Không thể mở video'")
    expect(videoSource).toContain("download: 'Không thể tải video xuống'")
    expect(videoSource).toContain("favorite: 'Không thể cập nhật mục yêu thích'")
    expect(videoSource).toContain("load: 'Failed to load videos'")
    expect(videoSource).toContain("loadMore: 'Failed to load more videos'")
    expect(videoSource).toContain("view: 'View failed'")
    expect(videoSource).toContain("download: 'Download failed'")
    expect(videoSource).toContain("favorite: 'Failed to update favorite'")
    expect(videoSource).toContain('formatApiError(e, errorCopy.value.load)')
    expect(videoSource).toContain('formatApiError(e, errorCopy.value.loadMore)')
    expect(videoSource).toContain('formatApiError(e, errorCopy.value.view)')
    expect(videoSource).toContain('formatApiError(e, errorCopy.value.download)')
    expect(videoSource).toContain('formatApiError(e, errorCopy.value.favorite)')
  })

  it('keeps preview, secondary actions, load-more and responsive density', () => {
    expect(videoSource).toContain('@click="openLightbox(item)"')
    expect(videoSource).toContain('@click="openActions(item)"')
    expect(videoSource).toContain('v-if="nextBefore"')
    expect(videoSource).toContain('@media (min-width: 768px)')
    expect(videoSource).toContain('@media (min-width: 1200px)')
  })
})
