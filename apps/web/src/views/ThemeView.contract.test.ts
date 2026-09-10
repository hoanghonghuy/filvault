/**
 * @vitest-environment node
 */
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const source = readFileSync(fileURLToPath(new URL('./ThemeView.vue', import.meta.url)), 'utf-8')

describe('ThemeView product contract', () => {
  it('keeps primary theme controls and removes dead navigation', () => {
    expect(source).not.toContain('role="tablist"')
    expect(source).not.toContain('role="tab"')
    expect(source).not.toContain('activeTab')
    expect(source).toContain('class="theme-header-title"')
    expect(source).toContain('role="radiogroup"')
    expect(source).toContain('class="swatches-grid"')
    expect(source).toContain('class="seasonal-grid"')
  })

  it('previews real Filvault storage surfaces instead of calendar/task UI', () => {
    expect(source).toContain('class="storage-preview"')
    expect(source).toContain('class="storage-preview-folder selected-preview-item"')
    expect(source).toContain('class="storage-preview-file"')
    expect(source).toContain('class="storage-preview-progress"')
    expect(source).not.toContain('mockup-dates-grid')
    expect(source).not.toContain('mockup-task-item')
    expect(source).not.toContain('themePreviewTask1')
  })

  it('defines a labelled keyboard-contained preview dialog with safe motion and mobile insets', () => {
    expect(source).toContain('aria-labelledby="theme-preview-title"')
    expect(source).toContain('@keydown="handlePreviewKeydown"')
    expect(source).toContain("event.key === 'Escape'")
    expect(source).toContain("event.key !== 'Tab'")
    expect(source).toContain('previewCloseButton.value?.focus()')
    expect(source).toContain('previewOpener?.focus()')
    expect(source).toContain('env(safe-area-inset-top)')
    expect(source).toContain('@media (prefers-reduced-motion:reduce)')
  })
})
