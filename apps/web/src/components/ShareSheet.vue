<script setup lang="ts">
import { computed, ref, useAttrs, watch } from 'vue'
import BottomSheet from '@/components/BottomSheet.vue'
import Icon from '@/components/AppIcon.vue'
import type { ShareLinkTTL } from '@/api/types'
import { useI18n } from '@/lib/i18n'

defineOptions({ inheritAttrs: false })

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
  close: []
}>()

const attrs = useAttrs()
const { locale } = useI18n()

const copy = computed(() =>
  locale.value === 'en'
    ? {
        anyone: 'Anyone with this link can view and download',
        copying: 'Copying link…',
        copyLink: 'Copy link',
        expires: 'Expires',
        neverExpires: 'Never expires',
        revoking: 'Revoking…',
        revoke: 'Revoke link',
        close: 'Close',
        expiresAfter: 'Link expires after',
        expiryGroup: 'Link expiry',
        forever: 'Forever',
        oneHour: '1 hour',
        oneDay: '24 hours',
        sevenDays: '7 days',
        creating: 'Creating…',
        create: 'Create link',
      }
    : {
        anyone: 'Bất kỳ ai có liên kết này đều có thể xem và tải xuống',
        copying: 'Đang sao chép liên kết…',
        copyLink: 'Sao chép liên kết',
        expires: 'Hết hạn',
        neverExpires: 'Không hết hạn',
        revoking: 'Đang thu hồi…',
        revoke: 'Thu hồi liên kết',
        close: 'Đóng',
        expiresAfter: 'Liên kết hết hạn sau',
        expiryGroup: 'Thời hạn liên kết',
        forever: 'Vĩnh viễn',
        oneHour: '1 giờ',
        oneDay: '24 giờ',
        sevenDays: '7 ngày',
        creating: 'Đang tạo…',
        create: 'Tạo liên kết',
      },
)

type Listener = (...args: unknown[]) => unknown

function getListener(name: 'onCreate' | 'onCopy' | 'onRevoke'): Listener | null {
  const listener = attrs[name]
  return typeof listener === 'function' ? (listener as Listener) : null
}

async function invokeListener(name: 'onCreate' | 'onCopy' | 'onRevoke', ...args: unknown[]) {
  const listener = getListener(name)
  if (!listener) return
  try {
    await listener(...args)
  } catch {
    // FilesView owns user-facing operation errors. The sheet still guarantees
    // that its pending state settles even if a future listener throws.
  }
}

const ttlOptions = computed<Array<{ value: ShareLinkTTL | null; label: string }>>(() => [
  { value: null, label: copy.value.forever },
  { value: '1h', label: copy.value.oneHour },
  { value: '24h', label: copy.value.oneDay },
  { value: '7d', label: copy.value.sevenDays },
])

const selectedTTL = ref<ShareLinkTTL | null>(null)
const creating = ref(false)
const copying = ref(false)
const revoking = ref(false)
const busy = computed(() => creating.value || copying.value || revoking.value)

const fullUrl = computed(() => (props.existing ? window.location.origin + props.existing.url : ''))
const expiryLabel = computed(() => {
  if (!props.existing?.expiresAt) return copy.value.neverExpires
  const formatted = new Date(props.existing.expiresAt).toLocaleString(locale.value === 'vi' ? 'vi-VN' : 'en-US')
  return `${copy.value.expires} ${formatted}`
})

watch(
  () => props.open,
  (open) => {
    if (open) selectedTTL.value = null
  },
)

function selectTtlByKeyboard(event: KeyboardEvent, index: number) {
  let nextIndex: number | null = null

  if (event.key === 'ArrowRight' || event.key === 'ArrowDown') {
    nextIndex = (index + 1) % ttlOptions.value.length
  } else if (event.key === 'ArrowLeft' || event.key === 'ArrowUp') {
    nextIndex = (index - 1 + ttlOptions.value.length) % ttlOptions.value.length
  } else if (event.key === 'Home') {
    nextIndex = 0
  } else if (event.key === 'End') {
    nextIndex = ttlOptions.value.length - 1
  }

  if (nextIndex === null) return

  event.preventDefault()
  selectedTTL.value = ttlOptions.value[nextIndex]?.value ?? null

  const group = (event.currentTarget as HTMLElement | null)?.closest('.ttl-options')
  group?.querySelectorAll<HTMLButtonElement>('[role="radio"]').item(nextIndex).focus()
}

async function onCreate() {
  if (creating.value || !getListener('onCreate')) return
  creating.value = true
  try {
    await invokeListener('onCreate', selectedTTL.value)
  } finally {
    creating.value = false
  }
}

async function onCopy(url: string) {
  if (copying.value || !getListener('onCopy')) return
  copying.value = true
  try {
    await invokeListener('onCopy', url)
  } finally {
    copying.value = false
  }
}

async function onRevoke() {
  if (revoking.value || !getListener('onRevoke')) return
  revoking.value = true
  try {
    await invokeListener('onRevoke')
  } finally {
    revoking.value = false
  }
}
</script>

<template>
  <BottomSheet :open="open" :title="name" @close="emit('close')">
    <div v-if="existing" class="existing">
      <p class="field-label">{{ copy.anyone }}</p>
      <div class="link-box">
        <span class="link-url">{{ fullUrl }}</span>
        <button
          type="button"
          class="btn icon-only"
          :aria-label="copying ? copy.copying : copy.copyLink"
          :aria-busy="copying ? 'true' : undefined"
          :disabled="busy"
          @click="onCopy(existing.url)"
        >
          <Icon name="copy" :size="18" />
        </button>
      </div>
      <p class="expiry muted">{{ expiryLabel }}</p>
      <button
        type="button"
        class="btn block danger"
        :disabled="busy"
        :aria-busy="revoking ? 'true' : undefined"
        @click="onRevoke"
      >
        {{ revoking ? copy.revoking : copy.revoke }}
      </button>
      <button type="button" class="btn block ghost" @click="emit('close')">{{ copy.close }}</button>
    </div>

    <div v-else class="create">
      <p class="field-label">{{ copy.expiresAfter }}</p>
      <div class="ttl-options" role="radiogroup" :aria-label="copy.expiryGroup">
        <button
          v-for="(option, index) in ttlOptions"
          :key="option.value ?? 'forever'"
          type="button"
          role="radio"
          class="ttl-option"
          :class="{ active: selectedTTL === option.value }"
          :aria-checked="selectedTTL === option.value"
          :tabindex="selectedTTL === option.value ? 0 : -1"
          :disabled="creating"
          @click="selectedTTL = option.value"
          @keydown="selectTtlByKeyboard($event, index)"
        >
          {{ option.label }}
        </button>
      </div>
      <button
        type="button"
        class="btn block ink"
        :disabled="creating"
        :aria-busy="creating ? 'true' : undefined"
        @click="onCreate"
      >
        {{ creating ? copy.creating : copy.create }}
      </button>
      <button type="button" class="btn block ghost" @click="emit('close')">{{ copy.close }}</button>
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
