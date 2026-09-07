<script setup lang="ts">
import { computed, ref } from 'vue'
import AppIcon from '@/components/AppIcon.vue'
import {
  EMOJI_CATEGORIES,
  EMOJIS,
  type EmojiCategory,
} from '@/lib/emojis'
import {
  STICKERS,
  STICKER_CATEGORIES,
  type Sticker,
  type StickerCategory,
} from '@/lib/stickers'

const props = withDefaults(
  defineProps<{
    showStickers?: boolean
    activeReactions?: string[]
    showClose?: boolean
  }>(),
  {
    showStickers: false,
    activeReactions: () => [],
    showClose: false,
  }
)

const emit = defineEmits<{
  (e: 'select-emoji', emoji: string): void
  (e: 'select-sticker', sticker: Sticker): void
  (e: 'close'): void
}>()

const activeType = ref<'emoji' | 'sticker'>('emoji')
const activeEmojiCategory = ref<EmojiCategory['id']>('popular')
const activeStickerCategory = ref<StickerCategory['id']>('expressions')

const currentEmojis = computed(() => EMOJIS[activeEmojiCategory.value] ?? EMOJIS.popular)
const currentStickers = computed(() => STICKERS.filter((s) => s.category === activeStickerCategory.value))

function isReacted(emoji: string): boolean {
  return props.activeReactions.includes(emoji)
}

function handleEmojiClick(emoji: string) {
  emit('select-emoji', emoji)
}

function handleStickerClick(sticker: Sticker) {
  emit('select-sticker', sticker)
}
</script>

<template>
  <div class="emoji-picker" role="region" aria-label="Bộ chọn biểu tượng cảm xúc và nhãn dán">
    <!-- Header with Type Switcher (if stickers enabled) and Close button -->
    <div v-if="showStickers || showClose" class="picker-top-bar">
      <div v-if="showStickers" class="picker-type-switch" role="tablist" aria-label="Loại biểu tượng">
        <button
          type="button"
          class="type-switch-btn"
          :class="{ active: activeType === 'emoji' }"
          role="tab"
          :aria-selected="activeType === 'emoji'"
          @click="activeType = 'emoji'"
        >
          <span>Biểu tượng</span>
        </button>
        <button
          type="button"
          class="type-switch-btn"
          :class="{ active: activeType === 'sticker' }"
          role="tab"
          :aria-selected="activeType === 'sticker'"
          @click="activeType = 'sticker'"
        >
          <span>Nhãn dán</span>
        </button>
      </div>

      <button
        v-if="showClose"
        type="button"
        class="picker-close-btn"
        aria-label="Đóng"
        @click="emit('close')"
      >
        <AppIcon name="close" :size="16" />
      </button>
    </div>

    <!-- Category Bar for Emojis -->
    <div v-if="activeType === 'emoji'" class="picker-categories" role="tablist" aria-label="Danh mục biểu tượng">
      <button
        v-for="cat in EMOJI_CATEGORIES"
        :key="cat.id"
        type="button"
        class="picker-cat-btn"
        :class="{ active: activeEmojiCategory === cat.id }"
        role="tab"
        :aria-selected="activeEmojiCategory === cat.id"
        @click="activeEmojiCategory = cat.id"
      >
        <span class="cat-icon">{{ cat.icon }}</span>
        <span class="cat-label">{{ cat.label }}</span>
      </button>
    </div>

    <!-- Category Bar for Stickers -->
    <div v-else class="picker-categories" role="tablist" aria-label="Danh mục nhãn dán">
      <button
        v-for="cat in STICKER_CATEGORIES"
        :key="cat.id"
        type="button"
        class="picker-cat-btn"
        :class="{ active: activeStickerCategory === cat.id }"
        role="tab"
        :aria-selected="activeStickerCategory === cat.id"
        @click="activeStickerCategory = cat.id"
      >
        <span class="cat-icon">{{ cat.icon }}</span>
        <span class="cat-label">{{ cat.label }}</span>
      </button>
    </div>

    <!-- Emoji Grid -->
    <div v-if="activeType === 'emoji'" class="picker-grid emoji-grid" role="list">
      <button
        v-for="emoji in currentEmojis"
        :key="emoji"
        type="button"
        class="picker-item-btn emoji-item"
        :class="{ 'item-reacted': isReacted(emoji) }"
        :aria-label="emoji"
        @click="handleEmojiClick(emoji)"
      >
        <span class="item-symbol">{{ emoji }}</span>
      </button>
    </div>

    <!-- Sticker Grid -->
    <div v-else class="picker-grid sticker-grid" role="list">
      <button
        v-for="stk in currentStickers"
        :key="stk.id"
        type="button"
        class="picker-item-btn sticker-item"
        :title="stk.name"
        :aria-label="stk.name"
        @click="handleStickerClick(stk)"
      >
        <span class="item-symbol">{{ stk.symbol }}</span>
      </button>
    </div>
  </div>
