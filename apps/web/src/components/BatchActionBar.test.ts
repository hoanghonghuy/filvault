/**
 * @vitest-environment jsdom
 */
import { beforeEach, describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import BatchActionBar from './BatchActionBar.vue'
import { setLocale } from '@/lib/i18n'

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

  it('disables file-only actions for folder-only selection', () => {
    const wrapper = mount(BatchActionBar, {
      props: {
        selectedCount: 2,
        selectedFileCount: 0,
        selectedFolderCount: 2,
        totalCount: 10,
      },
    })

    expect(wrapper.text()).toContain('2 selected')

    const fileOnlyBtns = wrapper.findAll('button[aria-label="Files only"]')
    expect(fileOnlyBtns.length).toBe(3)
    fileOnlyBtns.forEach((btn) => {
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

    const fileOnlyBtns = wrapper.findAll('button[aria-label="Files only"]')
    expect(fileOnlyBtns.length).toBe(3)
    fileOnlyBtns.forEach((btn) => {
      expect(btn.attributes('disabled')).toBeDefined()
    })

    const moveBtn = wrapper.find('button[aria-label="Move selected"]')
    expect(moveBtn.attributes('disabled')).toBeDefined()

    const deleteBtn = wrapper.find('button[aria-label="Move to trash"]')
    expect(deleteBtn.attributes('disabled')).toBeDefined()
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
