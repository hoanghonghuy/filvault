<script setup lang="ts">
import BottomSheet from '@/components/BottomSheet.vue'
import Icon from '@/components/AppIcon.vue'
import { useI18n } from '@/lib/i18n'

defineProps<{
  open: boolean
  name: string
  favorited?: boolean
}>()

const emit = defineEmits<{
  view: []
  download: []
  favorite: []
  close: []
  'after-leave': []
}>()

const { t } = useI18n()
</script>

<template>
  <BottomSheet :open="open" :title="name" @close="emit('close')" @after-leave="emit('after-leave')">
    <div class="actions">
      <button type="button" class="btn block ink" @click="emit('view')">{{ t.preview }}</button>
      <button type="button" class="btn block" @click="emit('download')">{{ t.download }}</button>
      <button type="button" class="btn block favorite-action" @click="emit('favorite')">
        <Icon :name="favorited ? 'star-filled' : 'star'" :size="18" class="star-icon" />
        {{ favorited ? t.removeFromFavorites : t.addToFavorites }}
      </button>
      <button type="button" class="btn block ghost" @click="emit('close')">{{ t.close }}</button>
    </div>
  </BottomSheet>
</template>

<style scoped>
.favorite-action {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: var(--space-xs);
}

.star-icon {
  color: var(--warning);
}
</style>
