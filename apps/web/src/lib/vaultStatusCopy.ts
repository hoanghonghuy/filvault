export type VaultStatusCopy = {
  loading: string
  loadFailed: string
}

const COPY: Record<'vi' | 'en', VaultStatusCopy> = {
  vi: {
    loading: 'Đang tải Kho cá nhân…',
    loadFailed: 'Không thể tải trạng thái Kho cá nhân. Hãy thử lại.',
  },
  en: {
    loading: 'Loading Personal Vault…',
    loadFailed: 'Could not load Personal Vault status. Please try again.',
  },
}

export function getVaultStatusCopy(locale: string): VaultStatusCopy {
  return COPY[locale === 'vi' ? 'vi' : 'en']
}
