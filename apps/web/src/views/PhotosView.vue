<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '@/api/client'
import PhotoPlaceholder from '@/components/PhotoPlaceholder.vue'
import type { Album, Timeline } from '@/api/types'

const router = useRouter()
const timeline = ref<Timeline | null>(null)
const albums = ref<Album[]>([])
const newAlbumName = ref('')
const error = ref('')

async function load() {
  error.value = ''
  timeline.value = await api<Timeline>('/photos/timeline')
  const list = await api<{ albums: Album[] }>('/photos/albums')
  albums.value = list.albums
}

async function createAlbum() {
  if (!newAlbumName.value.trim()) return
  await api('/photos/albums', {
    method: 'POST',
    body: JSON.stringify({ name: newAlbumName.value.trim() }),
  })
  newAlbumName.value = ''
  await load()
}

onMounted(load)
</script>

<template>
  <div>
    <h1 class="page-title">Photos</h1>
    <p v-if="error" class="error">{{ error }}</p>

    <section class="card" style="margin-bottom: 1rem">
      <h2>Albums</h2>
      <div class="toolbar">
        <input v-model="newAlbumName" placeholder="New album name" />
        <button class="btn primary" type="button" @click="createAlbum">Create album</button>
      </div>
      <div class="list">
        <div v-for="album in albums" :key="album.id" class="row">
          <button class="btn linkish name" type="button" @click="router.push(`/photos/albums/${album.id}`)">
            {{ album.name }}
          </button>
        </div>
      </div>
    </section>

    <section v-for="group in timeline?.groups ?? []" :key="group.date" class="card" style="margin-bottom: 1rem">
      <h2>{{ group.date }}</h2>
      <div class="grid photos">
        <PhotoPlaceholder v-for="item in group.items" :key="item.id" :mime-type="item.mimeType" :name="item.name" />
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
