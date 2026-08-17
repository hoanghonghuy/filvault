import type { ApiErrorBody, Session } from './types'

const BASE = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080/api/v1'

const ACCESS_KEY = 'filnest.accessToken'
const REFRESH_KEY = 'filnest.refreshToken'

let accessToken: string | null = localStorage.getItem(ACCESS_KEY)
let refreshToken: string | null = localStorage.getItem(REFRESH_KEY)

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

async function refreshAccess(): Promise<boolean> {
  if (!refreshToken) {
    return false
  }
  const res = await fetch(`${BASE}/auth/refresh`, {
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

  const res = await fetch(`${BASE}${path}`, { ...init, headers })
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

export function uploadToPresigned(
  url: string,
  file: File,
  contentType: string,
  onProgress?: (ratio: number) => void,
): Promise<void> {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest()
    xhr.upload.onprogress = (event) => {
      if (event.lengthComputable && onProgress) {
        onProgress(event.loaded / event.total)
      }
    }
    xhr.onload = () => {
      if (xhr.status >= 200 && xhr.status < 300) {
        resolve()
        return
      }
      reject(new ApiError('UPLOAD_FAILED', `Upload failed (${xhr.status})`, xhr.status))
    }
    xhr.onerror = () => reject(new ApiError('UPLOAD_FAILED', 'Upload network error', 0))
    xhr.open('PUT', url)
    xhr.setRequestHeader('Content-Type', contentType)
    xhr.send(file)
  })
}

export function formatBytes(n: number): string {
  if (n < 1024) return `${n} B`
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KiB`
  if (n < 1024 * 1024 * 1024) return `${(n / (1024 * 1024)).toFixed(1)} MiB`
  return `${(n / (1024 * 1024 * 1024)).toFixed(2)} GiB`
}
