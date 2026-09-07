<script setup lang="ts">
import { computed, provide, ref } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import StorageBar from '@/components/StorageBar.vue'
import Icon from '@/components/AppIcon.vue'
import ToastHost from '@/components/ToastHost.vue'
import GlobalConfirm from '@/components/GlobalConfirm.vue'
import GlobalPrompt from '@/components/GlobalPrompt.vue'
import GlobalActionSheet from '@/components/GlobalActionSheet.vue'
import { HOME_PATH, PROFILE_PATH, SHELL_NAV, pageTitleForRoute, showStorageBar } from '@/lib/shellNav'
import { userInitials } from '@/lib/userInitials'
import { useI18n } from '@/lib/i18n'

const auth = useAuthStore()
const route = useRoute()
const storageBarRef = ref<InstanceType<typeof StorageBar> | null>(null)
const { t } = useI18n()

provide('reloadStorage', async () => {
  await storageBarRef.value?.reload()
})

const pageTitle = computed(() =>
  pageTitleForRoute(route.path, route.name, {
    overview: t.value.navOverview,
    files: t.value.navFiles,
    photos: t.value.navPhotos,
    trash: t.value.navTrash,
    settings: t.value.navSettings,
    profile: t.value.profile,
    album: t.value.album,
    chat: t.value.chat,
    shared: t.value.sharedWithMe,
  }),
)
const showShell = computed(() => auth.isAuthenticated && auth.isVerified && !route.meta.bare)
const storageVisible = computed(() => showStorageBar(route.path))
const avatarInitials = computed(() =>
  userInitials(auth.user?.displayName ?? '', auth.user?.email ?? ''),
)
const profileLabel = computed(
  () => `Open profile for ${auth.user?.displayName || auth.user?.email || 'account'}`,
)
const navLabels = computed(() => ({
  '/chat': t.value.chat,
  '/files': t.value.navFiles,
  '/photos': t.value.navPhotos,
  '/trash': t.value.navTrash,
  '/settings': t.value.navSettings,
}))
</script>

<template>
  <ToastHost />
  <GlobalConfirm />
  <GlobalPrompt />
  <GlobalActionSheet />

  <div v-if="showShell" class="shell">
    <aside class="side-nav" aria-label="Main navigation">
      <RouterLink :to="HOME_PATH" class="brand" aria-label="Go to overview">
        <span class="header-brand-mark" aria-hidden="true">F</span>
        <span>Filvault</span>
      </RouterLink>
      <nav class="side-links" aria-label="Destinations">
        <RouterLink
          v-for="item in SHELL_NAV"
          :key="item.to"
          :to="item.to"
          class="side-link"
        >
          <span class="nav-indicator" aria-hidden="true" />
          <Icon :name="item.icon" :size="20" />
          <span>{{ navLabels[item.to] }}</span>
        </RouterLink>
      </nav>
      <div class="side-user">
        <RouterLink :to="PROFILE_PATH" class="side-profile" :aria-label="profileLabel">
          <span class="avatar" aria-hidden="true">{{ avatarInitials }}</span>
          <span class="user-name">{{ auth.user?.displayName }}</span>
        </RouterLink>
      </div>
    </aside>

    <div class="shell-main">
      <header class="header">
        <RouterLink :to="HOME_PATH" class="header-home" aria-label="Go to overview">
          <span class="header-brand-mark" aria-hidden="true">F</span>
        </RouterLink>
        <div class="header-text">
          <p class="header-brand-name">Filvault</p>
          <h1 class="header-title">{{ pageTitle }}</h1>
        </div>
        <RouterLink to="/chat" class="header-chat-btn" :aria-label="t.chat">
          <Icon name="chat" :size="20" />
        </RouterLink>
        <RouterLink :to="PROFILE_PATH" class="header-profile" :aria-label="profileLabel">
          <span class="avatar" aria-hidden="true">{{ avatarInitials }}</span>
        </RouterLink>
      </header>
      <StorageBar v-show="storageVisible" ref="storageBarRef" />
      <main class="main">
        <RouterView v-slot="{ Component }">
          <Transition name="page" mode="out-in">
            <component :is="Component" />
          </Transition>
        </RouterView>
      </main>
    </div>

    <nav class="bottom-nav" aria-label="Main navigation">
        <RouterLink
          v-for="item in SHELL_NAV"
          :key="item.to"
          :to="item.to"
          class="bottom-link"
        >
        <span class="bottom-icon">
          <span class="nav-indicator" aria-hidden="true" />
          <Icon :name="item.icon" :size="24" />
        </span>
        <span class="bottom-label">{{ navLabels[item.to] }}</span>
      </RouterLink>
    </nav>
  </div>
  <RouterView v-else />
