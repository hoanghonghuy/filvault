import { describe, expect, it } from 'vitest'
import { cellDelay, EASE_ENTER, EASE_EXIT, EASE_STANDARD, MOTION_ENTER_MS, MOTION_EXIT_MS, MOTION_PRESS_MS } from './motion'

describe('motion tokens', () => {
  it('stays within DESIGN.md 150–250ms and exits faster than enter', () => {
    expect(MOTION_PRESS_MS).toBe(150)
    expect(MOTION_ENTER_MS).toBe(250)
    expect(MOTION_EXIT_MS).toBe(200)
    expect(MOTION_EXIT_MS).toBeLessThan(MOTION_ENTER_MS)
  })

  it('uses Material 3 Standard curves for web', () => {
    expect(EASE_STANDARD).toBe('cubic-bezier(0.2, 0, 0, 1)')
    expect(EASE_ENTER).toBe('cubic-bezier(0, 0, 0, 1)')
    expect(EASE_EXIT).toBe('cubic-bezier(0.3, 0, 1, 1)')
  })
})

describe('cellDelay', () => {
  it('staggers cells at a fixed step', () => {
    expect(cellDelay(0)).toBe('0ms')
    expect(cellDelay(1)).toBe('30ms')
    expect(cellDelay(2)).toBe('60ms')
  })

  it('caps the delay to keep the grid feeling snappy', () => {
    expect(cellDelay(50)).toBe('240ms')
    expect(cellDelay(200)).toBe('240ms')
  })
})
