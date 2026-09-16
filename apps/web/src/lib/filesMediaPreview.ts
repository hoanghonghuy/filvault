import type { DownloadURL } from '@/api/types'

export type FilesMediaPreviewPayload = {
  id: string
  name: string
  mimeType: string
  url: string
}

export type FilesMediaPreviewDownloadResult =
  | { status: 'success'; preview: FilesMediaPreviewPayload }
  | { status: 'error'; error: unknown }
  | { status: 'stale' }

export function createFilesMediaPreviewSession() {
  let previewSequence = 0
  return {
    beginPreview(): number {
      previewSequence += 1
      return previewSequence
    },
    isCurrentPreview(requestSequence: number): boolean {
      return requestSequence === previewSequence
    },
    invalidatePreview(): void {
      previewSequence += 1
    },
  }
}

export async function runFilesMediaPreviewDownload(options: {
  file: { id: string; name: string; mimeType: string }
  requestSequence: number
  isCurrentPreview: (requestSequence: number) => boolean
  fetchDownload: (fileId: string) => Promise<DownloadURL>
}): Promise<FilesMediaPreviewDownloadResult> {
  const { file, requestSequence, isCurrentPreview, fetchDownload } = options
  try {
    const out = await fetchDownload(file.id)
    if (!isCurrentPreview(requestSequence)) return { status: 'stale' }
    return {
      status: 'success',
      preview: {
        id: file.id,
        name: file.name,
        mimeType: file.mimeType,
        url: out.downloadUrl,
      },
    }
  } catch (error) {
    if (!isCurrentPreview(requestSequence)) return { status: 'stale' }
    return { status: 'error', error }
  }
}
