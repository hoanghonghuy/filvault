export function mimeIcon(mimeType: string): string {
  if (mimeType.startsWith('image/')) return 'image'
  if (mimeType.startsWith('video/')) return 'video'
  if (mimeType === 'application/zip') return 'archive'
  if (mimeType.startsWith('application/') || mimeType.startsWith('text/')) return 'doc'
  return 'file'
}
