export function formatSearchResultCount(count: number, locale: string): string {
  if (locale === 'en') {
    return `${count} ${count === 1 ? 'result' : 'results'}`
  }

  return `${count} kết quả`
}
