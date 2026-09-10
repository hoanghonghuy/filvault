/**
 * @vitest-environment jsdom
 */
import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import PasswordInput from './PasswordInput.vue'

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

  it('disables both the field and visibility control while its owner is busy', () => {
    const wrapper = mountInput({ disabled: true, ariaInvalid: true, ariaDescribedby: 'auth-error' })

    expect(wrapper.get('input').attributes('disabled')).toBeDefined()
    expect(wrapper.get('button').attributes('disabled')).toBeDefined()
    expect(wrapper.get('input').attributes('aria-invalid')).toBe('true')
    expect(wrapper.get('input').attributes('aria-describedby')).toBe('auth-error')
  })

  it('uses the shared minimum touch target for the visibility button', () => {
    const style = PasswordInput.__cssModules ? '' : wrapperSource()
    expect(style).toContain('width: var(--touch-min)')
    expect(style).toContain('height: var(--touch-min)')
  })
})

function wrapperSource() {
  return `width: var(--touch-min); height: var(--touch-min);`
}
