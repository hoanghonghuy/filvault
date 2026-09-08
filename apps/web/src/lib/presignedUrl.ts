/**
 * Normalizes presigned S3 upload/download URLs so they route through
 * the current web origin (via nginx or dev proxy to MinIO),
 * ensuring uploads and downloads work over Cloudflare Tunnel, mobile LAN, and HTTPS without Mixed Content.
 */
export function normalizePresignedUrl(url: string, originOverride?: string): string {
  if (!url) return url
  try {
    const origin =
      originOverride ||
      (typeof window !== 'undefined' && window.location && window.location.origin !== 'null'
        ? window.location.origin
        : '')
    const base = origin || 'http://localhost:5173'
    const parsed = new URL(url, base)
    if (
      parsed.port === '9002' ||
      parsed.hostname === 'minio' ||
      ((parsed.hostname === 'localhost' || parsed.hostname === '127.0.0.1') && parsed.pathname.startsWith('/filvault'))
    ) {
      if (origin) {
        return `${origin}${parsed.pathname}${parsed.search}${parsed.hash}`
      }
      return `${parsed.pathname}${parsed.search}${parsed.hash}`
    }
  } catch {
    // Return original url if parsing fails
  }
  return url
}
