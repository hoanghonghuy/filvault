import { describe, expect, it } from 'vitest'
import { shareSheetCopy } from './shareSheetCopy'

describe('shareSheetCopy', () => {
  it('provides Vietnamese public-link lifecycle labels', () => {
    expect(shareSheetCopy('vi')).toEqual({
      anyone: 'Bất kỳ ai có liên kết này đều có thể xem và tải xuống',
      copying: 'Đang sao chép liên kết…',
      copyLink: 'Sao chép liên kết',
      expires: 'Hết hạn',
      neverExpires: 'Không hết hạn',
      revoking: 'Đang thu hồi…',
      revoke: 'Thu hồi liên kết',
      close: 'Đóng',
      expiresAfter: 'Liên kết hết hạn sau',
      expiryGroup: 'Thời hạn liên kết',
      forever: 'Vĩnh viễn',
      oneHour: '1 giờ',
      oneDay: '24 giờ',
      sevenDays: '7 ngày',
      creating: 'Đang tạo…',
      create: 'Tạo liên kết',
    })
  })

  it('provides English public-link lifecycle labels', () => {
    expect(shareSheetCopy('en')).toEqual({
      anyone: 'Anyone with this link can view and download',
      copying: 'Copying link…',
      copyLink: 'Copy link',
      expires: 'Expires',
      neverExpires: 'Never expires',
      revoking: 'Revoking…',
      revoke: 'Revoke link',
      close: 'Close',
      expiresAfter: 'Link expires after',
      expiryGroup: 'Link expiry',
      forever: 'Forever',
      oneHour: '1 hour',
      oneDay: '24 hours',
      sevenDays: '7 days',
      creating: 'Creating…',
      create: 'Create link',
    })
  })
})
