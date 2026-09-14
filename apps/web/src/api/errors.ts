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

export function formatAuthError(e: unknown, fallback: string): string {
  if (!(e instanceof ApiError)) {
    return fallback
  }
  return authFriendly[e.code] ?? fallback
}
