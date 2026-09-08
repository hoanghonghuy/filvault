<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import {
  CHAT_BUBBLE_STYLES,
  activeBubbleStyleId,
  getBubbleStyle,
  setBubbleStyle,
  type ChatBubbleStyle,
} from '@/lib/chatBubbles'
import { useUiStore } from '@/stores/ui'

const props = defineProps<{
  open: boolean
}>()

const emit = defineEmits<{
  close: []
  saved: [id: string]
}>()

const ui = useUiStore()

const previewId = ref<string>(activeBubbleStyleId.value)
const isChanged = computed(() => previewId.value !== activeBubbleStyleId.value)

watch(
  () => props.open,
  (val) => {
    if (val) {
      previewId.value = activeBubbleStyleId.value
    }
  },
)

const activeStyle = computed<ChatBubbleStyle>(() => getBubbleStyle(previewId.value))

function selectStyle(id: string) {
  previewId.value = id
}

function handleSave() {
  setBubbleStyle(previewId.value)
  ui.showToast(`Đã đổi kiểu bong bóng sang "${activeStyle.value.name}"`, 'success')
  emit('saved', previewId.value)
  emit('close')
}

function handleCancel() {
  previewId.value = activeBubbleStyleId.value
  emit('close')
}
</script>

<template>
  <Teleport to="body">
    <Transition name="bubble-modal">
      <div v-if="open" class="bubble-picker-backdrop" @click.self="handleCancel">
        <div class="bubble-picker-window" role="dialog" aria-modal="true" aria-labelledby="bubble-modal-title">
          <!-- Top Bar -->
          <header class="bubble-modal-header">
            <button type="button" class="action-btn cancel-btn" @click="handleCancel">
              Hủy
            </button>
            <h2 id="bubble-modal-title" class="modal-title">Chọn kiểu bong bóng</h2>
            <button
              type="button"
              class="action-btn save-btn"
              :class="{ changed: isChanged }"
              @click="handleSave"
            >
              Lưu
            </button>
          </header>

          <!-- Upper Live Preview Area -->
          <div class="live-preview-section">
            <div class="preview-stage">
              <div
                class="preview-bubble"
                :style="{
                  background: activeStyle.bg,
                  color: activeStyle.color,
                  border: activeStyle.border || 'none',
                }"
              >
                <!-- Decorations peeking on corners/borders -->
                <div
                  v-for="(dec, idx) in activeStyle.decorations"
                  :key="idx"
                  class="bubble-decoration"
                  :class="`dec-${dec.position}`"
                  v-html="dec.svg"
                />
                <p class="preview-text">
                  Nay bạn có thể thay đổi kiểu bong bóng và cuộc trò chuyện sẽ có giao diện mới. Quá ngầu!
                </p>
              </div>
            </div>
          </div>

          <!-- Bottom Selection Sheet -->
          <div class="bubble-sheet-panel">
            <div class="sheet-drag-pill" aria-hidden="true" />
            <div class="sheet-header">
              <h3 class="sheet-title">Gợi ý</h3>
              <p class="sheet-subtext">
                Bong bóng này áp dụng cho tất cả cuộc trò chuyện. Bong bóng này chỉ ảnh hưởng đến các tin nhắn bạn gửi sau khi lưu.
              </p>
            </div>

            <!-- 3-Column Grid of 21 Styles -->
            <div class="styles-grid" role="radiogroup" aria-label="Kiểu bong bóng">
              <button
                v-for="item in CHAT_BUBBLE_STYLES"
                :key="item.id"
                type="button"
                class="style-card"
                :class="{ active: previewId === item.id }"
                role="radio"
                :aria-checked="previewId === item.id"
                :aria-label="item.name"
                @click="selectStyle(item.id)"
              >
                <div class="mini-bubble-wrap" v-html="item.previewSvg" />
                <span class="style-name">{{ item.name }}</span>
              </button>
            </div>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<style scoped>
.bubble-picker-backdrop {
  position: fixed;
  inset: 0;
  z-index: 9999;
  background: rgba(0, 0, 0, 0.85);
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  display: flex;
  align-items: flex-end;
  justify-content: center;
}

@media (min-width: 768px) {
  .bubble-picker-backdrop {
    align-items: center;
    padding: var(--space-md);
  }
}

.bubble-picker-window {
  display: flex;
  flex-direction: column;
  width: 100%;
  max-width: 480px;
  height: 94vh;
  height: 94dvh;
  background: #0d0d0f;
  border-top-left-radius: 24px;
  border-top-right-radius: 24px;
  overflow: hidden;
  box-shadow: 0 -4px 32px rgba(0, 0, 0, 0.6);
}

@media (min-width: 768px) {
  .bubble-picker-window {
    height: 88vh;
    max-height: 840px;
    border-radius: 24px;
  }
}

