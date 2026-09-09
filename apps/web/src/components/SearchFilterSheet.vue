<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import BottomSheet from '@/components/BottomSheet.vue'
import type { SearchFilters } from '@/api/types'
import { isValidDateRange, resetSearchFilters } from '@/lib/filesSearchState'
import { useI18n } from '@/lib/i18n'

const props = defineProps<{
  open: boolean
  filters: SearchFilters
}>()

const emit = defineEmits<{
  close: []
  apply: [filters: SearchFilters]
}>()

const { t } = useI18n()

const type = ref<SearchFilters['type']>('all')
const sort = ref<SearchFilters['sort']>('relevance')
const order = ref<Exclude<SearchFilters['order'], undefined>>('desc')
const fromDate = ref('')
const toDate = ref('')
const dateError = ref('')

const typeOptions = computed(() => [
  { value: 'all' as const, label: t.value.all },
  { value: 'image' as const, label: t.value.filterTypeImage },
  { value: 'video' as const, label: t.value.filterTypeVideo },
  { value: 'document' as const, label: t.value.filterTypeDocument },
  { value: 'archive' as const, label: t.value.filterTypeArchive },
  { value: 'folder' as const, label: t.value.filterTypeFolder },
])

const sortOptions = computed(() => [
  { value: 'relevance' as const, label: t.value.searchSortRelevance },
  { value: 'name' as const, label: t.value.searchSortName },
  { value: 'date' as const, label: t.value.searchSortDate },
  { value: 'size' as const, label: t.value.searchSortSize },
])

const orderOptions = computed(() => [
  { value: 'desc' as const, label: t.value.orderDescending },
  { value: 'asc' as const, label: t.value.orderAscending },
])

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
    dateError.value = ''
  },
)

watch([fromDate, toDate], () => {
  if (!fromDate.value || !toDate.value) {
    dateError.value = ''
    return
  }
  dateError.value = isValidDateRange(fromDate.value, toDate.value) ? '' : t.value.dateRangeInvalid
})

function onReset() {
  const defaults = resetSearchFilters()
  type.value = defaults.type
  sort.value = defaults.sort
  order.value = defaults.order ?? 'desc'
  fromDate.value = ''
  toDate.value = ''
  dateError.value = ''
}

function onApply() {
  if (!isValidDateRange(fromDate.value, toDate.value)) {
    dateError.value = t.value.dateRangeInvalid
    return
  }
  emit('apply', {
    type: type.value,
    sort: sort.value,
    order: order.value,
    ...(fromDate.value ? { from: fromDate.value } : {}),
    ...(toDate.value ? { to: toDate.value } : {}),
    ...(props.filters.folderId ? { folderId: props.filters.folderId } : {}),
  })
}
</script>

<template>
  <BottomSheet :open="open" :title="t.searchFiltersTitle" @close="emit('close')">
    <form class="filter-form" @submit.prevent="onApply">
      <fieldset class="field-group">
        <legend>{{ t.filterByType }}</legend>
        <div class="option-grid">
          <label v-for="opt in typeOptions" :key="opt.value" class="option-pill">
            <input v-model="type" type="radio" name="filter-type" :value="opt.value" />
            <span>{{ opt.label }}</span>
          </label>
        </div>
      </fieldset>

      <fieldset class="field-group">
        <legend>{{ t.searchSortBy }}</legend>
        <div class="option-grid">
          <label v-for="opt in sortOptions" :key="opt.value" class="option-pill">
            <input v-model="sort" type="radio" name="filter-sort" :value="opt.value" />
            <span>{{ opt.label }}</span>
          </label>
        </div>
      </fieldset>

      <fieldset class="field-group">
        <legend>{{ t.searchOrder }}</legend>
        <div class="option-grid option-grid-narrow">
          <label v-for="opt in orderOptions" :key="opt.value" class="option-pill">
            <input v-model="order" type="radio" name="filter-order" :value="opt.value" />
            <span>{{ opt.label }}</span>
          </label>
        </div>
      </fieldset>

      <fieldset class="field-group">
        <legend>{{ t.dateRange }}</legend>
        <div class="date-row">
          <label class="field">
            <span>{{ t.dateFrom }}</span>
            <input v-model="fromDate" type="date" :aria-invalid="dateError ? 'true' : undefined" />
          </label>
          <label class="field">
            <span>{{ t.dateTo }}</span>
            <input v-model="toDate" type="date" :aria-invalid="dateError ? 'true' : undefined" />
          </label>
        </div>
        <p v-if="dateError" class="date-error" role="alert">{{ dateError }}</p>
      </fieldset>

      <div class="filter-actions">
        <button type="button" class="btn" @click="onReset">{{ t.resetFilters }}</button>
        <button type="submit" class="btn block ink" :disabled="Boolean(dateError)">{{ t.applyFilters }}</button>
      </div>
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

.date-error {
  margin: var(--space-xs) 0 0;
  font-size: 0.8125rem;
  color: var(--danger);
}

.filter-actions {
  display: flex;
  flex-direction: column;
  gap: var(--space-sm);
}

@media (min-width: 480px) {
  .filter-actions {
    flex-direction: row;
  }

  .filter-actions .btn.block {
    flex: 1;
  }
}
</style>
