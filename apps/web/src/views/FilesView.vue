<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, formatBytes, uploadToPresigned, ApiError } from '@/api/client'
import type { Browser, DownloadURL, SearchResult, UploadSession } from '@/api/types'

const route = useRoute()
const router = useRouter()

const browser = ref<Browser | null>(null)
const searchResults = ref<SearchResult | null>(null)
const error = ref('')
const uploadProgress = ref<number | null>(null)
const searchQuery = ref('')
const newFolderName = ref('')

const folderId = computed(() => {
  const raw = route.query.folderId
  return typeof raw === 'string' && raw ? raw : null
})

async function loadBrowser() {
  error.value = ''
  searchResults.value = null
  try {
    const q = folderId.value ? `?folderId=${folderId.value}` : ''
    browser.value = await api<Browser>(`/browser${q}`)
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : 'Failed to load files'
  }
}

async function runSearch() {
  if (!searchQuery.value.trim()) {
    searchResults.value = null
    return
  }
  searchResults.value = await api<SearchResult>(`/search?q=${encodeURIComponent(searchQuery.value.trim())}`)
}

async function openFolder(id: string | null) {
  await router.push(id ? { path: '/files', query: { folderId: id } } : { path: '/files' })
}

async function createFolder() {
  if (!newFolderName.value.trim()) return
  await api('/folders', {
    method: 'POST',
    body: JSON.stringify({ name: newFolderName.value.trim(), parentId: folderId.value }),
  })
  newFolderName.value = ''
  await loadBrowser()
}

async function onUploadChange(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  input.value = ''
  if (!file) return
  error.value = ''
  uploadProgress.value = 0
  try {
    const session = await api<UploadSession>('/files/upload-sessions', {
      method: 'POST',
      body: JSON.stringify({
        name: file.name,
        size: file.size,
        contentType: file.type || 'application/octet-stream',
        folderId: folderId.value,
      }),
    })
    await uploadToPresigned(session.uploadUrl, file, file.type || 'application/octet-stream', (p) => {
      uploadProgress.value = p
    })
    await api(`/files/${session.fileId}/complete`, { method: 'POST', body: '{}' })
    await loadBrowser()
  } catch (e) {
    error.value = e instanceof ApiError ? e.message : 'Upload failed'
  } finally {
    uploadProgress.value = null
  }
}

async function downloadFile(id: string) {
  const out = await api<DownloadURL>(`/files/${id}/download`)
  window.open(out.downloadUrl, '_blank', 'noopener')
}

async function renameFile(id: string, current: string) {
  const name = window.prompt('New file name', current)
  if (!name || name === current) return
  await api(`/files/${id}`, { method: 'PATCH', body: JSON.stringify({ name }) })
  await loadBrowser()
}

async function deleteFile(id: string) {
  if (!window.confirm('Move file to trash?')) return
  await api(`/files/${id}`, { method: 'DELETE' })
  await loadBrowser()
}

async function deleteFolder(id: string) {
  if (!window.confirm('Move folder to trash?')) return
  await api(`/folders/${id}`, { method: 'DELETE' })
  await loadBrowser()
}

watch(() => route.query.folderId, loadBrowser, { immediate: true })
onMounted(loadBrowser)
</script>

<template>
  <div>
    <h1 class="page-title">My Files</h1>

    <div class="breadcrumb">
      <a href="#" @click.prevent="openFolder(null)">Root</a>
      <template v-for="item in browser?.breadcrumb ?? []" :key="item.id">
        <span>/</span>
        <a href="#" @click.prevent="openFolder(item.id)">{{ item.name }}</a>
      </template>
      <template v-if="browser?.folder">
        <span>/</span>
        <span class="current">{{ browser.folder.name }}</span>
      </template>
    </div>

    <div class="toolbar">
      <label class="btn primary">
        Upload
        <input type="file" hidden @change="onUploadChange" />
      </label>
      <input v-model="newFolderName" placeholder="New folder name" />
      <button class="btn" type="button" @click="createFolder">Create folder</button>
      <input v-model="searchQuery" placeholder="Search by name" @keyup.enter="runSearch" />
      <button class="btn" type="button" @click="runSearch">Search</button>
      <button v-if="searchResults" class="btn" type="button" @click="searchResults = null; searchQuery = ''">
        Clear search
      </button>
    </div>

    <p v-if="uploadProgress !== null" class="muted">Uploading… {{ Math.round(uploadProgress * 100) }}%</p>
    <p v-if="error" class="error">{{ error }}</p>

    <section v-if="searchResults" class="list">
      <h2>Search results</h2>
      <div v-for="folder in searchResults.folders" :key="folder.id" class="row">
        <span class="name">📁 {{ folder.name }}</span>
        <button class="btn" type="button" @click="openFolder(folder.id)">Open</button>
      </div>
      <div v-for="file in searchResults.files" :key="file.id" class="row">
        <span class="name">{{ file.name }}</span>
        <span class="meta">{{ formatBytes(file.sizeBytes) }}</span>
        <button class="btn" type="button" @click="downloadFile(file.id)">Download</button>
      </div>
    </section>

    <section v-else class="list">
      <div
        v-for="folder in browser?.folders ?? []"
        :key="folder.id"
        class="row"
      >
        <button class="btn linkish name" type="button" @click="openFolder(folder.id)">📁 {{ folder.name }}</button>
        <button class="btn danger" type="button" @click="deleteFolder(folder.id)">Delete</button>
      </div>
      <div v-for="file in browser?.files ?? []" :key="file.id" class="row">
        <span class="name">{{ file.name }}</span>
        <span class="meta">{{ file.mimeType }} · {{ formatBytes(file.sizeBytes) }}</span>
        <button class="btn" type="button" @click="downloadFile(file.id)">Download</button>
        <button class="btn" type="button" @click="renameFile(file.id, file.name)">Rename</button>
        <button class="btn danger" type="button" @click="deleteFile(file.id)">Delete</button>
      </div>
    </section>
  </div>
</template>

<style scoped>
.linkish {
  background: none;
  border: none;
  text-align: left;
  cursor: pointer;
}
</style>
