<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { RouterLink, useRouter } from 'vue-router'
import { api, formatBytes } from '@/api/client'
import { formatApiError } from '@/api/errors'
import { useAuthStore } from '@/stores/auth'
import Icon from '@/components/AppIcon.vue'
import OverviewStorageCard from '@/components/OverviewStorageCard.vue'
import PhotoThumb from '@/components/PhotoThumb.vue'
import { mimeIcon } from '@/lib/mimeIcon'
import { recentFilesFromBrowser, recentPhotosFromTimeline } from '@/lib/shellNav'
import { useI18n } from '@/lib/i18n'
import type { Browser, BrowserFile, FavoriteFile, Timeline, TimelineGroup } from '@/api/types'

const auth = useAuthStore()
const router = useRouter()
const { t } = useI18n()

const loading = ref(false)
const filesError = ref('')
const photosError = ref('')
const files = ref<BrowserFile[]>([])
const groups = ref<TimelineGroup[]>([])
const favorites = ref<FavoriteFile[]>([])

const searchQuery = ref('')
const activeTab = ref<'recent' | 'favorites' | 'photos'>('recent')

const recentFiles = computed(() => recentFilesFromBrowser(files.value))
const recentPhotos = computed(() => recentPhotosFromTimeline(groups.value))

const quickCategories = computed(() => [
  { to: '/files', label: t.value.myFiles, icon: 'folder', color: '#f59e0b' },
  { to: '/photos', label: t.value.navPhotos, icon: 'photos', color: '#10b981' },
  { to: { path: '/photos', query: { type: 'video' } }, label: t.value.videos || 'Video', icon: 'video', color: '#8b5cf6' },
  { to: { path: '/files', query: { view: 'favorites' } }, label: t.value.tabFavorites, icon: 'star', color: '#eab308' },
  { to: '/shared', label: t.value.sharedWithMe, icon: 'users', color: '#06b6d4' },
  { to: '/vault', label: t.value.personalVault, icon: 'lock', color: '#6366f1' },
  { to: '/trash', label: t.value.navTrash, icon: 'trash', color: '#ef4444' },
  { to: '/settings', label: t.value.navSettings, icon: 'settings', color: '#64748b' },
])

const seeAllRoute = computed(() => {
  if (activeTab.value === 'favorites') return { path: '/files', query: { view: 'favorites' } }
  if (activeTab.value === 'photos') return '/photos'
  return '/files'
})

function onSearchSubmit() {
  const q = searchQuery.value.trim()
  if (q) router.push({ path: '/files', query: { q } })
  else router.push('/files')
}

const bothFailed = computed(() => Boolean(filesError.value && photosError.value))

