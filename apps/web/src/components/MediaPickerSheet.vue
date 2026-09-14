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
let loadGeneration = 0

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
  const generation = ++loadGeneration
  loading.value = true
  loadingMore.value = false
  error.value = ''
  groups.value = []
  nextBefore.value = undefined
  try {
    const data = await fetchTimeline()
    if (generation !== loadGeneration || !props.open) return
    groups.value = data.groups
    nextBefore.value = data.nextBefore
  } catch (e) {
    if (generation !== loadGeneration || !props.open) return
    error.value = formatApiError(e, copy.value.loadFailed)
  } finally {
    if (generation === loadGeneration && props.open) {
      loading.value = false
    }
  }
}

async function loadMore() {
  if (!nextBefore.value || loadingMore.value) return
  const generation = ++loadGeneration
  const before = nextBefore.value
  loadingMore.value = true
  error.value = ''
  try {
    const data = await fetchTimeline(before)
    if (generation !== loadGeneration || !props.open) return
    groups.value = [...groups.value, ...data.groups]
    nextBefore.value = data.nextBefore
  } catch (e) {
    if (generation !== loadGeneration || !props.open) return
    error.value = formatApiError(e, copy.value.loadMoreFailed)
  } finally {
    if (generation === loadGeneration && props.open) {
      loadingMore.value = false
    }
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
      return
    }

    loadGeneration += 1
    loading.value = false
    loadingMore.value = false
  },
)
</script>

<template>
  <BottomSheet :open="open" :title="copy.title" @close="emit('close')">
    <p class="muted hint">{{ copy.hint }}</p>
    <p v-if="error && groups.length > 0" class="error" role="alert">{{ error }}</p>
    <p v-if="loading" class="muted">{{ t.loading }}</p>

    <div v-else-if="error && groups.length === 0" class="picker-error" role="alert">
      <p class="error">{{ error }}</p>
      <button type="button" class="btn block" @click="loadInitial">
        {{ t.retry }}
      </button>
    </div>

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
      v-if="!loading && !error && flattenItems().length === 0"
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

.picker-error {
  display: grid;
  gap: var(--space-sm);
}

.picker-error .error {
  margin: 0;
}

.picker-grid {
  margin-bottom: var(--space-md);
  max-height: 50vh;
  overflow-y: auto;
}
</style>
