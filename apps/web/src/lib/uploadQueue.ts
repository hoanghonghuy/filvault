import { ApiError } from '@/api/client'
import { generateUUID } from '@/lib/uuid'

export type UploadItemStatus = 'queued' | 'uploading' | 'completed' | 'failed' | 'cancelled'

export interface UploadQueueItem {
  id: string
  file: File
  displayName: string
  folderParts: string[]
  rootFolderId: string | null
  status: UploadItemStatus
  progress: number
  error?: string
  controller: AbortController
  attempt: number
  resolvedFolderId?: string | null
}

export interface UploadSessionResult {
  fileId: string
  uploadUrl: string
}

export interface UploadQueueDeps {
  onChange?: () => void
  createSession: (params: {
    name: string
    size: number
    contentType: string
    folderId: string | null
  }) => Promise<UploadSessionResult>
  uploadBytes: (
    url: string,
    file: File,
    contentType: string,
    onProgress?: (ratio: number) => void,
    signal?: AbortSignal,
  ) => Promise<void>
  completeUpload: (fileId: string, attempt: number) => Promise<void>
  ensureFolderPath: (
    parts: string[],
    rootParentId: string | null,
    cache: Map<string, string>,
  ) => Promise<string | null>
  resolveContentType: (file: File) => string | null
  formatError: (error: unknown) => string
}

const MAX_CONFLICT_RENAMES = 10

function fileSize(item: UploadQueueItem): number {
  return Math.max(0, item.file.size)
}

export function computeAggregateProgress(items: UploadQueueItem[]): number | null {
  const active = items.filter((item) => item.status !== 'cancelled')
  if (active.length === 0) return null

  const totalBytes = active.reduce((sum, item) => sum + fileSize(item), 0)
  if (totalBytes === 0) {
    const unresolved = active.some(
      (item) => item.status === 'queued' || item.status === 'uploading' || item.status === 'failed',
    )
    return unresolved ? 0 : 1
  }

  const uploadedBytes = active.reduce((sum, item) => {
    if (item.status === 'completed') return sum + fileSize(item)
    if (item.status === 'uploading') return sum + fileSize(item) * item.progress
    return sum
  }, 0)

  const ratio = uploadedBytes / totalBytes
  const unresolved = active.some(
    (item) => item.status === 'queued' || item.status === 'uploading' || item.status === 'failed',
  )
  if (unresolved && ratio >= 1) return 0.99
  return Math.min(1, ratio)
}

function getNextNumberedName(originalName: string, counter: number): string {
  const lastDot = originalName.lastIndexOf('.')
  if (lastDot > 0) {
    const base = originalName.substring(0, lastDot)
    const ext = originalName.substring(lastDot)
    return `${base} (${counter})${ext}`
  }
  return `${originalName} (${counter})`
}

function folderPartsFromRelativePath(relativePath: string): string[] {
  const parts = relativePath.split('/')
  parts.pop()
  return parts
}

export class FileUploadQueue {
  readonly items: UploadQueueItem[] = []
  private readonly deps: UploadQueueDeps
  private readonly folderCache = new Map<string, string>()
  private running = false
  private runRequested = false

  constructor(deps: UploadQueueDeps) {
    this.deps = deps
  }

  private notifyChange(): void {
    this.deps.onChange?.()
  }

  private shouldStop(item: UploadQueueItem, attempt: number): boolean {
    return item.status === 'cancelled' || item.attempt !== attempt
  }

  enqueueFiles(
    files: File[],
    rootFolderId: string | null,
    options: { isFolderUpload?: boolean } = {},
  ): void {
    for (const file of files) {
      const relativePath =
        options.isFolderUpload
          ? (file as File & { webkitRelativePath?: string }).webkitRelativePath ?? file.name
          : file.name
      const displayName = relativePath
      const folderParts = options.isFolderUpload ? folderPartsFromRelativePath(relativePath) : []

      this.items.push({
        id: generateUUID(),
        file,
        displayName,
        folderParts,
        rootFolderId,
        status: 'queued',
        progress: 0,
        controller: new AbortController(),
        attempt: 1,
      })
    }
  }

  retry(id: string): boolean {
    const item = this.items.find((entry) => entry.id === id)
    if (!item || item.status !== 'failed') return false
    item.status = 'queued'
    item.error = undefined
    item.progress = 0
    item.attempt += 1
    item.controller = new AbortController()
    return true
  }

