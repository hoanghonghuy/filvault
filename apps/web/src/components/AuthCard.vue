<script setup lang="ts">
import { useId } from 'vue'

defineProps<{
  title: string
  subtitle?: string
}>()

const titleId = useId()
</script>

<template>
  <div class="auth-page">
    <div class="auth-shell">
      <p class="auth-brand">
        <span class="auth-brand-mark" aria-hidden="true">F</span>
        <span class="auth-brand-name">Filvault</span>
      </p>

      <div class="card auth-card">
        <header class="auth-card-header">
          <h1 :id="titleId">{{ title }}</h1>
          <p v-if="subtitle" class="auth-card-subtitle">{{ subtitle }}</p>
        </header>

        <div class="auth-card-body">
          <slot :title-id="titleId" />
        </div>

        <footer v-if="$slots.footer" class="auth-card-footer">
          <slot name="footer" />
        </footer>
      </div>
    </div>
  </div>
</template>
