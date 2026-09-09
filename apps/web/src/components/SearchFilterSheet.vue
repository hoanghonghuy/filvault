<script setup lang="ts">
import { ref, watch } from 'vue'
import BottomSheet from '@/components/BottomSheet.vue'
import type { SearchFilters } from '@/api/types'

const props = defineProps<{
  open: boolean
  filters: SearchFilters
}>()

const emit = defineEmits<{
  close: []
  apply: [filters: SearchFilters]
}>()

const TYPE_OPTIONS: Array<{ value: SearchFilters['type']; label: string }> = [
  { value: 'all', label: 'All' },
  { value: 'image', label: 'Images' },
  { value: 'video', label: 'Videos' },
  { value: 'document', label: 'Documents' },
  { value: 'archive', label: 'Archives' },
  { value: 'folder', label: 'Folders' },
]

const SORT_OPTIONS: Array<{ value: SearchFilters['sort']; label: string }> = [
  { value: 'relevance', label: 'Best match' },
  { value: 'name', label: 'Name' },
  { value: 'date', label: 'Date' },
  { value: 'size', label: 'Size' },
]

type OrderValue = Exclude<SearchFilters['order'], undefined>

const ORDER_OPTIONS: Array<{ value: OrderValue; label: string }> = [
  { value: 'desc', label: 'Descending' },
  { value: 'asc', label: 'Ascending' },
]

const type = ref<SearchFilters['type']>('all')
const sort = ref<SearchFilters['sort']>('relevance')
const order = ref<OrderValue>('desc')
const fromDate = ref('')
const toDate = ref('')

watch(
  () => props.open,
  (open) => {
    if (!open) return
    const f = props.filters
    type.value = f.type ?? 'all'
    sort.value = f.sort ?? 'relevance'
    order.value = f.order ?? 'desc'
    fromDate.value = f.from ?? ''
    toDate.value = f.to ?? ''
  },
)

function onApply() {
  emit('apply', {
    type: type.value,
    sort: sort.value,
    order: order.value,
    ...(fromDate.value ? { from: fromDate.value } : {}),
    ...(toDate.value ? { to: toDate.value } : {}),
  })
}
</script>

<template>
  <BottomSheet :open="open" title="Search filters" @close="emit('close')">
    <form class="filter-form" @submit.prevent="onApply">
      <fieldset class="field-group">
        <legend>Type</legend>
        <div class="option-grid">
          <label v-for="opt in TYPE_OPTIONS" :key="opt.value" class="option-pill">
            <input v-model="type" type="radio" name="filter-type" :value="opt.value" />
            <span>{{ opt.label }}</span>
          </label>
        </div>
      </fieldset>

      <fieldset class="field-group">
        <legend>Sort by</legend>
        <div class="option-grid">
          <label v-for="opt in SORT_OPTIONS" :key="opt.value" class="option-pill">
            <input v-model="sort" type="radio" name="filter-sort" :value="opt.value" />
            <span>{{ opt.label }}</span>
          </label>
        </div>
      </fieldset>

      <fieldset class="field-group">
        <legend>Order</legend>
        <div class="option-grid option-grid-narrow">
          <label v-for="opt in ORDER_OPTIONS" :key="opt.value" class="option-pill">
            <input v-model="order" type="radio" name="filter-order" :value="opt.value" />
            <span>{{ opt.label }}</span>
          </label>
        </div>
      </fieldset>

      <fieldset class="field-group">
        <legend>Date range</legend>
        <div class="date-row">
          <label class="field">
            <span>From</span>
            <input v-model="fromDate" type="date" />
          </label>
          <label class="field">
            <span>To</span>
            <input v-model="toDate" type="date" />
          </label>
        </div>
      </fieldset>

      <button type="submit" class="btn block ink">Apply filters</button>
    </form>
  </BottomSheet>
</template>

<style scoped>
.filter-form {
  display: flex;
  flex-direction: column;
  gap: var(--space-md);
}

.field-group {
  border: none;
  margin: 0;
  padding: 0;
}

.field-group legend {
  font-size: 0.8125rem;
  font-weight: 600;
  color: var(--muted);
  margin-bottom: var(--space-xs);
}

.option-grid {
  display: flex;
  flex-wrap: wrap;
  gap: var(--space-xs);
}

.option-pill {
  position: relative;
  cursor: pointer;
}

.option-pill input {
  position: absolute;
  opacity: 0;
  pointer-events: none;
}

.option-pill span {
  display: inline-flex;
  align-items: center;
  min-height: var(--touch-min);
  padding: 0 var(--space-sm);
  border: 1px solid var(--hairline);
  border-radius: var(--radius-pill);
  background: var(--surface);
  color: var(--ink);
  transition:
    border-color var(--duration-short) var(--ease-standard),
    background-color var(--duration-short) var(--ease-standard);
}

.option-pill input:checked + span {
  border-color: var(--accent);
  background: color-mix(in srgb, var(--accent) 12%, transparent);
  font-weight: 600;
}

.option-pill input:focus-visible + span {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

.date-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-sm);
}
</style>
