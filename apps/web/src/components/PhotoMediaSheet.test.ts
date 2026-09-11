import { mount } from '@vue/test-utils'
import { afterAll, beforeEach, describe, expect, it } from 'vitest'
import { setLocale } from '@/lib/i18n'
import PhotoMediaSheet from './PhotoMediaSheet.vue'

const bottomSheetStub = {
  props: ['open', 'title'],
  emits: ['close', 'after-leave'],
  template: '<section v-if="open"><slot /></section>',
}

const iconStub = { template: '<span data-testid="icon-stub" />' }
const originalLanguage = typeof navigator !== 'undefined' ? navigator.language : ''

function mountSheet(favorited = false) {
  return mount(PhotoMediaSheet, {
    props: { open: true, name: 'sunset.jpg', favorited },
    global: {
      stubs: {
        BottomSheet: bottomSheetStub,
        AppIcon: iconStub,
      },
    },
  })
}

describe('PhotoMediaSheet', () => {
  beforeEach(() => {
    setLocale('vi')
  })

  afterAll(() => {
    if (originalLanguage.toLowerCase().startsWith('en')) setLocale('en')
    else setLocale('vi')
  })

  it('localizes actions and reacts to locale changes without remounting', async () => {
    const wrapper = mountSheet()

    expect(wrapper.text()).toContain('Xem trước')
    expect(wrapper.text()).toContain('Tải xuống')
    expect(wrapper.text()).toContain('Thêm vào yêu thích')
    expect(wrapper.text()).toContain('Đóng')

    setLocale('en')
    await wrapper.vm.$nextTick()

    expect(wrapper.text()).toContain('Preview')
    expect(wrapper.text()).toContain('Download')
    expect(wrapper.text()).toContain('Add to favorites')
    expect(wrapper.text()).toContain('Close')
  })

  it('uses remove-favorite copy when the media is already favorited', async () => {
    const wrapper = mountSheet(true)

    expect(wrapper.text()).toContain('Bỏ khỏi yêu thích')

    setLocale('en')
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toContain('Remove from favorites')
  })

  it('preserves existing media action emits', async () => {
    const wrapper = mountSheet()
    const buttons = wrapper.findAll('button')

    await buttons[0]?.trigger('click')
    await buttons[1]?.trigger('click')
    await buttons[2]?.trigger('click')
    await buttons[3]?.trigger('click')

    expect(wrapper.emitted('view')).toHaveLength(1)
    expect(wrapper.emitted('download')).toHaveLength(1)
    expect(wrapper.emitted('favorite')).toHaveLength(1)
    expect(wrapper.emitted('close')).toHaveLength(1)
  })
})
