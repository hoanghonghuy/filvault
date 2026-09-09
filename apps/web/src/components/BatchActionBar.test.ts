/**
 * @vitest-environment jsdom
 */
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { beforeEach, describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import BatchActionBar from './BatchActionBar.vue'
import { setLocale } from '@/lib/i18n'

const readSrc = (relativePath: string) =>
  readFileSync(fileURLToPath(new URL(relativePath, import.meta.url)), 'utf-8')

describe('BatchActionBar', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    setLocale('en')
  })

  it('renders selected count and emits actions for file-only selection', async () => {
    const wrapper = mount(BatchActionBar, {
      props: {
        selectedCount: 3,
        selectedFileCount: 3,
        selectedFolderCount: 0,
        totalCount: 10,
      },
    })

    expect(wrapper.text()).toContain('3 selected')

    const closeBtn = wrapper.find('button[aria-label="Close selection"]')
    expect(closeBtn.exists()).toBe(true)
    await closeBtn.trigger('click')
    expect(wrapper.emitted('close')).toBeTruthy()

    const deleteBtn = wrapper.find('button[aria-label="Move to trash"]')
    await deleteBtn.trigger('click')
    expect(wrapper.emitted('delete')).toBeTruthy()

    const downloadBtn = wrapper.find('button[aria-label="Download selected"]')
    expect(downloadBtn.attributes('disabled')).toBeUndefined()
    await downloadBtn.trigger('click')
    expect(wrapper.emitted('download')).toBeTruthy()

    const moveBtn = wrapper.find('button[aria-label="Move selected"]')
    await moveBtn.trigger('click')
    expect(wrapper.emitted('move')).toBeTruthy()

    const favBtn = wrapper.find('button[aria-label="Add to favorites"]')
    expect(favBtn.attributes('disabled')).toBeUndefined()
    await favBtn.trigger('click')
    expect(wrapper.emitted('favorite')).toBeTruthy()

    const vaultBtn = wrapper.find('button[aria-label="Move to personal vault"]')
    expect(vaultBtn.attributes('disabled')).toBeUndefined()

    const selectAllBtn = wrapper.find('.btn-text')
    expect(selectAllBtn.text()).toBe('Select all')
    await selectAllBtn.trigger('click')
    expect(wrapper.emitted('select-all')).toBeTruthy()

    await wrapper.setProps({ selectedCount: 10, selectedFileCount: 7, selectedFolderCount: 3, totalCount: 10 })
    const deselectAllBtn = wrapper.find('.btn-text')
    expect(deselectAllBtn.text()).toBe('Deselect all')
    await deselectAllBtn.trigger('click')
    expect(wrapper.emitted('clear-selection')).toBeTruthy()
  })

  it('disables file-only actions for folder-only selection with distinct accessible names', () => {
    const wrapper = mount(BatchActionBar, {
      props: {
        selectedCount: 2,
        selectedFileCount: 0,
        selectedFolderCount: 2,
        totalCount: 10,
      },
    })

    expect(wrapper.text()).toContain('2 selected')

    const disabledLabels = [
      'Download — files only',
      'Add to favorites — files only',
      'Move to Vault — files only',
    ]

    disabledLabels.forEach((label) => {
      const btn = wrapper.find(`button[aria-label="${label}"]`)
      expect(btn.exists()).toBe(true)
      expect(btn.attributes('disabled')).toBeDefined()
    })

    const moveBtn = wrapper.find('button[aria-label="Move selected"]')
    expect(moveBtn.attributes('disabled')).toBeUndefined()

    const deleteBtn = wrapper.find('button[aria-label="Move to trash"]')
    expect(deleteBtn.attributes('disabled')).toBeUndefined()

    expect(wrapper.find('button[aria-label="More actions"]').exists()).toBe(false)
  })

  it('scopes file-only action labels for mixed selection', () => {
    const wrapper = mount(BatchActionBar, {
      props: {
        selectedCount: 5,
        selectedFileCount: 2,
        selectedFolderCount: 3,
        totalCount: 10,
      },
    })

    expect(wrapper.text()).toContain('5 selected')
    expect(wrapper.text()).toContain('2 files, 3 folders')

    const downloadBtn = wrapper.find('button[aria-label="Download 2 files"]')
    expect(downloadBtn.exists()).toBe(true)
    expect(downloadBtn.attributes('disabled')).toBeUndefined()

    const favBtn = wrapper.find('button[aria-label="Favorite 2 files"]')
    expect(favBtn.exists()).toBe(true)
    expect(favBtn.attributes('disabled')).toBeUndefined()

    const vaultBtn = wrapper.find('button[aria-label="Move 2 files to vault"]')
    expect(vaultBtn.exists()).toBe(true)
    expect(vaultBtn.attributes('disabled')).toBeUndefined()

    expect(wrapper.find('button[aria-label="More actions"]').exists()).toBe(true)
  })

  it('disables all action buttons when selected count is 0', () => {
    const wrapper = mount(BatchActionBar, {
      props: {
        selectedCount: 0,
        selectedFileCount: 0,
        selectedFolderCount: 0,
        totalCount: 10,
      },
    })

    const disabledLabels = [
      'Download — files only',
      'Add to favorites — files only',
      'Move to Vault — files only',
    ]

    disabledLabels.forEach((label) => {
      const btn = wrapper.find(`button[aria-label="${label}"]`)
      expect(btn.exists()).toBe(true)
      expect(btn.attributes('disabled')).toBeDefined()
    })

    const moveBtn = wrapper.find('button[aria-label="Move selected"]')
    expect(moveBtn.attributes('disabled')).toBeDefined()

    const deleteBtn = wrapper.find('button[aria-label="Move to trash"]')
    expect(deleteBtn.attributes('disabled')).toBeDefined()
  })

  it('uses distinct disabled labels in Vietnamese', () => {
    setLocale('vi')
    const wrapper = mount(BatchActionBar, {
      props: {
        selectedCount: 1,
        selectedFileCount: 0,
        selectedFolderCount: 1,
        totalCount: 5,
      },
    })

    const disabledLabels = [
      'Tải xuống — chỉ áp dụng cho tệp',
      'Thêm vào yêu thích — chỉ áp dụng cho tệp',
      'Chuyển vào kho cá nhân — chỉ áp dụng cho tệp',
    ]

    disabledLabels.forEach((label) => {
      expect(wrapper.find(`button[aria-label="${label}"]`).exists()).toBe(true)
    })
  })

  it('closes selection on Escape', async () => {
    const wrapper = mount(BatchActionBar, {
      props: {
        selectedCount: 1,
        selectedFileCount: 1,
        selectedFolderCount: 0,
        totalCount: 5,
      },
    })

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'Escape' }))
    await wrapper.vm.$nextTick()
    expect(wrapper.emitted('close')).toBeTruthy()
  })
})

