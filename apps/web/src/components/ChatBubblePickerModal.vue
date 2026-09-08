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
                >
                  <img v-if="dec.imgUrl" :src="dec.imgUrl" class="bubble-decoration-img" alt="" />
                  <div v-else-if="dec.svg" v-html="dec.svg" />
                </div>
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
                <div class="mini-bubble-wrap">
                  <img :src="item.thumbUrl" :alt="item.name" class="mini-bubble-img" />
                </div>
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
  background: #000000;
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
  background: #000000;
  flex-shrink: 0;
}

.modal-title {
  font-size: 1.1rem;
  font-weight: 700;
  color: #ffffff;
  margin: 0;
  text-align: center;
}

.action-btn {
  background: transparent;
  border: none;
  font-size: 0.95rem;
  font-weight: 500;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  transition: opacity 0.15s ease, color 0.15s ease;
}

.cancel-btn {
  color: #ffffff;
}

.cancel-btn:hover {
  opacity: 0.8;
}

.save-btn {
  color: #7f1d1d;
  font-weight: 600;
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
  padding: 24px 16px 28px;
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
  line-height: 1.45;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.25);
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

.bubble-decoration-img {
  display: block;
  max-height: 40px;
  width: auto;
  object-fit: contain;
  pointer-events: none;
  user-select: none;
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.2));
}

.dec-top-left {
  top: -20px;
  left: 8px;
}

.dec-top-right {
  top: -20px;
  right: 8px;
}

.dec-bottom-left {
  bottom: -14px;
  left: 8px;
}

.dec-bottom-right {
  bottom: -12px;
  right: 8px;
}

.dec-left {
  left: -18px;
  top: 50%;
  transform: translateY(-50%);
}

.dec-right {
  right: -18px;
  top: 50%;
  transform: translateY(-50%);
}

.dec-top {
  top: -18px;
  left: 50%;
  transform: translateX(-50%);
}

.dec-bottom {
  bottom: -10px;
  left: 50%;
  transform: translateX(-50%);
}

/* Bottom Selection Sheet */
.bubble-sheet-panel {
  flex: 1;
  background: #1e1e1e;
  border-top-left-radius: 22px;
  border-top-right-radius: 22px;
  overflow-y: auto;
  padding: 0 16px 36px;
  overscroll-behavior: contain;
  -webkit-overflow-scrolling: touch;
}

.sheet-drag-pill {
  width: 36px;
  height: 4px;
  border-radius: 2px;
  background: #3e3e40;
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
  color: #8e8e93;
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
  background: #2c2c2e;
  border-color: rgba(255, 255, 255, 0.08);
}

.mini-bubble-wrap {
  width: 100%;
  height: 46px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 6px;
}

.mini-bubble-img {
  max-height: 42px;
  max-width: 95%;
  object-fit: contain;
  filter: drop-shadow(0 2px 4px rgba(0, 0, 0, 0.25));
  user-select: none;
  pointer-events: none;
}

.style-name {
  font-size: 0.8rem;
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
  font-weight: 600;
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
