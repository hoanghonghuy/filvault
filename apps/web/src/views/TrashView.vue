<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api, formatBytes } from '@/api/client'
import type { TrashList } from '@/api/types'

const trash = ref<TrashList | null>(null)
const error = ref('')

async function load() {
  trash.value = await api<TrashList>('/trash')
}

async function restoreFile(id: string) {
  await api(`/files/${id}/restore`, { method: 'POST', body: '{}' })
  await load()
}

async function restoreFolder(id: string) {
  await api(`/folders/${id}/restore`, { method: 'POST', body: '{}' })
  await load()
}

async function permanentDelete(type: 'files' | 'folders', id: string) {
  if (!window.confirm('Permanently delete? This cannot be undone.')) return
  await api(`/trash/${type}/${id}`, { method: 'DELETE' })
  await load()
}

onMounted(load)
</script>

<template>
  <div>
    <h1 class="page-title">Trash</h1>
    <p v-if="error" class="error">{{ error }}</p>

    <section class="list">
      <h2>Folders</h2>
      <div v-for="folder in trash?.folders ?? []" :key="folder.id" class="row">
        <span class="name">📁 {{ folder.name }}</span>
        <button class="btn" type="button" @click="restoreFolder(folder.id)">Restore</button>
        <button class="btn danger" type="button" @click="permanentDelete('folders', folder.id)">Delete forever</button>
      </div>
    </section>

    <section class="list" style="margin-top: 1rem">
      <h2>Files</h2>
      <div v-for="file in trash?.files ?? []" :key="file.id" class="row">
        <span class="name">{{ file.name }}</span>
        <span class="meta">{{ file.mimeType }} · {{ formatBytes(file.sizeBytes ?? 0) }}</span>
        <button class="btn" type="button" @click="restoreFile(file.id)">Restore</button>
        <button class="btn danger" type="button" @click="permanentDelete('files', file.id)">Delete forever</button>
      </div>
    </section>
  </div>
</template>