describe('BatchActionBar placement contract', () => {
  const bar = readSrc('./BatchActionBar.vue')

  it('does not apply bottom-nav offset in the base rule', () => {
    const baseBlock = bar.match(/\.batch-bar\s*\{[^}]*\}/)?.[0] ?? ''
    expect(baseBlock).toMatch(/bottom:\s*0/)
    expect(baseBlock).not.toContain('--bottom-nav-h')
    expect(baseBlock).not.toMatch(/env\(safe-area-inset-bottom\)/)
  })

  it('offsets above bottom nav and safe-area only on mobile (max-width: 767px)', () => {
    const mobileBlock =
      bar.match(/@media\s*\(\s*max-width:\s*767px\s*\)[\s\S]*?(?=@media)/)?.[0] ?? ''
    expect(mobileBlock).toMatch(
      /bottom:\s*calc\(var\(--bottom-nav-h\)\s*\+\s*env\(safe-area-inset-bottom\)\)/,
    )
    expect(mobileBlock).toMatch(/padding-bottom:\s*calc\(var\(--space-xs\)\s*\+\s*env\(safe-area-inset-bottom\)\)/)
  })

  it('uses floating tablet/desktop placement without bottom-nav height', () => {
    const tabletBlock =
      bar.match(/@media\s*\(\s*min-width:\s*768px\s*\)[\s\S]*?(?=\.batch-left)/)?.[0] ?? ''
    expect(tabletBlock).toContain('bottom: var(--space-lg)')
    expect(tabletBlock).not.toContain('--bottom-nav-h')
    expect(tabletBlock).not.toMatch(/env\(safe-area-inset-bottom\)/)
  })
})
