import type { SearchFilters } from '@/api/types'
import type { LocationQuery } from 'vue-router'

export const DEFAULT_SEARCH_FILTERS: SearchFilters = {
  type: 'all',
  sort: 'relevance',
  order: 'desc',
}

export type FilesSearchChipKey = 'type' | 'folderId' | 'from' | 'to' | 'sort' | 'order'

export type RouteQueryRecord = LocationQuery

const FILTER_TYPES = new Set<SearchFilters['type']>([
  'all',
  'image',
  'video',
  'document',
  'archive',
  'folder',
])

const FILTER_SORTS = new Set<SearchFilters['sort']>(['relevance', 'name', 'date', 'size'])

const FILTER_ORDERS = new Set<SearchFilters['order']>(['asc', 'desc'])

/** Returns false when both dates are set and from is after to. */
export function isValidDateRange(from?: string, to?: string): boolean {
  if (!from || !to) return true
  return from <= to
}

export function parseSearchQueryFromRoute(query: RouteQueryRecord): string {
  const raw = query.q
  return typeof raw === 'string' ? raw : ''
}

export function parseSearchFiltersFromRoute(query: RouteQueryRecord): SearchFilters {
  const filters: SearchFilters = { ...DEFAULT_SEARCH_FILTERS }

  const type = query.type
  if (typeof type === 'string' && FILTER_TYPES.has(type as SearchFilters['type'])) {
    filters.type = type as SearchFilters['type']
  }

  const sort = query.sort
  if (typeof sort === 'string' && FILTER_SORTS.has(sort as SearchFilters['sort'])) {
    filters.sort = sort as SearchFilters['sort']
  }

  const order = query.order
  if (typeof order === 'string' && FILTER_ORDERS.has(order as SearchFilters['order'])) {
    filters.order = order as SearchFilters['order']
  }

  const from = query.from
  if (typeof from === 'string' && from) filters.from = from

  const to = query.to
  if (typeof to === 'string' && to) filters.to = to

  const scopedFolder = query.sfolderId
  if (typeof scopedFolder === 'string' && scopedFolder) filters.folderId = scopedFolder

  return filters
}

export function hasActiveSearchFilters(filters: SearchFilters): boolean {
  return (
    filters.type !== 'all' ||
    filters.sort !== 'relevance' ||
    filters.order !== 'desc' ||
    Boolean(filters.folderId) ||
    Boolean(filters.from) ||
    Boolean(filters.to)
  )
}

export function buildSearchApiQueryString(q: string, filters: SearchFilters): string {
  const params = new URLSearchParams()
  params.set('q', q.trim())
  if (filters.type !== 'all') params.set('type', filters.type)
  if (filters.folderId) params.set('folderId', filters.folderId)
  if (filters.from) params.set('from', filters.from)
  if (filters.to) params.set('to', filters.to)
  if (filters.sort !== 'relevance') params.set('sort', filters.sort)
  if (filters.order !== 'desc') params.set('order', filters.order)
  return params.toString()
}

const FILTER_QUERY_KEYS = ['q', 'type', 'from', 'to', 'sort', 'order', 'sfolderId'] as const

export function buildFilesRouteQuery(
  baseQuery: RouteQueryRecord,
  q: string,
  filters: SearchFilters,
): RouteQueryRecord {
  const next: RouteQueryRecord = { ...baseQuery }

  for (const key of FILTER_QUERY_KEYS) {
    delete next[key]
  }

  const trimmed = q.trim()
  if (trimmed) next.q = trimmed

  if (filters.type !== 'all') next.type = filters.type
  if (filters.from) next.from = filters.from
  if (filters.to) next.to = filters.to
  if (filters.sort !== 'relevance') next.sort = filters.sort
  if (filters.order !== 'desc') next.order = filters.order
  if (filters.folderId) next.sfolderId = filters.folderId

  return next
}

export function routeQueriesEqual(a: RouteQueryRecord, b: RouteQueryRecord): boolean {
  const keys = new Set([...Object.keys(a), ...Object.keys(b)])
  for (const key of keys) {
    if (a[key] !== b[key]) return false
  }
  return true
}

export function getActiveFilterChipKeys(filters: SearchFilters): FilesSearchChipKey[] {
  const chips: FilesSearchChipKey[] = []
  if (filters.type !== 'all') chips.push('type')
  if (filters.folderId) chips.push('folderId')
  if (filters.from) chips.push('from')
  if (filters.to) chips.push('to')
  if (filters.sort !== 'relevance') chips.push('sort')
  if (filters.sort !== 'relevance' && filters.order !== 'desc') chips.push('order')
  return chips
}

export function removeSearchFilter(filters: SearchFilters, key: FilesSearchChipKey): SearchFilters {
  const next: SearchFilters = { ...filters }
  switch (key) {
    case 'type':
      next.type = 'all'
      break
    case 'folderId':
      delete next.folderId
      break
    case 'from':
      delete next.from
      break
    case 'to':
      delete next.to
      break
    case 'sort':
      next.sort = 'relevance'
      next.order = 'desc'
      break
    case 'order':
      next.order = 'desc'
      break
  }
  return next
}

export function resetSearchFilters(): SearchFilters {
  return { ...DEFAULT_SEARCH_FILTERS }
}

export type EmptySearchCause = 'query' | 'filters' | 'both'

/** Classify why a search returned no rows. */
export function getEmptySearchCause(q: string, filters: SearchFilters): EmptySearchCause {
  const hasQuery = q.trim().length > 0
  const hasFilters = hasActiveSearchFilters(filters)
  if (hasQuery && hasFilters) return 'both'
  if (hasFilters) return 'filters'
  return 'query'
}