</template>

<style scoped>
.emoji-picker {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 4px 4px 12px;
  width: 100%;
  box-sizing: border-box;
}

.picker-top-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  padding: 2px 4px;
}

.picker-type-switch {
  display: inline-flex;
  background: var(--surface-soft, rgba(255, 255, 255, 0.08));
  border-radius: 999px;
  padding: 3px;
  gap: 2px;
}

.type-switch-btn {
  border: none;
  background: transparent;
  padding: 5px 14px;
  border-radius: 999px;
  font-size: 13px;
  font-weight: 500;
  color: var(--muted, #8a8d91);
  cursor: pointer;
  transition: all 0.15s ease;
  user-select: none;
}

.type-switch-btn:hover {
  color: var(--ink, #ffffff);
}

.type-switch-btn.active {
  background: var(--surface-card, #ffffff);
  color: var(--ink, #050505);
  box-shadow: 0 2px 6px rgba(0, 0, 0, 0.12);
}

[data-theme='dark'] .type-switch-btn.active,
:root:not([data-theme='light']) .type-switch-btn.active {
  background: #3a3b3c;
  color: #ffffff;
}

.picker-close-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 50%;
  background: transparent;
  color: var(--muted, #8a8d91);
  cursor: pointer;
  transition: background 0.15s ease;
  margin-left: auto;
}

.picker-close-btn:hover {
  background: var(--surface-soft, rgba(255, 255, 255, 0.1));
  color: var(--ink, #ffffff);
}

.picker-categories {
  display: flex;
  align-items: center;
  gap: 6px;
  overflow-x: auto;
  scrollbar-width: none;
  -webkit-overflow-scrolling: touch;
  padding: 2px 2px 4px;
}

.picker-categories::-webkit-scrollbar {
  display: none;
}

.picker-cat-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 12px;
  border-radius: 999px;
  border: 1px solid var(--hairline, rgba(255, 255, 255, 0.1));
  background: var(--surface-card, rgba(255, 255, 255, 0.05));
  color: var(--muted, #8a8d91);
  font-size: 13px;
  font-weight: 500;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.15s ease;
  user-select: none;
  touch-action: manipulation;
}

.picker-cat-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: var(--ink, #ffffff);
}

[data-theme='light'] .picker-cat-btn:hover,
:root:not([data-theme='dark']) .picker-cat-btn:hover {
  background: rgba(0, 0, 0, 0.05);
  color: #050505;
}

.picker-cat-btn.active {
  background: var(--brand, #0084ff);
  border-color: var(--brand, #0084ff);
  color: #ffffff;
  box-shadow: 0 2px 8px rgba(0, 132, 255, 0.35);
}

.cat-icon {
  font-size: 14px;
  line-height: 1;
}

.cat-label {
  line-height: 1;
}

.picker-grid {
  display: grid;
  gap: 6px;
  max-height: 260px;
  overflow-y: auto;
  overscroll-behavior: contain;
  padding: 4px;
  scrollbar-width: thin;
}

.emoji-grid {
  grid-template-columns: repeat(auto-fill, minmax(46px, 1fr));
}

.sticker-grid {
  grid-template-columns: repeat(auto-fill, minmax(56px, 1fr));
}

.picker-item-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  aspect-ratio: 1;
  min-height: 44px;
  border: none;
  border-radius: 12px;
  background: transparent;
  cursor: pointer;
  padding: 0;
  transition: transform 0.15s cubic-bezier(0.34, 1.56, 0.64, 1), background 0.15s ease;
  user-select: none;
  touch-action: manipulation;
}

.emoji-item .item-symbol {
  font-size: 28px;
  line-height: 1;
}

.sticker-item .item-symbol {
  font-size: 34px;
  line-height: 1;
}

.picker-item-btn:hover,
.picker-item-btn:active {
  transform: scale(1.22);
  background: rgba(255, 255, 255, 0.1);
}

[data-theme='light'] .picker-item-btn:hover,
:root:not([data-theme='dark']) .picker-item-btn:hover,
[data-theme='light'] .picker-item-btn:active,
:root:not([data-theme='dark']) .picker-item-btn:active {
  background: rgba(0, 0, 0, 0.06);
}

.picker-item-btn.item-reacted {
  background: rgba(0, 132, 255, 0.2);
  box-shadow: inset 0 0 0 1.5px var(--brand, #0084ff);
}
</style>
