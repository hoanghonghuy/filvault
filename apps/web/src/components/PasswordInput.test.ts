/**
 * @vitest-environment jsdom
 */
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { beforeEach, describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import PasswordInput from './PasswordInput.vue'
import { setLocale } from '@/lib/i18n'

const readSrc = (relativePath: string) =>
  readFileSync(fileURLToPath(new URL(relativePath, import.meta.url)), 'utf-8')

function mountInput(overrides: Record<string, unknown> = {}) {
  return mount(PasswordInput, {
    props: {
      id: 'password-field',
      name: 'password',
      autocomplete: 'current-password',
      modelValue: 'secret-value',
      required: true,
      minlength: 8,
      ...overrides,
    },
  })
}

describe('PasswordInput', () => {
  beforeEach(() => setLocale('en'))

  it('toggles visibility without changing the password value or input semantics', async () => {
    const wrapper = mountInput({ enterkeyhint: 'go' })
    const input = wrapper.get<HTMLInputElement>('input')
    const toggle = wrapper.get<HTMLButtonElement>('button')

    expect(input.attributes('type')).toBe('password')
    expect(input.element.value).toBe('secret-value')
    expect(input.attributes('name')).toBe('password')
    expect(input.attributes('autocomplete')).toBe('current-password')
    expect(input.attributes('enterkeyhint')).toBe('go')
    expect(input.attributes('required')).toBeDefined()
    expect(input.attributes('minlength')).toBe('8')
    expect(toggle.attributes('aria-label')).toBe('Show password')
    expect(toggle.attributes('aria-pressed')).toBe('false')

    await toggle.trigger('click')

    expect(input.attributes('type')).toBe('text')
    expect(input.element.value).toBe('secret-value')
    expect(toggle.attributes('aria-label')).toBe('Hide password')
    expect(toggle.attributes('aria-pressed')).toBe('true')
  })

  it('reacts to the active locale for visibility labels', async () => {
    setLocale('vi')
    const wrapper = mountInput()
    const toggle = wrapper.get<HTMLButtonElement>('button')

    expect(toggle.attributes('aria-label')).toBe('Hiện mật khẩu')
    await toggle.trigger('click')
    expect(toggle.attributes('aria-label')).toBe('Ẩn mật khẩu')

    setLocale('en')
    await wrapper.vm.$nextTick()
    expect(toggle.attributes('aria-label')).toBe('Hide password')
  })

  it('disables both the field and visibility control while its owner is busy', () => {
    const wrapper = mountInput({ disabled: true, ariaInvalid: true, ariaDescribedby: 'auth-error' })

    expect(wrapper.get('input').attributes('disabled')).toBeDefined()
    expect(wrapper.get('button').attributes('disabled')).toBeDefined()
    expect(wrapper.get('input').attributes('aria-invalid')).toBe('true')
    expect(wrapper.get('input').attributes('aria-describedby')).toBe('auth-error')
  })

  it('uses the shared minimum touch target for the visibility button', () => {
    const source = readSrc('./PasswordInput.vue')
    expect(source).toMatch(/\.password-toggle\s*\{[^}]*width:\s*var\(--touch-min\)/s)
    expect(source).toMatch(/\.password-toggle\s*\{[^}]*height:\s*var\(--touch-min\)/s)
  })
})