</template>

<style scoped>
.shell {
  min-height: 100vh;
  min-height: 100dvh;
  display: flex;
  flex-direction: column;
  background: var(--surface-soft);
}

.side-nav {
  display: none;
}

.shell-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  padding-bottom: calc(var(--bottom-nav-h) + env(safe-area-inset-bottom));
}

.header {
  display: flex;
  align-items: center;
  gap: var(--space-sm);
  min-height: var(--header-h);
  padding: calc(var(--space-xs) + env(safe-area-inset-top)) var(--space-md) var(--space-xs);
  background: var(--canvas);
  border-bottom: 1px solid var(--hairline);
}

.header-home {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: var(--touch-min);
  height: var(--touch-min);
  margin-left: -6px;
  border-radius: var(--radius-md);
  color: inherit;
}

.header-home:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

.header-home.router-link-exact-active .header-brand-mark {
  background: var(--accent);
  color: var(--on-accent);
}

.header-brand-mark {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: 32px;
  height: 32px;
  border-radius: var(--radius-md);
  background: var(--accent-soft);
  color: var(--accent-hover);
  font-size: 0.875rem;
  font-weight: 700;
  letter-spacing: -0.02em;
}

.header-text {
  min-width: 0;
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  gap: 1px;
}

.header-chat-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: var(--touch-min);
  height: var(--touch-min);
  border-radius: var(--radius-pill);
  color: var(--ink);
  background: transparent;
  text-decoration: none;
  transition: transform var(--duration-short) var(--ease-standard),
    background var(--duration-short) var(--ease-standard),
    color var(--duration-short) var(--ease-standard);
}

.header-chat-btn:hover {
  background: rgba(0, 132, 255, 0.1);
  color: #0084ff;
  transform: scale(1.05);
}

.header-chat-btn:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

.header-profile {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  width: var(--touch-min);
  height: var(--touch-min);
  margin-right: -6px;
  border-radius: var(--radius-md);
}

.header-profile:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: 2px;
}

.header-profile.router-link-active .avatar {
  background: var(--accent);
  color: var(--on-accent);
}

.avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border-radius: var(--radius-pill);
  background: var(--accent-soft);
  color: var(--accent-hover);
  font-size: 0.75rem;
  font-weight: 700;
  letter-spacing: -0.02em;
}

.header-brand-name {
  margin: 0;
  font-size: 12px;
  font-weight: 600;
  line-height: 1.2;
  color: var(--muted);
}

.header-title {
  margin: 0;
  font-size: 1.25rem;
  font-weight: 600;
  letter-spacing: -0.02em;
  line-height: 1.25;
  color: var(--ink);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.main {
  flex: 1;
  width: 100%;
  max-width: var(--content-max);
  margin: 0 auto;
  padding: var(--space-md);
}

.page-enter-active {
  transition: opacity var(--duration-medium) var(--ease-standard),
    transform var(--duration-medium) var(--ease-standard);
}

.page-leave-active {
  transition: opacity var(--duration-short) var(--ease-standard);
}

.page-enter-from {
  opacity: 0;
  transform: translateY(8px);
}

.page-leave-to {
  opacity: 0;
}

.bottom-nav {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 40;
  display: flex;
  align-items: stretch;
  justify-content: space-between;
  min-height: var(--bottom-nav-h);
  padding: 6px 0 calc(6px + env(safe-area-inset-bottom));
  background: var(--canvas);
  border-top: 1px solid var(--hairline);
}

.bottom-link {
  flex: 1 1 0;
  width: 0;
  min-width: 0;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  min-height: var(--touch-min);
  padding: 0 4px;
  color: var(--muted);
  font-size: 12px;
  font-weight: 600;
  line-height: 1.2;
  transition: color var(--motion-press) var(--ease-standard);
}

.bottom-link:focus-visible {
  outline: 2px solid var(--accent);
  outline-offset: -4px;
  border-radius: var(--radius-md);
}

.bottom-link:active .bottom-icon {
  background: var(--surface-card);
  border-radius: var(--radius-md);
}

.bottom-link.router-link-active {
  color: var(--accent);
}

.bottom-link.router-link-active .nav-indicator {
  transform: translateX(-50%) scaleX(1);
}

.bottom-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  flex-shrink: 0;
}