async function load() {
  loading.value = true
  filesError.value = ''
  photosError.value = ''
  try {
    const [browserResult, timelineResult, favoritesResult] = await Promise.allSettled([
      api<Browser>('/browser'),
      api<Timeline>('/photos/timeline'),
      api<{ files: FavoriteFile[] }>('/files/favorites?limit=5'),
    ])
    if (browserResult.status === 'fulfilled') files.value = browserResult.value.files
    else {
      files.value = []
      filesError.value = formatApiError(browserResult.reason, 'Could not load files')
    }
    if (timelineResult.status === 'fulfilled') groups.value = timelineResult.value.groups
    else {
      groups.value = []
      photosError.value = formatApiError(timelineResult.reason, 'Could not load photos')
    }
    if (favoritesResult.status === 'fulfilled') favorites.value = favoritesResult.value.files
    else favorites.value = []
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="overview">
    <header class="overview-header">
      <div class="user-greeting">
        <h1 class="hero-title">{{ auth.user?.displayName ?? t.welcome }}</h1>
      </div>

      <form class="search-bar" role="search" @submit.prevent="onSearchSubmit">
        <Icon name="search" :size="18" class="search-icon" />
        <input
          v-model="searchQuery"
          type="search"
          class="search-input"
          :placeholder="t.searchInFilvault"
          :aria-label="t.searchInFilvault"
        />
        <button
          v-if="searchQuery"
          type="button"
          class="clear-search-btn"
          aria-label="Xóa tìm kiếm"
          @click="searchQuery = ''"
        >
          <Icon name="close" :size="16" />
        </button>
      </form>
    </header>

    <OverviewStorageCard />

    <section class="categories-section" aria-label="Danh mục tính năng">
      <div class="categories-grid">
        <RouterLink
          v-for="item in quickCategories"
          :key="item.label"
          :to="item.to"
          class="category-btn tappable"
        >
          <div class="category-icon-box" :style="{ '--cat-color': item.color }">
            <Icon :name="item.icon" :size="22" />
          </div>
          <span class="category-label">{{ item.label }}</span>
        </RouterLink>
      </div>
    </section>

    <p v-if="bothFailed" class="error" role="alert">Could not load overview. Try again later.</p>

    <div v-if="loading" class="overview-loading" aria-busy="true" aria-live="polite">
      <div class="skeleton sk-storage" />
      <div class="skeleton-grid">
        <div v-for="i in 8" :key="i" class="skeleton sk-cat" />
      </div>
      <div class="list">
        <div v-for="i in 4" :key="i" class="skeleton sk-row" />
      </div>
    </div>

    <section v-else class="content-tabs-section">
      <div class="tabs-header">
        <div class="tabs-pill-list" role="tablist">
          <button
            type="button"
            role="tab"
            :aria-selected="activeTab === 'recent'"
            class="tab-pill"
            :class="{ active: activeTab === 'recent' }"
            @click="activeTab = 'recent'"
          >
            {{ t.tabRecent }}
          </button>
          <button
            type="button"
            role="tab"
            :aria-selected="activeTab === 'favorites'"
            class="tab-pill"
            :class="{ active: activeTab === 'favorites' }"
            @click="activeTab = 'favorites'"
          >
            {{ t.tabFavorites }}
            <span v-if="favorites.length" class="tab-badge">{{ favorites.length }}</span>
          </button>
          <button
            type="button"
            role="tab"
            :aria-selected="activeTab === 'photos'"
            class="tab-pill"
            :class="{ active: activeTab === 'photos' }"
            @click="activeTab = 'photos'"
          >
            {{ t.tabPhotos }}
          </button>
        </div>

        <RouterLink :to="seeAllRoute" class="see-all">
          {{ t.seeAll }}
          <Icon name="chevron-right" :size="14" />
        </RouterLink>
      </div>

      <div v-show="activeTab === 'recent'" class="tab-pane">
        <p v-if="filesError" class="section-error" role="alert">{{ filesError }}</p>
        <div v-else-if="recentFiles.length" class="list">
          <RouterLink v-for="file in recentFiles" :key="file.id" class="row tappable" to="/files">
            <span class="name">
              <Icon :name="mimeIcon(file.mimeType)" :size="18" class="row-icon" />
              <span class="file-name-text">{{ file.name }}</span>
            </span>
            <span class="meta">{{ formatBytes(file.sizeBytes) }}</span>
          </RouterLink>
        </div>
        <p v-else class="empty-inline">
          {{ t.noFilesYet }}
          <RouterLink class="empty-link" to="/files">{{ t.openMyFiles }}</RouterLink>
        </p>
      </div>

      <div v-show="activeTab === 'favorites'" class="tab-pane">
        <div v-if="favorites.length" class="list">
          <RouterLink v-for="file in favorites" :key="file.id" class="row tappable" to="/files">
            <span class="name">
              <Icon name="star-filled" :size="18" class="row-icon star-icon" />
              <span class="file-name-text">{{ file.name }}</span>
            </span>
            <span class="meta">{{ formatBytes(file.sizeBytes) }}</span>
          </RouterLink>
        </div>
        <p v-else class="empty-inline">
          {{ t.noFavorites }}
          <RouterLink class="empty-link" to="/files">{{ t.openMyFiles }}</RouterLink>
        </p>
      </div>

      <div v-show="activeTab === 'photos'" class="tab-pane">
        <p v-if="photosError" class="section-error" role="alert">{{ photosError }}</p>
        <div v-else-if="recentPhotos.length" class="grid photos">
          <PhotoThumb
            v-for="item in recentPhotos"
            :key="item.id"
            :mime-type="item.mimeType"
            :name="item.name"
            :thumbnail-url="item.thumbnailUrl"
            @click="router.push('/photos')"
          />
        </div>
        <p v-else class="empty-inline">
          {{ t.noPhotosYet }}
          <RouterLink class="empty-link" to="/photos">{{ t.openPhotos }}</RouterLink>
        </p>
      </div>
    </section>

    <RouterLink to="/files" class="overview-fab" title="Mở Tệp" aria-label="Mở Tệp">
      <Icon name="plus" :size="24" />
    </RouterLink>
  </div>
</template>

<style scoped>
.overview { position: relative; padding-bottom: 80px; }
.overview-header { display: flex; flex-direction: column; gap: var(--space-sm); margin-bottom: var(--space-md); }
.hero-title { margin: 0; font-size: 1.5rem; font-weight: 700; letter-spacing: -0.02em; color: var(--ink); }
.search-bar { display: flex; align-items: center; gap: 10px; background: var(--surface-soft, rgba(255,255,255,.05)); border: 1px solid var(--hairline); border-radius: 9999px; padding: 10px 16px; transition: border-color .2s ease, box-shadow .2s ease; }
.search-bar:focus-within { border-color: var(--accent); box-shadow: 0 0 0 3px var(--accent-soft); }
.search-icon { color: var(--muted); flex-shrink: 0; }
.search-input { border: none; background: transparent; outline: none; width: 100%; font-size: 14px; color: var(--ink); }
.search-input::placeholder { color: var(--muted); }
.clear-search-btn { background: transparent; border: none; cursor: pointer; color: var(--muted); padding: 0; display: flex; align-items: center; justify-content: center; }
.categories-section { margin-bottom: var(--space-xl); }
.categories-grid { display: grid; grid-template-columns: repeat(4,1fr); gap: 14px 8px; text-align: center; }
.category-btn { display: flex; flex-direction: column; align-items: center; gap: 6px; text-decoration: none; color: var(--ink); user-select: none; touch-action: manipulation; }
.category-icon-box { width: 52px; height: 52px; border-radius: 16px; background: color-mix(in srgb, var(--cat-color) 12%, transparent); color: var(--cat-color); display: flex; align-items: center; justify-content: center; transition: transform .15s ease, filter .15s ease; }
.category-btn:active .category-icon-box { transform: scale(.92); }
.category-label { font-size: 12px; font-weight: 500; line-height: 1.3; color: var(--ink); display: -webkit-box; -webkit-line-clamp: 1; -webkit-box-orient: vertical; overflow: hidden; }
.content-tabs-section { margin-bottom: var(--space-xl); }
.tabs-header { display: flex; align-items: center; justify-content: space-between; margin-bottom: var(--space-md); border-bottom: 1px solid var(--hairline); padding-bottom: 6px; }
.tabs-pill-list { display: flex; align-items: center; gap: 16px; }
.tab-pill { background: transparent; border: none; padding: 6px 0; font-size: 15px; font-weight: 600; color: var(--muted); cursor: pointer; position: relative; transition: color .15s ease; display: inline-flex; align-items: center; gap: 4px; }
.tab-pill.active { color: var(--ink); }
.tab-pill.active::after { content: ''; position: absolute; bottom: -7px; left: 0; right: 0; height: 3px; border-radius: 3px; background: var(--accent); }
.tab-badge { font-size: 11px; padding: 1px 6px; border-radius: 9999px; background: var(--accent-soft); color: var(--accent); font-weight: 600; }
.see-all { display: inline-flex; align-items: center; gap: 2px; font-size: 13px; font-weight: 600; color: var(--accent); text-decoration: none; }
.tab-pane .row { display: flex; align-items: center; justify-content: space-between; padding: 12px 14px; border-radius: var(--radius-md); border: 1px solid var(--hairline); margin-bottom: 8px; background: var(--canvas); text-decoration: none; color: var(--ink); transition: background-color .15s ease, border-color .15s ease; }
.tab-pane .row:active { background: var(--accent-soft); border-color: var(--accent); }
.tab-pane .name { display: flex; align-items: center; gap: 8px; min-width: 0; flex: 1; }
.file-name-text { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; font-size: 14px; font-weight: 500; }
.tab-pane .meta { font-size: 12px; color: var(--muted); flex-shrink: 0; margin-left: 12px; font-variant-numeric: tabular-nums; }
.row-icon { flex-shrink: 0; color: var(--muted); }
.star-icon { color: var(--warning, #eab308); }
.grid.photos { display: grid; grid-template-columns: repeat(3,1fr); gap: 6px; border-radius: var(--radius-md); overflow: hidden; }
.empty-inline { margin: var(--space-md) 0; font-size: 14px; color: var(--muted); text-align: center; }
.empty-link { font-weight: 600; color: var(--accent); margin-left: 4px; }
.section-error { margin: var(--space-sm) 0; font-size: 14px; color: var(--danger); }
.overview-fab { position: fixed; right: 20px; bottom: 84px; width: 52px; height: 52px; border-radius: 50%; background: var(--accent, #0084ff); color: #fff; display: flex; align-items: center; justify-content: center; box-shadow: 0 8px 24px color-mix(in srgb, var(--accent) 35%, transparent); z-index: 50; text-decoration: none; transition: transform .15s ease, box-shadow .15s ease; touch-action: manipulation; }
.overview-fab:active { transform: scale(.92); }
.overview-loading { display: flex; flex-direction: column; gap: var(--space-lg); }
.sk-storage { height: 76px; border-radius: var(--radius-lg, 16px); }
.skeleton-grid { display: grid; grid-template-columns: repeat(4,1fr); gap: 14px 8px; }
.sk-cat { height: 72px; border-radius: 16px; }
.sk-row { height: 48px; border-radius: var(--radius-md); }
@media (min-width: 768px) {
  .overview-header { flex-direction: row; align-items: center; justify-content: space-between; }
  .search-bar { max-width: 360px; width: 100%; }
  .categories-grid { grid-template-columns: repeat(8,1fr); gap: 16px; }
  .grid.photos { grid-template-columns: repeat(6,1fr); gap: 8px; }
  .overview-fab { bottom: 32px; right: 32px; }
}
</style>
