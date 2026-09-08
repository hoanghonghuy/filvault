import { describe, expect, it } from 'vitest'
import { normalizePresignedUrl } from './presignedUrl'

describe('normalizePresignedUrl', () => {
  it('rewrites localhost:9002 presigned URLs to current origin', () => {
    const origin = 'https://tunnel.filvault.com'
    const input = 'http://localhost:9002/filvault/user123/photo.jpg?X-Amz-Signature=abc'
    const result = normalizePresignedUrl(input, origin)
    expect(result).toBe(`${origin}/filvault/user123/photo.jpg?X-Amz-Signature=abc`)
  })

  it('rewrites 127.0.0.1:9002 presigned URLs to current origin', () => {
    const origin = 'https://tunnel.filvault.com'
    const input = 'http://127.0.0.1:9002/filvault/chat/item.png'
    const result = normalizePresignedUrl(input, origin)
    expect(result).toBe(`${origin}/filvault/chat/item.png`)
  })

  it('preserves query params, hash and paths intact', () => {
    const origin = 'https://tunnel.filvault.com'
    const input = 'http://localhost:9002/filvault/bucket/file?a=1&b=2#hash'
    const result = normalizePresignedUrl(input, origin)
    expect(result).toBe(`${origin}/filvault/bucket/file?a=1&b=2#hash`)
  })

  it('leaves already normalized or external URLs alone', () => {
    expect(normalizePresignedUrl('https://example.com/image.png')).toBe('https://example.com/image.png')
    expect(normalizePresignedUrl('')).toBe('')
  })
})
