import { describe, expect, it } from 'vitest'
import { shellAccessibilityCopy } from './shellAccessibilityCopy'

describe('shellAccessibilityCopy', () => {
  it('localizes shell navigation landmarks and home action', () => {
    const vi = shellAccessibilityCopy('vi')
    const en = shellAccessibilityCopy('en')

    expect(vi.mainNavigation).toBe('Điều hướng chính')
    expect(vi.destinations).toBe('Điểm đến')
    expect(vi.goToOverview).toBe('Đi tới tổng quan')
    expect(en.mainNavigation).toBe('Main navigation')
    expect(en.destinations).toBe('Destinations')
    expect(en.goToOverview).toBe('Go to overview')
  })

  it('keeps the user identifier while localizing the profile action', () => {
    expect(shellAccessibilityCopy('vi').openProfileFor('Huy')).toBe('Mở hồ sơ của Huy')
    expect(shellAccessibilityCopy('en').openProfileFor('Huy')).toBe('Open profile for Huy')
    expect(shellAccessibilityCopy('vi').accountFallback).toBe('tài khoản')
    expect(shellAccessibilityCopy('en').accountFallback).toBe('account')
  })
})
