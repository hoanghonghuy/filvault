/**
 * @vitest-environment jsdom
 */
import { beforeEach, describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import BatchActionBar from './BatchActionBar.vue'
import { setLocale } from '@/lib/i18n'

describe('BatchActionBar', () => {
  beforeEach(() => {
    setLocale('en')
  })
  it('renders selected count and emits actions', async () => {
    const wrapper = mount(BatchActionBar, {
      props: { selectedCount: 3, totalCount: 10 },
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
    await downloadBtn.trigger('click')
    expect(wrapper.emitted('download')).toBeTruthy()

    const moveBtn = wrapper.find('button[aria-label="Move selected"]')
    await moveBtn.trigger('click')
    expect(wrapper.emitted('move')).toBeTruthy()

    const favBtn = wrapper.find('button[aria-label="Add to favorites"]')
    await favBtn.trigger('click')
    expect(wrapper.emitted('favorite')).toBeTruthy()

    const selectAllBtn = wrapper.find('.btn-text')
    expect(selectAllBtn.text()).toBe('Select all')
    await selectAllBtn.trigger('click')
    expect(wrapper.emitted('select-all')).toBeTruthy()

    await wrapper.setProps({ selectedCount: 10, totalCount: 10 })
    const deselectAllBtn = wrapper.find('.btn-text')
    expect(deselectAllBtn.text()).toBe('Deselect all')
    await deselectAllBtn.trigger('click')
    expect(wrapper.emitted('clear-selection')).toBeTruthy()
  })

  it('disables action buttons when selected count is 0', () => {
    const wrapper = mount(BatchActionBar, {
      props: { selectedCount: 0, totalCount: 10 },
    })
    
    const actionButtons = [
      'Download selected',
      'Add to favorites',
      'Move selected',
      'Move to trash',
    ]

    actionButtons.forEach((label) => {
      const btn = wrapper.find(`button[aria-label="${label}"]`)
      expect(btn.attributes('disabled')).toBeDefined()
    })
  })
})
