/**
 * Generates an RFC 4122 version 4 UUID.
 *
 * Compatible with insecure contexts (such as testing mobile devices over LAN HTTP)
 * where window.crypto.randomUUID is undefined per W3C Web Cryptography spec.
 */
export function generateUUID(): string {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }

  if (typeof crypto !== 'undefined' && typeof crypto.getRandomValues === 'function') {
    const bytes = new Uint8Array(16)
    crypto.getRandomValues(bytes)
    // UUID v4: bits 12-15 of time_hi_and_version set to 0100
    bytes[6] = (bytes[6]! & 0x0f) | 0x40
    // UUID variant: bits 6-7 of clock_seq_hi_and_reserved set to 10
    bytes[8] = (bytes[8]! & 0x3f) | 0x80
    const hex = Array.from(bytes, (b) => b.toString(16).padStart(2, '0')).join('')
    return `${hex.slice(0, 8)}-${hex.slice(8, 12)}-${hex.slice(12, 16)}-${hex.slice(16, 20)}-${hex.slice(20)}`
  }

  // Fallback for environments where Web Crypto is completely unavailable
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0
    const v = c === 'x' ? r : (r & 0x3) | 0x8
    return v.toString(16)
  })
}
