export function formatVaultFileCount(count: number, locale: string): string {
  if (locale === 'en') {
    return `${count} ${count === 1 ? 'file' : 'files'}`
  }

  return `${count} tệp`
}
