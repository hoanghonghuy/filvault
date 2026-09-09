/** Initials for avatar mark — Drive-like account affordance without uploading a photo. */
export function userInitials(displayName: string, email = ''): string {
  const parts = displayName.trim().split(/\s+/).filter(Boolean)
  if (parts.length >= 2) {
    return `${parts[0]![0]!}${parts[1]![0]!}`.toUpperCase()
  }
  if (parts.length === 1 && parts[0]!.length > 0) {
    return parts[0]![0]!.toUpperCase()
  }
  const local = email.trim().split('@')[0] ?? ''
  if (local.length > 0) {
    return local[0]!.toUpperCase()
  }
  return 'U'
}
