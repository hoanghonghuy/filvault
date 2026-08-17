<script setup lang="ts">
import { RouterLink, RouterView } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import StorageBar from '@/components/StorageBar.vue'

const auth = useAuthStore()

async function logout() {
  await auth.logout()
  window.location.href = '/login'
}
</script>

<template>
  <div v-if="auth.isAuthenticated && auth.isVerified" class="shell">
    <header class="topbar">
      <div class="brand">Filnest</div>
      <nav>
        <RouterLink to="/files">My Files</RouterLink>
        <RouterLink to="/photos">Photos</RouterLink>
        <RouterLink to="/trash">Trash</RouterLink>
        <RouterLink to="/settings">Settings</RouterLink>
      </nav>
      <div class="user">
        <span>{{ auth.user?.displayName }}</span>
        <button type="button" class="linkish" @click="logout">Logout</button>
      </div>
    </header>
    <StorageBar />
    <main>
      <RouterView />
    </main>
  </div>
  <RouterView v-else />
</template>

<style scoped>
.shell {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.topbar {
  display: flex;
  align-items: center;
  gap: 1rem;
  padding: 0.75rem 1.25rem;
  border-bottom: 1px solid var(--border);
  background: var(--surface);
}

.brand {
  font-weight: 700;
  font-size: 1.1rem;
}

nav {
  display: flex;
  gap: 0.75rem;
  flex: 1;
}

nav a {
  color: var(--muted);
  padding: 0.25rem 0.5rem;
  border-radius: 6px;
}

nav a.router-link-active {
  color: var(--text);
  background: var(--surface-2);
}

.user {
  display: flex;
  align-items: center;
  gap: 0.75rem;
  font-size: 0.9rem;
}

.linkish {
  background: none;
  border: none;
  color: var(--accent);
  cursor: pointer;
  padding: 0;
}

main {
  flex: 1;
  padding: 1.25rem;
  max-width: 1100px;
  width: 100%;
  margin: 0 auto;
}
</style>
