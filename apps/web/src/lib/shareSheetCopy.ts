import type { Locale } from '@/lib/i18n'

export type ShareSheetCopy = {
  anyone: string
  copying: string
  copyLink: string
  expires: string
  neverExpires: string
  revoking: string
  revoke: string
  close: string
  expiresAfter: string
  expiryGroup: string
  forever: string
  oneHour: string
  oneDay: string
  sevenDays: string
  creating: string
  create: string
}

const copy: Record<Locale, ShareSheetCopy> = {
  vi: {
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
  },
  en: {
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
  },
}

export function shareSheetCopy(locale: Locale): ShareSheetCopy {
  return copy[locale]
}
