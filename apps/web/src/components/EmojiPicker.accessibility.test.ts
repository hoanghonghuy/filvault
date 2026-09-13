/// <reference types="node" />

import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const source = readFileSync(fileURLToPath(new URL('./EmojiPicker.vue', import.meta.url)), 'utf8')

describe('EmojiPicker accessibility contract', () => {
  it('keeps shared picker chrome at the minimum touch target', () => {
    const typeSwitch = source.match(/\.type-switch-btn\s*\{([\s\S]*?)\n\}/)?.[1] ?? ''
    const close = source.match(/\.picker-close-btn\s*\{([\s\S]*?)\n\}/)?.[1] ?? ''
    const category = source.match(/\.picker-cat-btn\s*\{([\s\S]*?)\n\}/)?.[1] ?? ''

    expect(typeSwitch).toContain('min-height: var(--touch-min);')
    expect(close).toContain('width: var(--touch-min);')
    expect(close).toContain('height: var(--touch-min);')
    expect(close).toContain('min-width: var(--touch-min);')
    expect(close).toContain('min-height: var(--touch-min);')
    expect(category).toContain('min-height: var(--touch-min);')

    expect(source).toContain(':aria-label="copy.closeAria"')
    expect(source).toContain('role="tablist"')
    expect(source).toContain('role="tab"')
  })
})
