/**
 * @vitest-environment jsdom
 */
import { describe, expect, it } from 'vitest'
import router from './index'

describe('not-found route recovery', () => {
  it('resolves unknown app URLs to the intentional not-found screen instead of redirecting home', () => {
    const route = router.resolve('/missing/deep/link?from=stale-bookmark')

    expect(route.name).toBe('not-found')
    expect(route.fullPath).toBe('/missing/deep/link?from=stale-bookmark')
    expect(route.redirectedFrom).toBeUndefined()
  })

  it('keeps the public-share route distinct from the catch-all', () => {
    expect(router.resolve('/s/example-token').name).toBe('public-share')
  })
})
