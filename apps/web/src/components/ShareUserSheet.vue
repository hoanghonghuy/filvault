<script setup lang="ts">
import { ref, watch } from 'vue'
import BottomSheet from '@/components/BottomSheet.vue'

const props = defineProps<{
  open: boolean
  name: string
}>()

const emit = defineEmits<{
  share: [email: string]
  close: []
}>()

const email = ref('')
const submitting = ref(false)

watch(
  () => props.open,
  (open) => {
    if (open) {
      email.value = ''
      submitting.value = false
    }
  },
)

function onSubmit() {
  const value = email.value.trim()
  if (!value || submitting.value) return
  submitting.value = true
  emit('share', value)
}
</script>

<template>
  <BottomSheet :open="open" :title="name" @close="emit('close')">
    <p class="hint">They'll get read access to this item. Unknown emails receive an invite to sign up.</p>
    <label class="field">
      <span class="field-label">Email</span>
      <input
        v-model="email"
        type="email"
        autocomplete="email"
        placeholder="name@example.com"
        autofocus
        @keyup.enter="onSubmit"
      />
    </label>
    <div class="actions">
      <button
        type="button"
        class="btn block ink"
        :disabled="submitting || !email.trim()"
        @click="onSubmit"
      >
        {{ submitting ? 'Sharing…' : 'Share' }}
      </button>
      <button type="button" class="btn block ghost" @click="emit('close')">Close</button>
    </div>
  </BottomSheet>
</template>

<style scoped>
.hint {
  margin: 0 0 var(--space-md);
  font-size: 0.8125rem;
  line-height: 1.45;
  color: var(--muted);
}

.field {
  display: flex;
  flex-direction: column;
  gap: var(--space-xxs);
  margin-bottom: var(--space-md);
}

.field-label {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--muted);
}

.actions {
  display: flex;
  flex-direction: column;
}
</style>
