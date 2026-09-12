import { computed, ref } from 'vue'
import { api } from '@/api/client'
import type { StorageUsage } from '@/api/types'
import { useI18n } from '@/lib/i18n'
import {
  storageFillRatio,
  storageProgressAria,
  storageToneClass,
  storageUsagePercent,
  storageUsageState,
  type StorageUsageState,
} from '@/lib/storageUsage'

export function useStorageUsage(formatBytes: (bytes: number) => string) {
  const { locale } = useI18n()
  const usage = ref<StorageUsage | null>(null)
  const loading = ref(true)
  const unavailable = ref(false)

  const usedBytes = computed(() => usage.value?.usedBytes ?? 0)
  const quotaBytes = computed(() => usage.value?.quotaBytes ?? 0)

  const state = computed<StorageUsageState>(() =>
    storageUsageState(usedBytes.value, quotaBytes.value, {
      loading: loading.value,
      unavailable: unavailable.value,
    }),
  )

  const percent = computed(() => storageUsagePercent(usedBytes.value, quotaBytes.value))
  const fillRatio = computed(() => storageFillRatio(usedBytes.value, quotaBytes.value))
  const toneClass = computed(() => storageToneClass(state.value))
  const progressAria = computed(() =>
    storageProgressAria(usedBytes.value, quotaBytes.value, formatBytes, locale.value),
  )

  async function reload() {
    loading.value = true
    unavailable.value = false
    try {
      usage.value = await api<StorageUsage>('/storage')
    } catch {
      usage.value = null
      unavailable.value = true
    } finally {
      loading.value = false
    }
  }

  return {
    usage,
    loading,
    unavailable,
    usedBytes,
    quotaBytes,
    state,
    percent,
    fillRatio,
    toneClass,
    progressAria,
    reload,
  }
}
