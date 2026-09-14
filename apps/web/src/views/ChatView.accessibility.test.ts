/// <reference types="node" />

import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const source = readFileSync(fileURLToPath(new URL('./ChatView.vue', import.meta.url)), 'utf-8')

describe('ChatView accessibility contract', () => {
  it('keeps mobile thread and composer actions at the shared touch minimum without changing desktop density', () => {
    expect(source).toContain(':aria-label="t.back"')
    expect(source).toContain(':aria-label="t.callVoice"')
    expect(source).toContain(':aria-label="t.callVideo"')
    expect(source).toContain(':aria-label="t.chatInfo"')

    const mobileStart = source.indexOf('@media (max-width: 767px)')
    const mobileEnd = source.indexOf('\n.clickable {', mobileStart)
    const mobile = source.slice(mobileStart, mobileEnd)

    expect(mobile).toMatch(/\.back-btn,\s*\n\s*\.thread-actions \.icon-btn,\s*\n\s*\.attach-btn,\s*\n\s*\.sticker-toggle-btn,\s*\n\s*\.like-btn,\s*\n\s*\.send-btn\s*\{[\s\S]*?width:\s*var\(--touch-min\);[\s\S]*?height:\s*var\(--touch-min\);[\s\S]*?min-width:\s*var\(--touch-min\);[\s\S]*?min-height:\s*var\(--touch-min\);/)

    const desktopRule = source.match(/\.thread-actions \.icon-btn\s*\{([\s\S]*?)\n\}/)?.[1] ?? ''
    expect(desktopRule).toContain('width: 36px;')
    expect(desktopRule).toContain('height: 36px;')
  })
  it('keeps reaction picker controls at the shared touch minimum', () => {
    const reactionButtonRule = source.match(/\.rx-btn\s*\{([\s\S]*?)\n\}/)?.[1] ?? ''
    expect(reactionButtonRule).toContain('width: var(--touch-min);')
    expect(reactionButtonRule).toContain('height: var(--touch-min);')

    const reactionPlusRule = source.match(/\.rx-plus-btn\s*\{([\s\S]*?)\n\}/)?.[1] ?? ''
    expect(reactionPlusRule).not.toMatch(/(?:width|height):\s*\d+px;/)
    expect(source).toContain('class="rx-btn rx-plus-btn"')
    expect(source).toContain(':aria-label="t.addReaction"')
  })

  it('keeps desktop Chat Info close action at the shared touch minimum', () => {
    expect(source).toContain('class="icon-btn desktop-info-close-btn"')
    expect(source).toContain(':aria-label="t.close"')
    expect(source).toContain(':title="t.close"')
    expect(source).toContain('@click="closeDesktopInfo"')

    const desktopInfoCloseRule = source.match(/\.desktop-info-close-btn\s*\{([\s\S]*?)\n\}/)?.[1] ?? ''
    expect(desktopInfoCloseRule).toContain('width: var(--touch-min);')
    expect(desktopInfoCloseRule).toContain('height: var(--touch-min);')
    expect(desktopInfoCloseRule).toContain('min-width: var(--touch-min);')
    expect(desktopInfoCloseRule).toContain('min-height: var(--touch-min);')
  })

  it('keeps Chat Info subpage close actions at the shared touch minimum', () => {
    expect(source.match(/class="subpage-close-btn"/g)?.length).toBe(2)
    expect(source.match(/class="subpage-close-btn" :aria-label="t.close"/g)?.length).toBe(2)

    const closeRule = source.match(/\.subpage-close-btn\s*\{([\s\S]*?)\n\}/)?.[1] ?? ''
    expect(closeRule).toContain('width: var(--touch-min);')
    expect(closeRule).toContain('height: var(--touch-min);')
    expect(closeRule).toContain('min-width: var(--touch-min);')
    expect(closeRule).toContain('min-height: var(--touch-min);')
  })

})