/* Sliding active indicator — background pill under the icon.
   z-index 0 keeps it in the positioned layer but below the icon,
   which gets position: relative to paint above it. */
.nav-indicator {
  position: absolute;
  left: 50%;
  width: 24px;
  height: 24px;
  border-radius: var(--radius-pill);
  background: var(--accent-soft);
  transform: translateX(-50%) scaleX(0);
  transition: transform var(--duration-medium) var(--ease-emphasized-decelerate);
  z-index: 0;
}

.bottom-icon > svg {
  position: relative;
}

.bottom-link {
  position: relative;
}

.bottom-label {
  display: block;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  text-align: center;
}

@media (min-width: 768px) {
  .shell {
    flex-direction: row;
  }

  .side-nav {
    position: sticky;
    top: 0;
    align-self: flex-start;
    height: 100vh;
    height: 100dvh;
    display: flex;
    flex-direction: column;
    width: 220px;
    flex-shrink: 0;
    padding: var(--space-lg) var(--space-md);
    background: var(--canvas);
    border-right: 1px solid var(--hairline);
    overflow-y: auto;
  }

  .brand {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    font-weight: 700;
    font-size: 1.125rem;
    color: var(--ink);
    margin-bottom: var(--space-lg);
    min-height: var(--touch-min);
  }

  .brand:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
    border-radius: var(--radius-md);
  }

  .brand.router-link-exact-active .header-brand-mark {
    background: var(--accent);
    color: var(--on-accent);
  }

  .side-links {
    display: flex;
    flex-direction: column;
    gap: var(--space-xs);
    flex: 1;
  }

  .side-link {
    position: relative;
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    min-height: var(--touch-min);
    padding: 0 var(--space-sm);
    border-radius: var(--radius-md);
    color: var(--muted);
    font-size: 14px;
    font-weight: 600;
    transition:
      background-color var(--motion-press) var(--ease-standard),
      color var(--motion-press) var(--ease-standard);
  }

  .side-link:hover {
    color: var(--ink);
    background: var(--surface-soft);
  }

  .side-link:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }

  .side-link.router-link-active {
    color: var(--accent);
    background: var(--accent-soft);
  }

  .side-link .nav-indicator {
    left: 0;
    width: 3px;
    height: 20px;
    transform: translateX(0) scaleY(0);
  }

  .side-link.router-link-active .nav-indicator {
    transform: translateX(0) scaleY(1);
  }

  .side-user {
    padding-top: var(--space-md);
    border-top: 1px solid var(--hairline-soft);
  }

  .side-profile {
    display: flex;
    align-items: center;
    gap: var(--space-sm);
    min-height: var(--touch-min);
    padding: 0 var(--space-sm);
    margin: 0 calc(var(--space-sm) * -1);
    border-radius: var(--radius-md);
    color: inherit;
  }

  .side-profile:hover {
    background: var(--surface-soft);
  }

  .side-profile:focus-visible {
    outline: 2px solid var(--accent);
    outline-offset: 2px;
  }

  .side-profile.router-link-active .avatar {
    background: var(--accent);
    color: var(--on-accent);
  }

  .user-name {
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--muted);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .shell-main {
    padding-bottom: 0;
  }

  .header {
    display: none;
  }

  .bottom-nav {
    display: none;
  }

  .main {
    padding: var(--space-lg);
  }
}

@media (prefers-reduced-motion: reduce) {
  .side-link {
    transition: none;
  }
}
</style>