/* Header */
.bubble-modal-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 18px;
  background: #0d0d0f;
  flex-shrink: 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.modal-title {
  font-size: 1.05rem;
  font-weight: 700;
  color: #ffffff;
  margin: 0;
  text-align: center;
}

.action-btn {
  background: transparent;
  border: none;
  font-size: 0.95rem;
  font-weight: 600;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  transition: opacity 0.15s ease, color 0.15s ease;
}

.cancel-btn {
  color: #e4e4e7;
}

.cancel-btn:hover {
  color: #ffffff;
}

.save-btn {
  color: #71717a;
}

.save-btn.changed {
  color: #ef4444;
  font-weight: 700;
}

.save-btn.changed:hover {
  color: #f87171;
}

/* Upper Live Preview Area */
.live-preview-section {
  padding: 24px 16px 20px;
  background: #000000;
  display: flex;
  justify-content: flex-end;
  flex-shrink: 0;
}

.preview-stage {
  width: 100%;
  display: flex;
  justify-content: flex-end;
  padding-right: 6px;
}

.preview-bubble {
  position: relative;
  max-width: 82%;
  padding: 12px 18px;
  border-radius: 18px;
  font-size: 15px;
  line-height: 1.4;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.2);
  transition: background 0.2s ease, color 0.2s ease;
  word-break: break-word;
}

.preview-text {
  margin: 0;
  position: relative;
  z-index: 2;
}

/* Decorations positioned on bubble */
.bubble-decoration {
  position: absolute;
  pointer-events: none;
  z-index: 3;
  display: flex;
  align-items: center;
  justify-content: center;
}

.dec-top-left {
  top: -14px;
  left: 6px;
}

.dec-top-right {
  top: -14px;
  right: 6px;
}

.dec-bottom-left {
  bottom: -10px;
  left: 6px;
}

.dec-bottom-right {
  bottom: -10px;
  right: 6px;
}

.dec-left {
  left: -14px;
  top: 50%;
  transform: translateY(-50%);
}

.dec-right {
  right: -14px;
  top: 50%;
  transform: translateY(-50%);
}

.dec-top {
  top: -12px;
  left: 50%;
  transform: translateX(-50%);
}

.dec-bottom {
  bottom: -6px;
  left: 50%;
  transform: translateX(-50%);
}

/* Bottom Selection Sheet */
.bubble-sheet-panel {
  flex: 1;
  background: #18181b;
  border-top-left-radius: 24px;
  border-top-right-radius: 24px;
  overflow-y: auto;
  padding: 0 16px 32px;
  overscroll-behavior: contain;
  -webkit-overflow-scrolling: touch;
}

.sheet-drag-pill {
  width: 36px;
  height: 4px;
  border-radius: 2px;
  background: #3f3f46;
  margin: 10px auto 14px;
}

.sheet-header {
  margin-bottom: 18px;
}

.sheet-title {
  font-size: 1.15rem;
  font-weight: 700;
  color: #ffffff;
  margin: 0 0 6px;
}

.sheet-subtext {
  font-size: 0.8rem;
  line-height: 1.45;
  color: #a1a1aa;
  margin: 0;
}

/* 3-Column Grid */
.styles-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 14px 10px;
}

.style-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 10px 4px 8px;
  background: transparent;
  border: 1.5px solid transparent;
  border-radius: 16px;
  cursor: pointer;
  transition: all 0.15s ease;
  user-select: none;
  -webkit-tap-highlight-color: transparent;
}

.style-card:hover {
  background: rgba(255, 255, 255, 0.04);
}

.style-card.active {
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.22);
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.3);
}

.mini-bubble-wrap {
  width: 100%;
  height: 44px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 8px;
}

.style-name {
  font-size: 0.78rem;
  font-weight: 500;
  color: #d4d4d8;
  text-align: center;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 100%;
}

.style-card.active .style-name {
  color: #ffffff;
  font-weight: 700;
}

/* Transition */
.bubble-modal-enter-active,
.bubble-modal-leave-active {
  transition: opacity 0.25s ease;
}

.bubble-modal-enter-from,
.bubble-modal-leave-to {
  opacity: 0;
}

.bubble-modal-enter-active .bubble-picker-window {
  transition: transform 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}

.bubble-modal-leave-active .bubble-picker-window {
  transition: transform 0.2s ease-in;
}

.bubble-modal-enter-from .bubble-picker-window,
.bubble-modal-leave-to .bubble-picker-window {
  transform: translateY(100%);
}

@media (min-width: 768px) {
  .bubble-modal-enter-from .bubble-picker-window,
  .bubble-modal-leave-to .bubble-picker-window {
    transform: scale(0.95) translateY(12px);
  }
}
</style>
