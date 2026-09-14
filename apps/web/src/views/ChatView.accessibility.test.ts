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
})
