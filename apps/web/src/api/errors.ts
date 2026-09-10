import { ApiError } from '@/api/client'

const friendly: Record<string, string> = {
  CONFLICT: 'Name already exists in this location.',
  QUOTA_EXCEEDED: 'Storage quota exceeded. Free up space and try again.',
  FILE_TOO_LARGE: 'File exceeds the 100 MiB limit.',
  VALIDATION_ERROR: 'Invalid input. Check the file type and try again.',
  INVALID_STATE: 'This action is not allowed right now.',
  NOT_FOUND: 'Item not found.',
  UPLOAD_EXPIRED: 'Upload session expired. Try uploading again.',
}

const authFriendly: Record<string, string> = {
  CONFLICT: 'An account with this email already exists.',
  FORBIDDEN: 'Invalid invite code.',
  REGISTER_DISABLED: 'Registration is currently disabled.',
  UNAUTHORIZED: 'Incorrect email or password.',
}

const NETWORK_UNAVAILABLE = "Can't reach the server. Try again in a moment."

function isNetworkError(e: unknown): boolean {
  return e instanceof TypeError
}

export function formatApiError(e: unknown, fallback: string): string {
  if (isNetworkError(e)) {
    return NETWORK_UNAVAILABLE
  }
  if (!(e instanceof ApiError)) {
    return fallback
  }
  if (e.code === 'CONFLICT') {
    if (e.message && e.message.toLowerCase() !== 'conflict') {
      return e.message
    }
    return friendly.CONFLICT ?? 'Name already exists in this location.'
  }
  return friendly[e.code] ?? e.message ?? fallback
}

const shareUserFriendly: Record<string, string> = {
  CONFLICT: 'This item is already shared with that user.',
  NOT_FOUND: 'Item not found.',
  VALIDATION_ERROR: 'Enter a valid email address.',
}

export interface ShareUserErrorCopy {
  conflict?: string
  notFound?: string
  validation?: string
  network?: string
}

export function formatShareUserError(
  e: unknown,
  fallback: string,
  copy: ShareUserErrorCopy = {},
): string {
  if (isNetworkError(e)) {
    return copy.network ?? NETWORK_UNAVAILABLE
  }
  if (!(e instanceof ApiError)) {
    return fallback
  }
  if (e.code === 'CONFLICT') {
    return copy.conflict ?? shareUserFriendly.CONFLICT ?? fallback
  }
  if (e.code === 'NOT_FOUND') {
    return copy.notFound ?? shareUserFriendly.NOT_FOUND ?? fallback
  }
  if (e.code === 'VALIDATION_ERROR') {
    return copy.validation ?? shareUserFriendly.VALIDATION_ERROR ?? fallback
  }
  return fallback
}

export function formatAuthError(e: unknown, fallback: string): string {
  if (!(e instanceof ApiError)) {
    return fallback
  }
  return authFriendly[e.code] ?? fallback
}
