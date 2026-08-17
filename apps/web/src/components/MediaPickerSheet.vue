<script setup lang="ts">
import { ref, watch } from 'vue'
import { api } from '@/api/client'
import { formatApiError } from '@/api/errors'
import BottomSheet from '@/components/BottomSheet.vue'
import EmptyState from '@/components/EmptyState.vue'
import PhotoPlaceholder from '@/components/PhotoPlaceholder.vue'
import type { Timeline, TimelineItem } from '@/api/types'

const props = defineProps<{
  open: boolean
  excludeIds?: string[]
}>()

const emit = defineEmits<{
  select: [fileId: string]
  close: []
}>()

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
    error.value = formatApiError(e, 'Failed to load photos')
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
    error.value = formatApiError(e, 'Failed to load more')
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
  <BottomSheet :open="open" title="Add to album" @close="emit('close')">
    <p class="muted hint">Tap a photo or video to add it.</p>
    <p v-if="error" class="error">{{ error }}</p>
    <p v-if="loading" class="muted">Loading…</p>

    <div v-else class="grid photos picker-grid">
      <PhotoPlaceholder
        v-for="item in flattenItems()"
        :key="item.id"
        :mime-type="item.mimeType"
        :name="item.name"
        @click="pick(item)"
      />
    </div>

    <EmptyState
      v-if="!loading && flattenItems().length === 0"
      title="No media available"
      description="Upload photos or videos in My Files first."
    />

    <button
      v-if="nextBefore"
      type="button"
      class="btn block"
      :disabled="loadingMore"
      @click="loadMore"
    >
      {{ loadingMore ? 'Loading…' : 'Load more' }}
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
