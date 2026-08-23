import { describe, expect, it } from 'vitest'
import {
  HOME_PATH,
  OVERVIEW_DESTINATIONS,
  PROFILE_PATH,
  pageTitleForRoute,
  recentFilesFromBrowser,
  recentPhotosFromTimeline,
  showStorageBar,
} from './shellNav'
import type { BrowserFile, TimelineGroup } from '@/api/types'

describe('HOME_PATH', () => {
  it('is the app overview', () => {
    expect(HOME_PATH).toBe('/')
  })
})

describe('PROFILE_PATH', () => {
  it('is a dedicated account route outside bottom nav', () => {
    expect(PROFILE_PATH).toBe('/profile')
  })
})

describe('OVERVIEW_DESTINATIONS', () => {
  it('exposes Files, Photos and Shared with me', () => {
    expect(OVERVIEW_DESTINATIONS.map((item) => item.to)).toEqual(['/files', '/photos', '/shared'])
  })
})

describe('showStorageBar', () => {
  it('hides on Overview where quota already appears in the hero', () => {
    expect(showStorageBar('/')).toBe(false)
    expect(showStorageBar('')).toBe(false)
  })

  it('shows on shell destinations that do not repeat quota', () => {
    expect(showStorageBar('/files')).toBe(true)
    expect(showStorageBar('/photos')).toBe(true)
    expect(showStorageBar('/photos/albums/01ABC')).toBe(true)
    expect(showStorageBar('/trash')).toBe(true)
    expect(showStorageBar('/shared')).toBe(true)
    expect(showStorageBar('/settings')).toBe(true)
  })
})

describe('pageTitleForRoute', () => {
  it('uses Overview on home', () => {
    expect(pageTitleForRoute('/')).toBe('Overview')
  })

  it('uses My Files on the files browser', () => {
    expect(pageTitleForRoute('/files')).toBe('My Files')
    expect(pageTitleForRoute('/files', 'files')).toBe('My Files')
  })

  it('keeps Album when viewing an album', () => {
    expect(pageTitleForRoute('/photos/albums/01ABC', 'album')).toBe('Album')
  })

  it('maps the other shell destinations', () => {
    expect(pageTitleForRoute('/photos')).toBe('Photos')
    expect(pageTitleForRoute('/trash')).toBe('Trash')
    expect(pageTitleForRoute('/shared')).toBe('Shared with me')
    expect(pageTitleForRoute('/settings')).toBe('Settings')
    expect(pageTitleForRoute('/profile')).toBe('Profile')
  })
})

describe('recentFilesFromBrowser', () => {
  it('sorts by updatedAt descending and caps the list', () => {
    const files: BrowserFile[] = [
      { id: '1', name: 'old.pdf', mimeType: 'application/pdf', sizeBytes: 1, updatedAt: '2026-01-01T00:00:00Z' },
      { id: '2', name: 'new.txt', mimeType: 'text/plain', sizeBytes: 1, updatedAt: '2026-08-01T00:00:00Z' },
      { id: '3', name: 'mid.zip', mimeType: 'application/zip', sizeBytes: 1, updatedAt: '2026-06-01T00:00:00Z' },
    ]
    const recent = recentFilesFromBrowser(files, 2)
    expect(recent.map((f) => f.id)).toEqual(['2', '3'])
  })
})

describe('recentPhotosFromTimeline', () => {
  it('flattens groups in order and caps the list', () => {
    const groups: TimelineGroup[] = [
      {
        date: '2026-08-18',
        items: [
          { id: 'a', name: 'a.jpg', mimeType: 'image/jpeg', sizeBytes: 1, createdAt: '2026-08-18T10:00:00Z' },
          { id: 'b', name: 'b.jpg', mimeType: 'image/jpeg', sizeBytes: 1, createdAt: '2026-08-18T09:00:00Z' },
        ],
      },
      {
        date: '2026-08-17',
        items: [{ id: 'c', name: 'c.mp4', mimeType: 'video/mp4', sizeBytes: 1, createdAt: '2026-08-17T08:00:00Z' }],
      },
    ]
    expect(recentPhotosFromTimeline(groups, 2).map((p) => p.id)).toEqual(['a', 'b'])
  })
})
