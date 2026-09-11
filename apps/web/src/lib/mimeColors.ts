export const MIME_CATEGORY_COLORS = {
  image: '#8b5cf6',
  video: '#ec4899',
  audio: '#06b6d4',
  pdf: '#ef4444',
  archive: '#f97316',
  spreadsheet: '#10b981',
  document: '#0084ff',
  folder: '#f59e0b',
  file: '#3b82f6',
  unknown: '#64748b',
} as const

export type MimeColorCategory = keyof typeof MIME_CATEGORY_COLORS

const ARCHIVE_EXTENSIONS = new Set(['zip', 'tar', 'gz', 'rar', '7z'])
const SPREADSHEET_EXTENSIONS = new Set(['xls', 'xlsx', 'csv', 'ods'])
const DOCUMENT_EXTENSIONS = new Set(['doc', 'docx', 'odt', 'rtf', 'txt', 'md'])

function extensionFromName(name?: string): string {
  if (!name) return ''
  const index = name.lastIndexOf('.')
  if (index < 0 || index === name.length - 1) return ''
  return name.slice(index + 1).toLowerCase()
}

export function mimeColorCategory(mimeType?: string, name?: string): MimeColorCategory {
  const mime = mimeType?.toLowerCase() ?? ''
  const ext = extensionFromName(name)

  if (mime.startsWith('image/')) return 'image'
  if (mime.startsWith('video/')) return 'video'
  if (mime.startsWith('audio/')) return 'audio'
  if (mime.includes('pdf') || ext === 'pdf') return 'pdf'
  if (
    mime.includes('zip') ||
    mime.includes('tar') ||
    mime.includes('rar') ||
    mime.includes('7z') ||
    mime.includes('compressed') ||
    ARCHIVE_EXTENSIONS.has(ext)
  ) return 'archive'
  if (
    mime.includes('sheet') ||
    mime.includes('excel') ||
    mime.includes('csv') ||
    SPREADSHEET_EXTENSIONS.has(ext)
  ) return 'spreadsheet'
  if (
    mime.includes('word') ||
    mime.includes('document') ||
    mime.startsWith('text/') ||
    DOCUMENT_EXTENSIONS.has(ext)
  ) return 'document'
  if (mime) return 'file'
  if (name) return 'file'
  return 'unknown'
}

export function mimeCategoryColor(mimeType?: string, name?: string): string {
  return MIME_CATEGORY_COLORS[mimeColorCategory(mimeType, name)]
}
