import type { BrowserFile, TimelineGroup, TimelineItem } from '@/api/types'

export const HOME_PATH = '/'
export const PROFILE_PATH = '/profile'

export const RECENT_FILE_LIMIT = 5
export const RECENT_PHOTO_LIMIT = 6

export const SHELL_NAV = [
  { to: '/', label: 'Home', shortLabel: 'Home', icon: 'home' },
  { to: '/files', label: 'Files', shortLabel: 'Files', icon: 'folder' },
  { to: '/photos', label: 'Photos', shortLabel: 'Photos', icon: 'photos' },
  { to: '/shared', label: 'Shared', shortLabel: 'Shared', icon: 'share' },
  { to: '/settings', label: 'Settings', shortLabel: 'Settings', icon: 'settings' },
] as const

/** Home shortcuts: product surfaces only — Trash/Settings stay in persistent nav. */
export const OVERVIEW_DESTINATIONS = [
  { to: '/files', label: 'My Files', hint: 'Browse and upload', icon: 'folder' },
  { to: '/photos', label: 'Photos', hint: 'Timeline and albums', icon: 'photos' },
  { to: '/shared', label: 'Shared with me', hint: 'Items shared to you', icon: 'users' },
  { to: '/chat', label: 'Chat', hint: 'Messages and video calls', icon: 'chat' },
] as const

export function pageTitleForRoute(
  path: string,
  routeName?: string | symbol | null,
  titles?: Partial<Record<'overview' | 'album' | 'files' | 'vault' | 'photos' | 'chat' | 'trash' | 'shared' | 'settings' | 'profile', string>>,
): string {
  if (path === '/' || path === '') return titles?.overview ?? 'Overview'
  if (routeName === 'album') return titles?.album ?? 'Album'
  if (path === '/files' || path.startsWith('/files/')) return titles?.files ?? 'My Files'
  if (path === '/vault' || path.startsWith('/vault/')) return titles?.vault ?? 'Personal Vault'
  if (path === '/photos' || path.startsWith('/photos/')) return titles?.photos ?? 'Photos'
  if (path === '/chat' || path.startsWith('/chat/')) return titles?.chat ?? 'Chat'
  if (path === '/trash') return titles?.trash ?? 'Trash'
  if (path === '/shared') return titles?.shared ?? 'Shared with me'
  if (path === '/settings') return titles?.settings ?? 'Settings'
  if (path === '/profile') return titles?.profile ?? 'Profile'
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
