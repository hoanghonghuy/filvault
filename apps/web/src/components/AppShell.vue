<script setup lang="ts">
import { computed, provide, ref } from 'vue'
import { RouterLink, RouterView, useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import StorageBar from '@/components/StorageBar.vue'
import ToastHost from '@/components/ToastHost.vue'
import GlobalConfirm from '@/components/GlobalConfirm.vue'
import GlobalPrompt from '@/components/GlobalPrompt.vue'
import GlobalActionSheet from '@/components/GlobalActionSheet.vue'

const auth = useAuthStore()
const route = useRoute()
const storageBarRef = ref<InstanceType<typeof StorageBar> | null>(null)

provide('reloadStorage', async () => {
  await storageBarRef.value?.reload()
})

const navItems = [
  { to: '/files', label: 'Files', shortLabel: 'Files' },
  { to: '/photos', label: 'Photos', shortLabel: 'Photos' },
  { to: '/trash', label: 'Trash', shortLabel: 'Trash' },
  { to: '/settings', label: 'Settings', shortLabel: 'Settings' },
]

const pageTitle = computed(() => {
  if (route.name === 'album') return 'Album'
  const match = navItems.find((item) => item.to === route.path)
  return match?.label ?? 'Filvault'
})

const showShell = computed(() => auth.isAuthenticated && auth.isVerified)
</script>

<template>
  <ToastHost />
  <GlobalConfirm />
  <GlobalPrompt />
  <GlobalActionSheet />

  <div v-if="showShell" class="shell">
    <aside class="side-nav" aria-label="Main navigation">
      <div class="brand">Filvault</div>
      <nav class="side-links">
        <RouterLink
          v-for="item in navItems"
          :key="item.to"
          :to="item.to"
          class="side-link"
        >
          {{ item.label }}
        </RouterLink>
      </nav>
      <div class="side-user">
        <span class="user-name">{{ auth.user?.displayName }}</span>
      </div>
    </aside>

    <div class="shell-main">
      <header class="header">
        <div class="header-brand mobile-only">Filvault</div>
        <h1 class="header-title">{{ pageTitle }}</h1>
      </header>
      <StorageBar ref="storageBarRef" />
      <main class="main">
        <RouterView />
      </main>
    </div>

    <nav class="bottom-nav" aria-label="Main navigation">
      <RouterLink
        v-for="item in navItems"
        :key="item.to"
        :to="item.to"
        class="bottom-link"
      >
        <span class="bottom-label">{{ item.shortLabel }}</span>
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
  padding: 0 var(--space-md);
  background: var(--canvas);
  border-bottom: 1px solid var(--hairline);
}

.header-brand {
  font-weight: 700;
  font-size: 1rem;
  color: var(--ink);
}

.header-title {
  margin: 0;
  font-size: 1.125rem;
  font-weight: 600;
  letter-spacing: -0.02em;
  color: var(--ink);
}

.mobile-only {
  display: block;
}

.main {
  flex: 1;
  width: 100%;
  max-width: var(--content-max);
  margin: 0 auto;
  padding: var(--space-md);
}

.bottom-nav {
  position: fixed;
  left: 0;
  right: 0;
  bottom: 0;
  z-index: 40;
  display: flex;
  align-items: stretch;
  min-height: var(--bottom-nav-h);
  padding-bottom: env(safe-area-inset-bottom);
  background: var(--canvas);
  border-top: 1px solid var(--hairline);
}

.bottom-link {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: var(--touch-min);
  color: var(--muted);
  font-size: 12px;
  font-weight: 600;
}

.bottom-link.router-link-active {
  color: var(--accent);
}

.bottom-label {
  padding: var(--space-xs) 0;
}

@media (min-width: 768px) {
  .shell {
    flex-direction: row;
  }

  .side-nav {
    display: flex;
    flex-direction: column;
    width: 220px;
    flex-shrink: 0;
    padding: var(--space-lg) var(--space-md);
    background: var(--canvas);
    border-right: 1px solid var(--hairline);
  }

  .brand {
    font-weight: 700;
    font-size: 1.125rem;
    color: var(--ink);
    margin-bottom: var(--space-lg);
  }

  .side-links {
    display: flex;
    flex-direction: column;
    gap: var(--space-xs);
    flex: 1;
  }

  .side-link {
    display: flex;
    align-items: center;
    min-height: var(--touch-min);
    padding: 0 var(--space-sm);
    border-radius: var(--radius-md);
    color: var(--muted);
    font-size: 14px;
    font-weight: 600;
  }

  .side-link.router-link-active {
    color: var(--ink);
    background: var(--surface-soft);
  }

  .side-user {
    padding-top: var(--space-md);
    border-top: 1px solid var(--hairline-soft);
  }

  .user-name {
    font-size: 0.875rem;
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
</style>
