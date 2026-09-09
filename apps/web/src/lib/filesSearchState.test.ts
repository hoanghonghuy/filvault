import { describe, expect, it } from 'vitest'
import {
  buildFilesRouteQuery,
  buildSearchApiQueryString,
  DEFAULT_SEARCH_FILTERS,
  getActiveFilterChipKeys,
  getEmptySearchCause,
  hasActiveSearchFilters,
  isValidDateRange,
  parseSearchFiltersFromRoute,
  parseSearchQueryFromRoute,
  removeSearchFilter,
  resetSearchFilters,
  routeQueriesEqual,
} from './filesSearchState'

describe('isValidDateRange', () => {
  it('accepts open-ended ranges', () => {
    expect(isValidDateRange('2026-01-01')).toBe(true)
    expect(isValidDateRange(undefined, '2026-12-31')).toBe(true)
  })

  it('accepts from <= to', () => {
    expect(isValidDateRange('2026-01-01', '2026-12-31')).toBe(true)
    expect(isValidDateRange('2026-06-01', '2026-06-01')).toBe(true)
  })

  it('rejects from > to', () => {
    expect(isValidDateRange('2026-12-31', '2026-01-01')).toBe(false)
  })
})

describe('parseSearchFiltersFromRoute', () => {
  it('returns defaults for an empty query', () => {
    expect(parseSearchFiltersFromRoute({})).toEqual(DEFAULT_SEARCH_FILTERS)
  })

  it('parses supported filter params', () => {
    expect(
      parseSearchFiltersFromRoute({
        type: 'image',
        sort: 'name',
        order: 'asc',
        from: '2026-01-01',
        to: '2026-12-31',
        sfolderId: 'folder-1',
      }),
    ).toEqual({
      type: 'image',
      sort: 'name',
      order: 'asc',
      from: '2026-01-01',
      to: '2026-12-31',
      folderId: 'folder-1',
    })
  })

  it('ignores invalid enum values', () => {
    expect(parseSearchFiltersFromRoute({ type: 'bogus', sort: 'nope' })).toEqual(DEFAULT_SEARCH_FILTERS)
  })
})

describe('parseSearchQueryFromRoute', () => {
  it('reads q from the route', () => {
    expect(parseSearchQueryFromRoute({ q: 'invoice' })).toBe('invoice')
    expect(parseSearchQueryFromRoute({})).toBe('')
  })
})

describe('buildSearchApiQueryString', () => {
  it('includes only non-default filters', () => {
    const qs = buildSearchApiQueryString('report', { ...DEFAULT_SEARCH_FILTERS, type: 'document' })
    expect(qs).toBe('q=report&type=document')
  })
})

describe('buildFilesRouteQuery', () => {
  it('preserves browse params while syncing search state', () => {
    const next = buildFilesRouteQuery(
      { folderId: 'browse-1', view: 'all' },
      'notes',
      { ...DEFAULT_SEARCH_FILTERS, type: 'document', from: '2026-01-01' },
    )
    expect(next).toEqual({
      folderId: 'browse-1',
      view: 'all',
      q: 'notes',
      type: 'document',
      from: '2026-01-01',
    })
  })

  it('clears stale filter params when reset', () => {
    const next = buildFilesRouteQuery(
      { q: 'old', type: 'image', sort: 'name', order: 'asc' },
      '',
      DEFAULT_SEARCH_FILTERS,
    )
    expect(next).toEqual({})
  })
})

describe('routeQueriesEqual', () => {
  it('compares query objects by value', () => {
    expect(routeQueriesEqual({ q: 'a' }, { q: 'a' })).toBe(true)
    expect(routeQueriesEqual({ q: 'a' }, { q: 'b' })).toBe(false)
  })
})

describe('filter chips', () => {
  it('lists one chip per active filter', () => {
    expect(
      getActiveFilterChipKeys({
        type: 'video',
        sort: 'name',
        order: 'asc',
        from: '2026-01-01',
        to: '2026-12-31',
        folderId: 'f1',
      }),
    ).toEqual(['type', 'folderId', 'from', 'to', 'sort', 'order'])
  })

  it('shows order chip for relevance+asc (active order independent of sort)', () => {
    expect(
      getActiveFilterChipKeys({ ...DEFAULT_SEARCH_FILTERS, order: 'asc' }),
    ).toEqual(['order'])
    expect(hasActiveSearchFilters({ ...DEFAULT_SEARCH_FILTERS, order: 'asc' })).toBe(true)
  })

  it('omits order chip for relevance+desc defaults', () => {
    expect(getActiveFilterChipKeys(DEFAULT_SEARCH_FILTERS)).toEqual([])
    expect(hasActiveSearchFilters(DEFAULT_SEARCH_FILTERS)).toBe(false)
  })

  it('shows sort and order chips for non-relevance+asc', () => {
    expect(
      getActiveFilterChipKeys({ ...DEFAULT_SEARCH_FILTERS, sort: 'name', order: 'asc' }),
    ).toEqual(['sort', 'order'])
  })

  it('removes individual filters predictably', () => {
    const base = {
      type: 'image' as const,
      sort: 'name' as const,
      order: 'asc' as const,
      from: '2026-01-01',
      to: '2026-12-31',
      folderId: 'f1',
    }
    expect(removeSearchFilter(base, 'type').type).toBe('all')
    expect(removeSearchFilter(base, 'from').from).toBeUndefined()
    expect(removeSearchFilter(base, 'sort')).toEqual({
      ...base,
      sort: 'relevance',
      order: 'desc',
    })
    expect(removeSearchFilter(base, 'order').order).toBe('desc')
  })

  it('clearing order chip resets to desc even when sort is relevance', () => {
    const filters = { ...DEFAULT_SEARCH_FILTERS, order: 'asc' as const }
    expect(removeSearchFilter(filters, 'order')).toEqual(DEFAULT_SEARCH_FILTERS)
    expect(getActiveFilterChipKeys(removeSearchFilter(filters, 'order'))).toEqual([])
  })
})

describe('resetSearchFilters', () => {
  it('returns defaults without touching search query', () => {
    expect(resetSearchFilters()).toEqual(DEFAULT_SEARCH_FILTERS)
    expect(hasActiveSearchFilters(resetSearchFilters())).toBe(false)
  })
})

describe('getEmptySearchCause', () => {
  it('classifies query-only, filter-only and combined emptiness', () => {
    expect(getEmptySearchCause('invoice', DEFAULT_SEARCH_FILTERS)).toBe('query')
    expect(getEmptySearchCause('', { ...DEFAULT_SEARCH_FILTERS, type: 'image' })).toBe('filters')
    expect(getEmptySearchCause('invoice', { ...DEFAULT_SEARCH_FILTERS, type: 'image' })).toBe('both')
  })
})
