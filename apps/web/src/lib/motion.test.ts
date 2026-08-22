import { describe, expect, it } from 'vitest'
import { cellDelay } from './motion'

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
