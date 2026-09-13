import type { Locale } from '@/lib/i18n'

export const STORAGE_NEAR_QUOTA_PERCENT = 90
export const STORAGE_FULL_QUOTA_PERCENT = 100

export type StorageUsageState =
  | 'loading'
  | 'normal'
  | 'near-quota'
  | 'full-quota'
  | 'usage-unavailable'

export type StorageToneClass = '' | 'warning' | 'danger'

export function isValidQuotaBytes(quotaBytes: number): boolean {
  return Number.isFinite(quotaBytes) && quotaBytes > 0
}

export function storageUsagePercent(usedBytes: number, quotaBytes: number): number {
  if (!isValidQuotaBytes(quotaBytes) || !Number.isFinite(usedBytes) || usedBytes < 0) {
    return 0
  }
  return Math.min(100, Math.max(0, Math.round((usedBytes / quotaBytes) * 100)))
}

export function storageFillRatio(usedBytes: number, quotaBytes: number): number {
  if (!isValidQuotaBytes(quotaBytes) || !Number.isFinite(usedBytes) || usedBytes < 0) {
    return 0
  }
  return Math.min(1, usedBytes / quotaBytes)
}

export function storageUsageState(
  usedBytes: number,
  quotaBytes: number,
  options?: { loading?: boolean; unavailable?: boolean },
): StorageUsageState {
  if (options?.loading) return 'loading'
  if (options?.unavailable) return 'usage-unavailable'

  const percent = storageUsagePercent(usedBytes, quotaBytes)
  if (percent >= STORAGE_FULL_QUOTA_PERCENT) return 'full-quota'
  if (percent >= STORAGE_NEAR_QUOTA_PERCENT) return 'near-quota'
  return 'normal'
}

export function storageToneClass(state: StorageUsageState): StorageToneClass {
  if (state === 'full-quota') return 'danger'
  if (state === 'near-quota') return 'warning'
  return ''
}

export function storageProgressAria(
  usedBytes: number,
  quotaBytes: number,
  formatBytes: (bytes: number) => string,
  locale: Locale = 'en',
): { valuemin: number; valuemax: number; valuenow: number; label: string } {
  const valuenow = storageUsagePercent(usedBytes, quotaBytes)
  const used = formatBytes(usedBytes)
  const label = isValidQuotaBytes(quotaBytes)
    ? locale === 'vi'
      ? `Đã dùng ${used} trên ${formatBytes(quotaBytes)}`
      : `${used} of ${formatBytes(quotaBytes)} used`
    : locale === 'vi'
      ? `Đã dùng ${used}`
      : `${used} used`

  return {
    valuemin: 0,
    valuemax: 100,
    valuenow,
    label,
  }
}
