import type { BrowserFile, TimelineGroup, TimelineItem } from '@/api/types'

export const HOME_PATH = '/'

export const RECENT_FILE_LIMIT = 5
export const RECENT_PHOTO_LIMIT = 6

export const SHELL_NAV = [
  { to: '/files', label: 'Files', shortLabel: 'Files', icon: 'folder' },
  { to: '/photos', label: 'Photos', shortLabel: 'Photos', icon: 'photos' },
  { to: '/chat', label: 'Chat', shortLabel: 'Chat', icon: 'chat' },
  { to: '/trash', label: 'Trash', shortLabel: 'Trash', icon: 'trash' },
  { to: '/settings', label: 'Settings', shortLabel: 'Settings', icon: 'settings' },
] as const

/** Home shortcuts: product surfaces only — Trash/Settings stay in persistent nav. */
export const OVERVIEW_DESTINATIONS = [
  { to: '/files', label: 'My Files', hint: 'Browse and upload', icon: 'folder' },
  { to: '/photos', label: 'Photos', hint: 'Timeline and albums', icon: 'photos' },
] as const

export function pageTitleForRoute(path: string, routeName?: string | symbol | null): string {
  if (path === '/' || path === '') return 'Overview'
  if (routeName === 'album') return 'Album'
  if (path === '/files' || path.startsWith('/files/')) return 'My Files'
  if (path === '/photos' || path.startsWith('/photos/')) return 'Photos'
  if (path === '/chat' || path.startsWith('/chat/')) return 'Chat'
  if (path === '/trash') return 'Trash'
  if (path === '/settings') return 'Settings'
  return 'Filvault'
}

/** Overview already shows quota in the hero — skip the compact storage bar there. */
export function showStorageBar(path: string): boolean {
  return path !== HOME_PATH && path !== ''
}

export function recentFilesFromBrowser(
  files: BrowserFile[],
  limit = RECENT_FILE_LIMIT,
): BrowserFile[] {
  return [...files].sort((a, b) => b.updatedAt.localeCompare(a.updatedAt)).slice(0, limit)
}

export function recentPhotosFromTimeline(
  groups: TimelineGroup[],
  limit = RECENT_PHOTO_LIMIT,
): TimelineItem[] {
  return groups.flatMap((group) => group.items).slice(0, limit)
}
