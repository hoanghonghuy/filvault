<script setup lang="ts">
import { onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { api } from '@/api/client'
import PhotoPlaceholder from '@/components/PhotoPlaceholder.vue'
import type { AlbumDetail } from '@/api/types'

const route = useRoute()
const album = ref<AlbumDetail | null>(null)
const error = ref('')

async function load() {
  error.value = ''
  album.value = await api<AlbumDetail>(`/photos/albums/${route.params.id}`)
}

watch(() => route.params.id, load, { immediate: true })
onMounted(load)
</script>

<template>
  <div>
    <h1 class="page-title">{{ album?.name ?? 'Album' }}</h1>
    <p v-if="error" class="error">{{ error }}</p>
    <div class="grid photos">
      <PhotoPlaceholder
        v-for="item in album?.items ?? []"
        :key="item.id"
        :mime-type="item.mimeType"
        :name="item.name"
      />
    </div>
  </div>
</template>
