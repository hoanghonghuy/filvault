<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { api } from '@/api/client'
import { formatApiError } from '@/api/errors'
import BottomSheet from '@/components/BottomSheet.vue'
import type { Browser } from '@/api/types'
import { useI18n } from '@/lib/i18n'
import { folderPickerCopy } from '@/lib/folderPickerCopy'

const props = defineProps<{
  open: boolean
  title: string
  excludeFolderId?: string | null
  confirmLabel?: string
}>()

const emit = defineEmits<{
  select: [folderId: string | null]
  close: []
}>()

const { t, locale } = useI18n()
const copy = computed(() => folderPickerCopy(locale.value))
const browseFolderId = ref<string | null>(null)
const browser = ref<Browser | null>(null)
const loading = ref(false)
const error = ref('')

async function loadBrowser() {
  loading.value = true
  error.value = ''
  try {
    const q = browseFolderId.value ? `?folderId=${browseFolderId.value}` : ''
    browser.value = await api<Browser>(`/browser${q}`)
  } catch (e) {
    error.value = formatApiError(e, copy.value.loadFailed)
  } finally {
    loading.value = false
  }
}

function resetBrowse() {
  browseFolderId.value = null
}

function enterFolder(id: string) {
  browseFolderId.value = id
}

function goUp() {
  if (!browser.value?.folder) {
    return
  }
  browseFolderId.value = browser.value.folder.parentId
}

function confirmSelection() {
  emit('select', browseFolderId.value)
}

function onClose() {
  emit('close')
}

watch(
  () => props.open,
  (open) => {
    if (open) {
      resetBrowse()
      void loadBrowser()
    }
  },
)

watch(browseFolderId, () => {
  if (props.open) {
    void loadBrowser()
  }
})
</script>

<template>
  <BottomSheet :open="open" :title="title" @close="onClose">
    <nav v-if="browser" class="picker-breadcrumb" :aria-label="copy.browseFolders">
      <button type="button" class="btn ghost crumb" @click="resetBrowse">{{ copy.root }}</button>
      <template v-for="item in browser.breadcrumb" :key="item.id">
        <span aria-hidden="true">/</span>
        <button type="button" class="btn ghost crumb" @click="enterFolder(item.id)">
          {{ item.name }}
        </button>
      </template>
      <template v-if="browser.folder">
        <span aria-hidden="true">/</span>
        <span class="current">{{ browser.folder.name }}</span>
      </template>
    </nav>

    <p v-if="error" class="error" role="alert">{{ error }}</p>
    <p v-if="loading" class="muted">{{ t.loading }}</p>

    <div v-else class="picker-list">
      <button
        v-if="browseFolderId"
        type="button"
        class="row tappable picker-row"
        @click="goUp"
      >
        <span class="name">{{ copy.parentFolder }}</span>
      </button>
      <button
        v-for="folder in browser?.folders ?? []"
        :key="folder.id"
        type="button"
        class="row tappable picker-row"
        :disabled="folder.id === excludeFolderId"
        @click="enterFolder(folder.id)"
      >
        <span class="name"><span class="icon-folder" />{{ folder.name }}</span>
      </button>
      <p v-if="(browser?.folders.length ?? 0) === 0 && !loading" class="muted picker-empty">
        {{ copy.empty }}
      </p>
    </div>

    <button
      type="button"
      class="btn block ink picker-confirm"
      :disabled="browseFolderId === excludeFolderId"
      @click="confirmSelection"
    >
      {{ confirmLabel ?? t.moveHere }}
    </button>
  </BottomSheet>
</template>

<style scoped>
.picker-breadcrumb {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 0.25rem;
  margin-bottom: var(--space-sm);
  font-size: 0.875rem;
}

.crumb {
  min-height: auto;
  padding: 0.25rem 0.5rem;
  font-size: 0.875rem;
}

.current {
  color: var(--ink);
  font-weight: 600;
}

.picker-list {
  display: flex;
  flex-direction: column;
  gap: var(--space-xs);
  max-height: 40vh;
  overflow-y: auto;
  margin-bottom: var(--space-md);
}

.picker-row {
  width: 100%;
  text-align: left;
  border: 1px solid var(--hairline);
}

.picker-row:disabled {
  opacity: 0.45;
  cursor: not-allowed;
}

.picker-empty {
  margin: var(--space-sm) 0;
  text-align: center;
}

.picker-confirm {
  margin-top: var(--space-xs);
}
</style>
