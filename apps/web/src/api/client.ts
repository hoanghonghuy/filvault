import type { ApiErrorBody, Session } from './types'

export const API_BASE =
  import.meta.env.VITE_API_BASE_URL ||
  (typeof window !== 'undefined' && window.location
    ? `${window.location.origin}/api/v1`
    : 'http://localhost:8080/api/v1')

const ACCESS_KEY = 'filvault.accessToken'
const REFRESH_KEY = 'filvault.refreshToken'

let accessToken: string | null = typeof localStorage !== 'undefined' ? localStorage.getItem(ACCESS_KEY) : null
let refreshToken: string | null = typeof localStorage !== 'undefined' ? localStorage.getItem(REFRESH_KEY) : null

export class ApiError extends Error {
  code: string
  status: number

  constructor(code: string, message: string, status: number) {
    super(message)
    this.code = code
    this.status = status
  }
}

export function getAccessToken(): string | null {
  return accessToken
}

export function setTokens(access: string, refresh: string): void {
  accessToken = access
  refreshToken = refresh
  localStorage.setItem(ACCESS_KEY, access)
  localStorage.setItem(REFRESH_KEY, refresh)
}

export function clearTokens(): void {
  accessToken = null
  refreshToken = null
  localStorage.removeItem(ACCESS_KEY)
  localStorage.removeItem(REFRESH_KEY)
}

async function parseError(res: Response): Promise<ApiError> {
  try {
    const body = (await res.json()) as ApiErrorBody
    return new ApiError(body.error.code, body.error.message, res.status)
  } catch {
    return new ApiError('INTERNAL', res.statusText || 'Request failed', res.status)
  }
}

let refreshPromise: Promise<boolean> | null = null

async function refreshAccess(): Promise<boolean> {
  if (!refreshToken) {
    return false
  }
  if (refreshPromise) {
    return refreshPromise
  }
  refreshPromise = (async () => {
    try {
      const res = await fetch(`${API_BASE}/auth/refresh`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refreshToken }),
      })
      if (!res.ok) {
        clearTokens()
        return false
      }
      const session = (await res.json()) as Session
      setTokens(session.accessToken, session.refreshToken)
      return true
    } catch {
      return false
    } finally {
      refreshPromise = null
    }
  })()
  return refreshPromise
}

export async function refreshAccessToken(): Promise<boolean> {
  return refreshAccess()
}

export async function api<T>(
  path: string,
  init: RequestInit = {},
  options: { auth?: boolean; retry?: boolean } = {},
): Promise<T> {
  const auth = options.auth ?? true
  const retry = options.retry ?? true
  const headers = new Headers(init.headers)
  if (init.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }
  if (auth && accessToken) {
    headers.set('Authorization', `Bearer ${accessToken}`)
  }

  const res = await fetch(`${API_BASE}${path}`, { ...init, headers })
  if (res.status === 401 && auth && retry && (await refreshAccess())) {
    return api<T>(path, init, { auth, retry: false })
  }
  if (res.status === 204) {
    return undefined as T
  }
  if (!res.ok) {
    throw await parseError(res)
  }
  return (await res.json()) as T
}

let wakeLockSentinel: { release: () => Promise<void>; released?: boolean } | null = null
let activeUploadCount = 0
let isAcquiringWakeLock = false

async function acquireWakeLock(): Promise<void> {
  activeUploadCount++
  if (activeUploadCount !== 1 || isAcquiringWakeLock || wakeLockSentinel) return
  if (!('wakeLock' in navigator)) return
  isAcquiringWakeLock = true
  try {
    const sentinel = await (navigator as unknown as { wakeLock: { request: (type: string) => Promise<{ release: () => Promise<void>; released?: boolean }> } }).wakeLock.request('screen')
    // All uploads may have finished while we were awaiting
    if ((activeUploadCount as number) === 0) {
      void sentinel.release().catch(() => {})
    } else {
      wakeLockSentinel = sentinel
    }
  } catch {
    // Wake lock not supported or not allowed, fail-open
  } finally {
    isAcquiringWakeLock = false
  }
}

function releaseWakeLock(): void {
  activeUploadCount = Math.max(0, activeUploadCount - 1)
  if (activeUploadCount === 0 && wakeLockSentinel) {
    const sentinel = wakeLockSentinel
    wakeLockSentinel = null
    void sentinel.release().catch(() => {})
  }
}

if (typeof document !== 'undefined') {
  document.addEventListener('visibilitychange', () => {
    if (document.visibilityState === 'visible' && activeUploadCount > 0 && (!wakeLockSentinel || wakeLockSentinel.released)) {
      wakeLockSentinel = null
      void acquireWakeLock()
    }
  })
}

export { normalizePresignedUrl } from '@/lib/presignedUrl'

export function uploadToPresigned(
  url: string,
  file: File,
  contentType: string,
  onProgress?: (ratio: number) => void,
  signal?: AbortSignal,
): Promise<void> {
  // Handle pre-aborted signal immediately
  if (signal?.aborted) {
    return Promise.reject(new ApiError('UPLOAD_CANCELED', 'Upload canceled', 0))
  }

  void acquireWakeLock()

  const targetUrl = normalizePresignedUrl(url)

  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest()
    let settled = false

    const cleanup = () => {
      if (settled) return
      settled = true
      releaseWakeLock()
      signal?.removeEventListener('abort', handleAbort)
    }

    const handleAbort = () => xhr.abort()
    signal?.addEventListener('abort', handleAbort, { once: true })

    xhr.upload.onprogress = (event) => {
      if (event.lengthComputable && onProgress && event.total > 0) {
        onProgress(event.loaded / event.total)
      }
    }
    xhr.onload = () => {
      cleanup()
      if (xhr.status >= 200 && xhr.status < 300) {
        resolve()
        return
      }
      reject(new ApiError('UPLOAD_FAILED', `Upload failed (${xhr.status})`, xhr.status))
    }
    xhr.onerror = () => {
      cleanup()
      reject(new ApiError('UPLOAD_FAILED', 'Upload network error', 0))
    }
    xhr.onabort = () => {
      cleanup()
      reject(new ApiError('UPLOAD_CANCELED', 'Upload canceled', 0))
    }

    try {
      xhr.open('PUT', targetUrl)
      xhr.setRequestHeader('Content-Type', contentType)
      xhr.send(file)
    } catch (err) {
      cleanup()
      reject(err instanceof ApiError ? err : new ApiError('UPLOAD_FAILED', 'Upload initialization failed', 0))
    }
  })
}

export function formatBytes(n: number): string {
  if (n < 1024) return `${n} B`
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KiB`
  if (n < 1024 * 1024 * 1024) return `${(n / (1024 * 1024)).toFixed(1)} MiB`
  return `${(n / (1024 * 1024 * 1024)).toFixed(2)} GiB`
}
