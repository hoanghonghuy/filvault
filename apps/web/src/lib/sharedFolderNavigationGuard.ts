export function createSharedFolderNavigationGuard() {
  let generation = 0

  function beginNavigation(): number {
    generation += 1
    return generation
  }

  function invalidateNavigation(): void {
    generation += 1
  }

  function isCurrentNavigation(token: number): boolean {
    return token === generation
  }

  return {
    beginNavigation,
    invalidateNavigation,
    isCurrentNavigation,
  }
}
