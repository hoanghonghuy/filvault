import { afterEach, describe, expect, it, vi } from 'vitest'
import { generateUUID } from './uuid'

describe('generateUUID', () => {
  const uuidV4Regex = /^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i

  afterEach(() => {
    vi.unstubAllGlobals()
  })

  it('generates a valid RFC 4122 v4 UUID with standard Web Crypto', () => {
    const id = generateUUID()
    expect(id).toMatch(uuidV4Regex)
  })

  it('generates distinct UUIDs on subsequent calls', () => {
    const id1 = generateUUID()
    const id2 = generateUUID()
    expect(id1).not.toBe(id2)
  })

  it('falls back to crypto.getRandomValues when crypto.randomUUID is undefined (insecure HTTP context)', () => {
    const originalCrypto = globalThis.crypto
    vi.stubGlobal('crypto', {
      getRandomValues: (arr: Uint8Array) => originalCrypto.getRandomValues(arr),
      // randomUUID is undefined on insecure HTTP on mobile
    })

    const id = generateUUID()
    expect(id).toMatch(uuidV4Regex)
  })

  it('falls back to Math.random when Web Crypto is completely unavailable', () => {
    vi.stubGlobal('crypto', undefined)

    const id = generateUUID()
    expect(id).toMatch(uuidV4Regex)
  })
})
