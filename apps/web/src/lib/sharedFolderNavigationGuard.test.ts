import { describe, expect, it } from 'vitest'
import { createSharedFolderNavigationGuard } from './sharedFolderNavigationGuard'

describe('sharedFolderNavigationGuard', () => {
  it('tracks the latest navigation token across begin and invalidate', () => {
    const guard = createSharedFolderNavigationGuard()

    const first = guard.beginNavigation()
    expect(guard.isCurrentNavigation(first)).toBe(true)

    guard.invalidateNavigation()
    expect(guard.isCurrentNavigation(first)).toBe(false)

    const second = guard.beginNavigation()
    expect(guard.isCurrentNavigation(first)).toBe(false)
    expect(guard.isCurrentNavigation(second)).toBe(true)
  })
})
