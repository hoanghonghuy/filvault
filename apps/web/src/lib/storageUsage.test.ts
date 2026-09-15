import { describe, expect, it } from 'vitest'
import {
  STORAGE_NEAR_QUOTA_PERCENT,
  storageFillRatio,
  storageProgressAria,
  storageToneClass,
  storageUsagePercent,
  storageUsageState,
} from './storageUsage'

describe('storageUsage', () => {
  const quota = 1000

  it('classifies normal usage below the near-quota threshold', () => {
    expect(storageUsageState(500, quota)).toBe('normal')
    expect(storageUsagePercent(500, quota)).toBe(50)
    expect(storageToneClass('normal')).toBe('')
  })

  it('classifies near-quota at and above 90%', () => {
    const used = Math.ceil((STORAGE_NEAR_QUOTA_PERCENT / 100) * quota)
    expect(storageUsageState(used, quota)).toBe('near-quota')
    expect(storageUsagePercent(used, quota)).toBeGreaterThanOrEqual(STORAGE_NEAR_QUOTA_PERCENT)
    expect(storageToneClass('near-quota')).toBe('warning')
  })

  it('classifies full quota at and above 100%', () => {
    expect(storageUsageState(quota, quota)).toBe('full-quota')
    expect(storageUsageState(quota + 50, quota)).toBe('full-quota')
    expect(storageUsagePercent(quota, quota)).toBe(100)
    expect(storageToneClass('full-quota')).toBe('danger')
  })

  it('returns usage-unavailable when flagged unavailable', () => {
    expect(storageUsageState(0, quota, { unavailable: true })).toBe('usage-unavailable')
    expect(storageUsageState(900, quota, { unavailable: true })).toBe('usage-unavailable')
  })

  it('returns loading when flagged loading', () => {
    expect(storageUsageState(0, quota, { loading: true })).toBe('loading')
  })

  it('handles zero or invalid quota without NaN or Infinity', () => {
    expect(storageUsagePercent(100, 0)).toBe(0)
    expect(storageUsagePercent(100, -1)).toBe(0)
    expect(storageUsagePercent(100, Number.NaN)).toBe(0)
    expect(storageFillRatio(100, 0)).toBe(0)
    expect(storageFillRatio(100, Number.POSITIVE_INFINITY)).toBe(0)

    const aria = storageProgressAria(100, 0, (n) => `${n} B`, 'en')
    expect(aria.valuenow).toBe(0)
    expect(Number.isFinite(aria.valuenow)).toBe(true)
    expect(aria.label).toBe('100 B used')
    expect(aria.label).not.toMatch(/NaN|Infinity/)
  })

  it('localizes progress labels for valid and unavailable quota values', () => {
    expect(storageProgressAria(250, quota, (n) => `${n} B`, 'en').label).toBe(
      '250 B of 1000 B used',
    )
    expect(storageProgressAria(250, quota, (n) => `${n} B`, 'vi').label).toBe(
      'Đã dùng 250 B trên 1000 B',
    )
    expect(storageProgressAria(250, 0, (n) => `${n} B`, 'vi').label).toBe('Đã dùng 250 B')
  })

  it('clamps fill ratio and aria values to safe ranges', () => {
    expect(storageFillRatio(1500, quota)).toBe(1)
    expect(storageUsagePercent(1500, quota)).toBe(100)

    const aria = storageProgressAria(1500, quota, (n) => `${n}`, 'en')
    expect(aria.valuenow).toBe(100)
    expect(aria.valuemin).toBe(0)
    expect(aria.valuemax).toBe(100)
  })
})
