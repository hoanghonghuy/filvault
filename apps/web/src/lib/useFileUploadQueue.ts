import { computed, ref } from 'vue'
import { api, uploadToPresigned } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { resolveContentType } from '@/lib/mimeIcon'
import { useI18n } from '@/lib/i18n'
import type { UploadSession } from '@/api/types'
import {
  FileUploadQueue,
  computeAggregateProgress,
  type UploadQueueItem,
} from '@/lib/uploadQueue'

function formatTemplate(template: string, values: Record<string, string | number>): string {
  return template.replace(/\{(\w+)\}/g, (_match, key: string) => String(values[key] ?? ''))
}

export function useFileUploadQueue(options: {
  getRootFolderId: () => string | null
  onBatchSettled: () => Promise<void>
  showToast: (message: string, tone?: 'success' | 'info') => void
}) {
  const { t } = useI18n()
  const queueRevision = ref(0)
  const showPanel = ref(false)

  const bump = () => {
    queueRevision.value += 1
  }

  const queue = new FileUploadQueue({
    onChange: bump,
    createSession: async (params) =>
      api<UploadSession>('/files/upload-sessions', {
        method: 'POST',
        body: JSON.stringify(params),
      }),
    uploadBytes: (url, file, contentType, onProgress, signal) =>
      uploadToPresigned(url, file, contentType, onProgress, signal),
    completeUpload: async (fileId, attempt) => {
      const item = queue.items.find(
        (entry) => entry.status === 'uploading' && entry.attempt === attempt,
      )
      if (!item) return
      await api(`/files/${fileId}/complete`, { method: 'POST', body: '{}' })
    },
    ensureFolderPath: async (parts, rootParentId, cache) => {
      let parentId = rootParentId
      let currentPath = ''
      for (const part of parts) {
        currentPath = currentPath ? `${currentPath}/${part}` : part
        const cached = cache.get(currentPath)
        if (cached) {
          parentId = cached
          continue
        }
        const created = await api<{ id: string }>('/folders/get-or-create', {
          method: 'POST',
          body: JSON.stringify({ name: part, parentId }),
        })
        parentId = created.id
        cache.set(currentPath, parentId)
      }
      return parentId
    },
    resolveContentType: (file) => resolveContentType(file),
    formatError: (error) => formatApiError(error, 'Upload failed'),
  })

  const items = computed<UploadQueueItem[]>(() => {
    const revision = queueRevision.value
    return revision >= 0 ? [...queue.items] : []
  })

  const aggregateProgress = computed(() => {
    const revision = queueRevision.value
    if (revision < 0 || !showPanel.value) return null
    return computeAggregateProgress(queue.items)
  })

  const progressItems = computed(() =>
    items.value
      .filter((item) => item.status !== 'cancelled')
      .map((item) => ({
        id: item.id,
        name: item.displayName,
        resolvedName: item.resolvedName,
        status: item.status,
        progress: item.progress,
        error: item.error,
      })),
  )

  const isUploading = computed(() => {
    const revision = queueRevision.value
    return revision >= 0 && queue.items.some((item) => item.status === 'queued' || item.status === 'uploading')
  })

  async function runQueue(): Promise<void> {
    await queue.run()
    bump()

    const active = queue.items.filter((item) => item.status !== 'cancelled')
    const allSuccess = active.length > 0 && active.every((item) => item.status === 'completed')
    const hasFailures = active.some((item) => item.status === 'failed')

    if (allSuccess) {
      options.showToast(
        active.length === 1
          ? t.value.uploadCompleteToast
          : formatTemplate(t.value.uploadCompleteManyToast, { count: active.length }),
        'success',
      )
      await options.onBatchSettled()
      queue.dismissSettled()
      if (!queue.hasActiveWork()) {
        showPanel.value = false
      }
      bump()
      return
    }

    if (hasFailures && !queue.items.some((item) => item.status === 'queued' || item.status === 'uploading')) {
      const completed = active.filter((item) => item.status === 'completed').length
      if (completed > 0) {
        await options.onBatchSettled()
      }
    }
  }

  function enqueueFiles(files: File[], isFolderUpload = false) {
    if (files.length === 0) return
    showPanel.value = true
    queue.enqueueFiles(files, options.getRootFolderId(), { isFolderUpload })
    bump()
    void runQueue()
  }

  function retryUpload(id: string) {
    if (!queue.retry(id)) return
    bump()
    void runQueue()
  }

  function cancelUpload(id: string) {
    if (!queue.cancel(id)) return
    bump()
    if (!queue.hasActiveWork()) {
      queue.dismissSettled()
      showPanel.value = queue.items.length > 0
    }
    void runQueue()
  }

  function dismissFailedUpload(id: string) {
    const item = queue.items.find((entry) => entry.id === id)
    if (!queue.dismissFailed(id)) return
    if (item) {
      options.showToast(
        formatTemplate(t.value.uploadDismissedFailedToast, { name: item.displayName }),
        'info',
      )
    }
    bump()
    if (!queue.hasActiveWork()) {
      showPanel.value = queue.items.length > 0
    }
  }

  function dismissUploadPanel() {
    queue.dismissSettled()
    showPanel.value = queue.hasActiveWork()
    bump()
  }

  return {
    aggregateProgress,
    progressItems,
    isUploading,
    enqueueFiles,
    retryUpload,
    cancelUpload,
    dismissFailedUpload,
    dismissUploadPanel,
  }
}
