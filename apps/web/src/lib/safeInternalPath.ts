/** Same-app path only. Rejects protocol-relative and absolute URLs. */
export function safeInternalPath(raw: unknown, fallback = '/files'): string {
  if (typeof raw !== 'string' || raw.length === 0) {
    return fallback
  }
  if (!raw.startsWith('/') || raw.startsWith('//') || raw.startsWith('/\\')) {
    return fallback
  }
  if (raw.includes('://')) {
    return fallback
  }
  return raw
}