  cancel(id: string): boolean {
    const item = this.items.find((entry) => entry.id === id)
    if (!item || item.status === 'completed' || item.status === 'cancelled') return false
    if (item.status === 'uploading') {
      item.controller.abort()
    }
    item.status = 'cancelled'
    item.progress = 0
    item.error = undefined
    this.notifyChange()
    return true
  }

  dismissCompleted(): void {
    for (let index = this.items.length - 1; index >= 0; index -= 1) {
      const item = this.items[index]
      if (item && (item.status === 'completed' || item.status === 'cancelled')) {
        this.items.splice(index, 1)
      }
    }
  }

  hasActiveWork(): boolean {
    return this.items.some(
      (item) => item.status === 'queued' || item.status === 'uploading' || item.status === 'failed',
    )
  }

  scheduleRun(): void {
    this.runRequested = true
    void this.run()
  }

  async run(): Promise<void> {
    this.runRequested = true
    if (this.running) return
    this.running = true

    try {
      while (this.runRequested) {
        this.runRequested = false
        const next = this.items.find((item) => item.status === 'queued')
        if (!next) break
        await this.processItem(next)
        if (this.items.some((item) => item.status === 'queued')) {
          this.runRequested = true
        }
      }
    } finally {
      this.running = false
      if (this.items.some((item) => item.status === 'queued')) {
        this.scheduleRun()
      }
    }
  }

  private async processItem(item: UploadQueueItem): Promise<void> {
    if (item.status !== 'queued') return

    const attempt = item.attempt
    item.status = 'uploading'
    item.progress = 0
    item.error = undefined
    this.notifyChange()

    try {
      const contentType = this.deps.resolveContentType(item.file)
      if (!contentType) {
        item.status = 'failed'
        item.error = `Unsupported file type for "${item.displayName}".`
        this.notifyChange()
        return
      }

      let targetFolderId = item.rootFolderId
      if (item.folderParts.length > 0) {
        targetFolderId = await this.deps.ensureFolderPath(
          item.folderParts,
          item.rootFolderId,
          this.folderCache,
        )
        item.resolvedFolderId = targetFolderId
      }

      if (this.shouldStop(item, attempt)) return

      const session = await this.createSessionWithConflictRetry(item, contentType, targetFolderId, attempt)
      if (!session || this.shouldStop(item, attempt)) return

      await this.deps.uploadBytes(
        session.uploadUrl,
        item.file,
        contentType,
        (ratio) => {
          if (item.attempt === attempt && item.status === 'uploading') {
            item.progress = ratio
            this.notifyChange()
          }
        },
        item.controller.signal,
      )

      if (this.shouldStop(item, attempt)) return

      await this.deps.completeUpload(session.fileId, attempt)
      if (this.shouldStop(item, attempt)) return

      item.status = 'completed'
      item.progress = 1
      item.error = undefined
      this.notifyChange()
    } catch (error) {
      if (item.attempt !== attempt) return
      if (error instanceof ApiError && error.code === 'UPLOAD_CANCELED') {
        item.status = 'cancelled'
        item.progress = 0
        item.error = undefined
        this.notifyChange()
        return
      }
      item.status = 'failed'
      item.progress = 0
      item.error = this.deps.formatError(error)
      this.notifyChange()
    }
  }

  private async createSessionWithConflictRetry(
    item: UploadQueueItem,
    contentType: string,
    folderId: string | null,
    attempt: number,
  ): Promise<UploadSessionResult | null> {
    let uploadName = item.file.name
    let renameAttempt = 0

    while (renameAttempt < MAX_CONFLICT_RENAMES) {
      if (this.shouldStop(item, attempt)) return null
      try {
        return await this.deps.createSession({
          name: uploadName,
          size: item.file.size,
          contentType,
          folderId,
        })
      } catch (error) {
        if (
          error instanceof ApiError &&
          error.code === 'CONFLICT' &&
          renameAttempt < MAX_CONFLICT_RENAMES - 1
        ) {
          renameAttempt += 1
          uploadName = getNextNumberedName(item.file.name, renameAttempt)
          continue
        }
        throw error
      }
    }

    return null
  }
}
