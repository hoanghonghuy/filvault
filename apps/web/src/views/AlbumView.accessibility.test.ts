import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const source = readFileSync(fileURLToPath(new URL('./AlbumView.vue', import.meta.url)), 'utf-8')

describe('AlbumView accessibility contract', () => {
  it('keeps a semantic path back to Photos at every breakpoint', () => {
    expect(source).toContain('<RouterLink to="/photos" class="btn ghost album-back"')
    expect(source).not.toContain('.mobile-back')
    expect(source).not.toMatch(/@media \(min-width: 768px\)[\s\S]*?album-back[\s\S]*?display:\s*none/)
  })

  it('separates preview from secondary media actions', () => {
    expect(source).toContain('@click="openLightbox(item)"')
    expect(source).toContain('class="album-media-more"')
    expect(source).toContain('@click="openItemActions(item)"')
    expect(source).toContain(":aria-label=\"`${t.moreActions}: ${item.name}`\"")
  })

  it('keeps initial load failure separate from a true-empty album and exposes retry', () => {
    expect(source).toContain('v-if="error && album"')
    expect(source).toContain('v-else-if="error && !album"')
    expect(source).toContain('class="album-error-state" role="alert"')
    expect(source).toContain('@click="load"')
    expect(source).toContain('{{ t.retry }}')
    expect(source).toContain(':disabled="!album || loading"')
    expect(source).toContain('.album-retry { min-height: var(--touch-min); }')
  })

  it('keeps touch and keyboard affordances explicit', () => {
    expect(source).toMatch(/\.album-media-more\s*\{[\s\S]*?width:\s*44px;[\s\S]*?height:\s*44px;/)
    expect(source).toContain('.album-media-more:focus-visible')
    expect(source).toContain('@media (hover: none), (pointer: coarse)')
    expect(source).toContain('@media (prefers-reduced-motion: reduce)')
  })
})
