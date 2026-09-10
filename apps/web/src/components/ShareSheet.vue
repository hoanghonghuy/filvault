<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import BottomSheet from '@/components/BottomSheet.vue'
import Icon from '@/components/AppIcon.vue'
import type { ShareLinkTTL } from '@/api/types'

const props = defineProps<{
  open: boolean
  name: string
  existing: {
    url: string
    expiresAt: string | null
    createdAt: string
  } | null
}>()

const emit = defineEmits<{
  create: [ttl: ShareLinkTTL | null]
  copy: [url: string]
  revoke: []
  close: []
}>()

const TTL_OPTIONS: Array<{ value: ShareLinkTTL | null; label: string }> = [
  { value: null, label: 'Forever' },
  { value: '1h', label: '1 hour' },
  { value: '24h', label: '24 hours' },
  { value: '7d', label: '7 days' },
]

const selectedTTL = ref<ShareLinkTTL | null>(null)
const creating = ref(false)

const fullUrl = computed(() => (props.existing ? window.location.origin + props.existing.url : ''))

watch(
  () => props.open,
  (open) => {
    if (open) {
      selectedTTL.value = null
      creating.value = false
    }
  },
)

function selectTtlByKeyboard(event: KeyboardEvent, index: number) {
  let nextIndex: number | null = null

  if (event.key === 'ArrowRight' || event.key === 'ArrowDown') {
    nextIndex = (index + 1) % TTL_OPTIONS.length
  } else if (event.key === 'ArrowLeft' || event.key === 'ArrowUp') {
    nextIndex = (index - 1 + TTL_OPTIONS.length) % TTL_OPTIONS.length
  } else if (event.key === 'Home') {
    nextIndex = 0
  } else if (event.key === 'End') {
    nextIndex = TTL_OPTIONS.length - 1
  }

  if (nextIndex === null) return

  event.preventDefault()
  selectedTTL.value = TTL_OPTIONS[nextIndex]?.value ?? null

  const group = (event.currentTarget as HTMLElement | null)?.closest('.ttl-options')
  group?.querySelectorAll<HTMLButtonElement>('[role="radio"]').item(nextIndex).focus()
}

async function onCreate() {
  creating.value = true
  emit('create', selectedTTL.value)
}
</script>

<template>
  <BottomSheet :open="open" :title="name" @close="emit('close')">
    <div v-if="existing" class="existing">
      <p class="field-label">Anyone with this link can view and download</p>
      <div class="link-box">
        <span class="link-url">{{ fullUrl }}</span>
        <button type="button" class="btn icon-only" aria-label="Copy link" @click="emit('copy', existing.url)">
          <Icon name="copy" :size="18" />
        </button>
      </div>
      <p class="expiry muted">
        {{ existing.expiresAt ? `Expires ${new Date(existing.expiresAt).toLocaleString()}` : 'Never expires' }}
      </p>
      <button type="button" class="btn block danger" @click="emit('revoke')">Revoke link</button>
      <button type="button" class="btn block ghost" @click="emit('close')">Close</button>
    </div>

    <div v-else class="create">
      <p class="field-label">Link expires after</p>
      <div class="ttl-options" role="radiogroup" aria-label="Link expiry">
        <button
          v-for="(option, index) in TTL_OPTIONS"
          :key="option.label"
          type="button"
          role="radio"
          class="ttl-option"
          :class="{ active: selectedTTL === option.value }"
          :aria-checked="selectedTTL === option.value"
          :tabindex="selectedTTL === option.value ? 0 : -1"
          @click="selectedTTL = option.value"
          @keydown="selectTtlByKeyboard($event, index)"
        >
          {{ option.label }}
        </button>
      </div>
      <button type="button" class="btn block ink" :disabled="creating" @click="onCreate">
        {{ creating ? 'Creating…' : 'Create link' }}
      </button>
      <button type="button" class="btn block ghost" @click="emit('close')">Close</button>
    </div>
  </BottomSheet>
</template>

<style scoped>
.field-label {
  margin: 0 0 var(--space-xs);
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--muted);
}

.ttl-options {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-xxs);
  margin-bottom: var(--space-md);
}

.ttl-option {
  min-height: var(--touch-min);
  border: 1px solid var(--hairline);
  border-radius: var(--radius-md);
  background: var(--surface);
  color: var(--ink);
  font-weight: 500;
  cursor: pointer;
  transition:
    background var(--duration-short) var(--ease-standard),
    border-color var(--duration-short) var(--ease-standard);
}

.ttl-option.active {
  border-color: var(--accent);
  background: color-mix(in srgb, var(--accent) 10%, transparent);
  color: var(--accent);
}

.existing {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

.create {
  display: flex;
  flex-direction: column;
}

.link-box {
  display: flex;
  align-items: center;
  gap: var(--space-xxs);
  padding: var(--space-xs) var(--space-sm);
  border: 1px solid var(--hairline);
  border-radius: var(--radius-md);
  background: var(--surface-soft);
}

.link-url {
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-family: monospace;
  font-size: 0.8125rem;
}

.expiry {
  margin: 0;
  font-size: 0.8125rem;
}

.muted {
  color: var(--muted);
}
</style>
