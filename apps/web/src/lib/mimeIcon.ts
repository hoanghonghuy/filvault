export function mimeIcon(mimeType: string): string {
  if (mimeType.startsWith('image/')) return 'image'
  if (mimeType.startsWith('video/')) return 'video'
  if (mimeType === 'application/zip') return 'archive'
  if (mimeType.startsWith('application/') || mimeType.startsWith('text/')) return 'doc'
  return 'file'
}

export function mimeLabel(mimeType: string): string {
  if (mimeType.startsWith('image/')) return 'Image'
  if (mimeType.startsWith('video/')) return 'Video'
  if (mimeType === 'application/zip') return 'Archive'
  if (mimeType === 'application/pdf') return 'PDF'
  if (mimeType.startsWith('text/')) return 'Text'
  if (mimeType.startsWith('application/')) return 'Document'
  return 'File'
}
