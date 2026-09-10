<script setup lang="ts">
import { computed, ref } from 'vue'
import Icon from '@/components/AppIcon.vue'

const props = withDefaults(
  defineProps<{
    id: string
    name: string
    autocomplete: string
    required?: boolean
    minlength?: number
    disabled?: boolean
    ariaInvalid?: boolean
    ariaDescribedby?: string
  }>(),
  {
    required: false,
    minlength: undefined,
    disabled: false,
    ariaInvalid: false,
    ariaDescribedby: undefined,
  },
)

const model = defineModel<string>({ required: true })
const revealed = ref(false)
const toggleLabel = computed(() => (revealed.value ? 'Hide password' : 'Show password'))
</script>

<template>
  <div class="password-control">
    <input
      :id="id"
      v-model="model"
      :name="name"
      :type="revealed ? 'text' : 'password'"
      :required="required"
      :minlength="minlength"
      :autocomplete="autocomplete"
      :disabled="disabled"
      :aria-invalid="ariaInvalid ? true : undefined"
      :aria-describedby="ariaDescribedby"
    />
    <button
      type="button"
      class="password-toggle"
      :disabled="disabled"
      :aria-label="toggleLabel"
      :aria-pressed="revealed"
      :title="toggleLabel"
      @click="revealed = !revealed"
    >
      <Icon :name="revealed ? 'eye-off' : 'eye'" :size="18" />
    </button>
  </div>
</template>

<style scoped>
.password-control {
  position: relative;
  width: 100%;
}

.password-control input {
  width: 100%;
  padding-inline-end: 52px;
}

.password-toggle {
  position: absolute;
  inset-inline-end: 2px;
  top: 50%;
  width: var(--touch-min);
  height: var(--touch-min);
  display: inline-grid;
  place-items: center;
  padding: 0;
  border: 0;
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--muted);
  cursor: pointer;
  transform: translateY(-50%);
}

.password-toggle:hover:not(:disabled) {
  background: var(--surface-soft);
  color: var(--ink);
}

.password-toggle:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 1px;
}

.password-toggle:disabled {
  cursor: not-allowed;
  opacity: 0.55;
}
</style>
