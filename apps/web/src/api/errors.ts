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

export function formatApiError(e: unknown, fallback: string): string {
  if (!(e instanceof ApiError)) {
    return fallback
  }
  if (e.code === 'CONFLICT' && e.message) {
    return e.message
  }
  return friendly[e.code] ?? e.message ?? fallback
}
