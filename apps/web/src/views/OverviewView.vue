<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { api, formatBytes } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { useAuthStore } from '@/stores/auth'
import Icon from '@/components/AppIcon.vue'
import PhotoPlaceholder from '@/components/PhotoPlaceholder.vue'
import { mimeIcon } from '@/lib/mimeIcon'
import {
  OVERVIEW_DESTINATIONS,
  recentFilesFromBrowser,
  recentPhotosFromTimeline,
} from '@/lib/shellNav'
import type { Browser, BrowserFile, Timeline, TimelineGroup } from '@/api/types'

const auth = useAuthStore()
const router = useRouter()

const loading = ref(false)
const filesError = ref('')
const photosError = ref('')
const files = ref<BrowserFile[]>([])
const groups = ref<TimelineGroup[]>([])

const recentFiles = computed(() => recentFilesFromBrowser(files.value))
const recentPhotos = computed(() => recentPhotosFromTimeline(groups.value))

const storageLine = computed(() => {
  const user = auth.user
  if (!user) return ''
  return `${formatBytes(user.storageUsed)} of ${formatBytes(user.storageQuota)} used`
})

const bothFailed = computed(() => Boolean(filesError.value && photosError.value))

async function load() {
  loading.value = true
  filesError.value = ''
  photosError.value = ''
  try {
    const [browserResult, timelineResult] = await Promise.allSettled([
      api<Browser>('/browser'),
      api<Timeline>('/photos/timeline'),
    ])
    if (browserResult.status === 'fulfilled') {
      files.value = browserResult.value.files
    } else {
      files.value = []
      filesError.value = formatApiError(browserResult.reason, 'Could not load files')
    }
    if (timelineResult.status === 'fulfilled') {
      groups.value = timelineResult.value.groups
    } else {
      groups.value = []
      photosError.value = formatApiError(timelineResult.reason, 'Could not load photos')
    }
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="overview">
    <h1 class="page-title desktop-only">Overview</h1>
    <header class="hero">
      <p class="hero-title">{{ auth.user?.displayName ?? 'Welcome' }}</p>
      <p v-if="storageLine" class="hero-meta">{{ storageLine }}</p>
    </header>

    <nav class="destinations" aria-label="Open a library">
      <RouterLink
        v-for="item in OVERVIEW_DESTINATIONS"
        :key="item.to"
        :to="item.to"
        class="dest-card"
      >
        <span class="dest-icon">
          <Icon :name="item.icon" :size="22" />
        </span>
        <span class="dest-copy">
          <span class="dest-label">{{ item.label }}</span>
          <span class="dest-hint">{{ item.hint }}</span>
        </span>
      </RouterLink>
    </nav>

    <p v-if="bothFailed" class="error" role="alert">Could not load overview. Try again later.</p>

    <div v-if="loading" class="overview-loading" aria-busy="true" aria-live="polite">
      <div class="destinations">
        <div class="skeleton dest-sk" />
        <div class="skeleton dest-sk" />
      </div>
      <div class="list">
        <div v-for="i in 4" :key="i" class="skeleton sk-row" />
      </div>
    </div>

    <template v-else>
      <section class="section" aria-labelledby="recent-files-heading">
        <div class="section-head">
          <h2 id="recent-files-heading" class="section-title">Recent files</h2>
          <RouterLink class="see-all" to="/files">See all</RouterLink>
        </div>
        <p v-if="filesError" class="section-error">{{ filesError }}</p>
        <div v-else-if="recentFiles.length" class="list">
          <RouterLink
            v-for="file in recentFiles"
            :key="file.id"
            class="row tappable"
            to="/files"
          >
            <span class="name">
              <Icon :name="mimeIcon(file.mimeType)" :size="18" class="row-icon" />
              {{ file.name }}
            </span>
            <span class="meta">{{ formatBytes(file.sizeBytes) }}</span>
          </RouterLink>
        </div>
        <p v-else class="empty-inline">
          No files yet.
          <RouterLink class="empty-link" to="/files">Open My Files</RouterLink>
        </p>
      </section>

      <section class="section" aria-labelledby="recent-photos-heading">
        <div class="section-head">
          <h2 id="recent-photos-heading" class="section-title">Recent photos</h2>
          <RouterLink class="see-all" to="/photos">See all</RouterLink>
        </div>
        <p v-if="photosError" class="section-error">{{ photosError }}</p>
        <div v-else-if="recentPhotos.length" class="grid photos">
          <PhotoPlaceholder
            v-for="item in recentPhotos"
            :key="item.id"
            :mime-type="item.mimeType"
            :name="item.name"
            @click="router.push('/photos')"
          />
        </div>
        <p v-else class="empty-inline">
          No photos yet.
          <RouterLink class="empty-link" to="/photos">Open Photos</RouterLink>
        </p>
      </section>
    </template>
  </div>
</template>

<style scoped>
.hero {
  margin-bottom: var(--space-lg);
}

.hero-title {
  margin: 0;
  font-size: 1.375rem;
  font-weight: 600;
  letter-spacing: -0.02em;
  line-height: 1.25;
  color: var(--ink);
}

.hero-meta {
  margin: 4px 0 0;
  font-size: 12px;
  font-weight: 500;
  line-height: 1.4;
  color: var(--muted);
}

.destinations {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: var(--space-sm);
  margin-bottom: var(--space-xl);
}

.dest-card {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: var(--space-sm);
  min-height: 96px;
  padding: var(--space-md);
  border: 1px solid var(--hairline);
  border-radius: var(--radius-lg);
  background: var(--canvas);
  color: var(--ink);
}

.dest-card:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

@media (hover: hover) {
  .dest-card:hover {
    border-color: var(--accent);
    background: var(--accent-soft);
  }
}

.dest-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: var(--radius-md);
  background: var(--accent-soft);
  color: var(--accent-hover);
}

.dest-copy {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
}

.dest-label {
  font-size: 14px;
  font-weight: 600;
}

.dest-hint {
  font-size: 12px;
  font-weight: 500;
  line-height: 1.4;
  color: var(--muted);
}

.section {
  margin-bottom: var(--space-xl);
}

.section-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-sm);
  min-height: var(--touch-min);
}

.section-head .section-title {
  margin: 0;
}

.see-all {
  flex-shrink: 0;
  min-height: var(--touch-min);
  display: inline-flex;
  align-items: center;
  font-size: 14px;
  font-weight: 600;
  color: var(--accent);
}

.see-all:focus-visible,
.empty-link:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
  border-radius: var(--radius-sm);
}

.row-icon {
  margin-right: 0.4rem;
  vertical-align: -3px;
  color: var(--muted);
}

.empty-inline {
  margin: 0;
  font-size: 14px;
  line-height: 1.45;
  color: var(--muted);
}

.empty-link {
  font-weight: 600;
  color: var(--accent);
}

.section-error {
  margin: 0;
  font-size: 14px;
  color: var(--danger);
}

.overview-loading {
  display: flex;
  flex-direction: column;
  gap: var(--space-xl);
}

.dest-sk {
  min-height: 96px;
  border-radius: var(--radius-lg);
}

.sk-row {
  height: var(--touch-min);
  border-radius: var(--radius-md);
}

.desktop-only {
  display: none;
}

@media (min-width: 768px) {
  .desktop-only {
    display: block;
  }

  .hero-title {
    font-size: 1.75rem;
  }

  .destinations {
    grid-template-columns: repeat(2, minmax(0, 280px));
  }
}

@media (prefers-reduced-motion: reduce) {
  .dest-card {
    transition: none;
  }
}
</style>
