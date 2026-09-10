import { describe, expect, it } from 'vitest'
import { trashRetentionNotice } from './trashRetention'

describe('trashRetentionNotice', () => {
  it('uses the configured retention period when automatic deletion is enabled', () => {
    expect(
      trashRetentionNotice(
        { trashAutoDeleteEnabled: true, trashRetentionDays: 45 },
        'en',
      ),
    ).toContain('45 days')
    expect(
      trashRetentionNotice(
        { trashAutoDeleteEnabled: true, trashRetentionDays: 45 },
        'vi',
      ),
    ).toContain('45 ngày')
  })

  it('does not imply a countdown when automatic deletion is disabled', () => {
    const english = trashRetentionNotice(
      { trashAutoDeleteEnabled: false, trashRetentionDays: 30 },
      'en',
    )
    const vietnamese = trashRetentionNotice(
      { trashAutoDeleteEnabled: false, trashRetentionDays: 30 },
      'vi',
    )

    expect(english).toContain('is off')
    expect(english).not.toContain('30 days')
    expect(vietnamese).toContain('đang tắt')
    expect(vietnamese).not.toContain('30 ngày')
  })

  it('uses a neutral unavailable state instead of inventing policy before account hydration', () => {
    expect(trashRetentionNotice(null, 'en')).toContain('unavailable')
    expect(trashRetentionNotice(undefined, 'vi')).toContain('Chưa thể xác định')
  })
})
