<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import type { ChatAttachment } from '@/api/types'
import { isHeic, getHeicDisplayUrl } from '@/lib/heic'
import { normalizePresignedUrl } from '@/api/client'
import { chatAttachmentCopy } from '@/lib/chatAttachmentCopy'
import { useI18n } from '@/lib/i18n'

const props = defineProps<{
  attachment: ChatAttachment
}>()

const emit = defineEmits<{
  click: [attachment: ChatAttachment]
}>()

const { locale } = useI18n()
const copy = computed(() => chatAttachmentCopy(locale.value))
const isHeicImage = computed(() => isHeic(props.attachment.name, props.attachment.mimeType))
const heicUrl = ref<string>('')
const loading = ref(false)
const failed = ref(false)

const displaySrc = computed(() => {
  if (isHeicImage.value && heicUrl.value) {
    return heicUrl.value
  }
  return normalizePresignedUrl(props.attachment.thumbnailUrl || '')
})

async function resolveImage() {
  if (!isHeicImage.value || !props.attachment.thumbnailUrl) return
  loading.value = true
  failed.value = false
  try {
    const sourceUrl = normalizePresignedUrl(props.attachment.thumbnailUrl)
    const url = await getHeicDisplayUrl(sourceUrl, 0.7)
    heicUrl.value = url
  } catch {
    failed.value = true
  } finally {
    loading.value = false
  }
}

watch(() => props.attachment.thumbnailUrl, resolveImage)
onMounted(resolveImage)
</script>

<template>
  <button
    type="button"
    class="inline-image"
    :class="{ 'is-heic': isHeicImage }"
    :aria-label="copy.viewImageAria(attachment.name)"
    @click="emit('click', attachment)"
  >
    <div v-if="loading" class="heic-loading-box" aria-live="polite">
      <div class="heic-spinner" />
      <span>HEIC</span>
    </div>
    <img
      v-else
      :src="displaySrc"
      :alt="attachment.name"
      loading="lazy"
      decoding="async"
      @error="failed = true"
    />
    <span v-if="isHeicImage && !loading" class="heic-badge">HEIC</span>
  </button>
</template>

<style scoped>
.inline-image {
  display: block;
  width: min(260px, 100%);
  margin-top: var(--space-xs);
  padding: 0;
  border: none;
  border-radius: var(--radius-lg);
  overflow: hidden;
  cursor: zoom-in;
  background: transparent;
  position: relative;
}

.inline-image img {
  display: block;
  width: 100%;
  max-height: 320px;
  object-fit: cover;
  border-radius: var(--radius-lg);
}

.inline-image:focus-visible {
  outline: 2px solid var(--on-ink, #ffffff);
  outline-offset: 2px;
}

.heic-loading-box {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--space-xs);
  min-height: 140px;
  width: 100%;
  background: rgba(0, 0, 0, 0.08);
  color: var(--muted);
  font-size: 0.75rem;
  font-weight: 600;
  border-radius: var(--radius-lg);
}

.heic-spinner {
  width: 22px;
  height: 22px;
  border: 2.5px solid var(--line);
  border-top-color: var(--accent);
  border-radius: 50%;
  animation: heic-spin 0.8s linear infinite;
}

@keyframes heic-spin {
  to {
    transform: rotate(360deg);
  }
}

.heic-badge {
  position: absolute;
  top: 6px;
  right: 6px;
  padding: 2px 6px;
  border-radius: var(--radius-sm);
  background: rgba(0, 0, 0, 0.65);
  color: #ffffff;
  font-size: 0.65rem;
  font-weight: 700;
  letter-spacing: 0.05em;
  backdrop-filter: blur(4px);
  pointer-events: none;
}
</style>
