<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { api } from '@/api/client'
import { formatApiError } from '@/api/errors'
import BottomSheet from '@/components/BottomSheet.vue'
import EmptyState from '@/components/EmptyState.vue'
import PhotoThumb from '@/components/PhotoThumb.vue'
import type { Timeline, TimelineItem } from '@/api/types'
import { useI18n } from '@/lib/i18n'
import { mediaPickerCopy } from '@/lib/mediaPickerCopy'

const props = defineProps<{
  open: boolean
  excludeIds?: string[]
}>()

const emit = defineEmits<{
  select: [fileId: string]
  close: []
}>()

const { t, locale } = useI18n()
const copy = computed(() => mediaPickerCopy(locale.value))
const groups = ref<Timeline['groups']>([])
const nextBefore = ref<string | undefined>()
const loading = ref(false)
const loadingMore = ref(false)
const error = ref('')

function flattenItems(): TimelineItem[] {
  const ids = new Set(props.excludeIds ?? [])
  const items: TimelineItem[] = []
  for (const group of groups.value) {
    for (const item of group.items) {
      if (!ids.has(item.id)) {
        items.push(item)
      }
    }
  }
  return items
}

async function fetchTimeline(before?: string) {
  const params = new URLSearchParams()
  if (before) params.set('before', before)
  return api<Timeline>(`/photos/timeline${params.toString() ? `?${params}` : ''}`)
}

async function loadInitial() {
  loading.value = true
  error.value = ''
  groups.value = []
  nextBefore.value = undefined
  try {
    const data = await fetchTimeline()
    groups.value = data.groups
    nextBefore.value = data.nextBefore
  } catch (e) {
    error.value = formatApiError(e, copy.value.loadFailed)
  } finally {
    loading.value = false
  }
}

async function loadMore() {
  if (!nextBefore.value || loadingMore.value) return
  loadingMore.value = true
  error.value = ''
  try {
    const data = await fetchTimeline(nextBefore.value)
    groups.value = [...groups.value, ...data.groups]
    nextBefore.value = data.nextBefore
  } catch (e) {
    error.value = formatApiError(e, copy.value.loadMoreFailed)
  } finally {
    loadingMore.value = false
  }
}

function pick(item: TimelineItem) {
  emit('select', item.id)
}

watch(
  () => props.open,
  (open) => {
    if (open) {
      void loadInitial()
    }
  },
)
</script>

<template>
  <BottomSheet :open="open" :title="copy.title" @close="emit('close')">
    <p class="muted hint">{{ copy.hint }}</p>
    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <p v-if="loading" class="muted">{{ t.loading }}</p>

    <div v-else class="grid photos picker-grid">
      <PhotoThumb
        v-for="item in flattenItems()"
        :key="item.id"
        :mime-type="item.mimeType"
        :name="item.name"
        :thumbnail-url="item.thumbnailUrl"
        @click="pick(item)"
      />
    </div>

    <EmptyState
      v-if="!loading && flattenItems().length === 0"
      :title="copy.emptyTitle"
      :description="copy.emptyDescription"
    />

    <button
      v-if="nextBefore"
      type="button"
      class="btn block"
      :disabled="loadingMore"
      @click="loadMore"
    >
      {{ loadingMore ? t.loading : t.loadMore }}
    </button>
  </BottomSheet>
</template>

<style scoped>
.hint {
  margin: 0 0 var(--space-sm);
}

.picker-grid {
  margin-bottom: var(--space-md);
  max-height: 50vh;
  overflow-y: auto;
}
</style>
