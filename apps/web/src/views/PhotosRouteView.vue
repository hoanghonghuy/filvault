<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import PhotosView from '@/views/PhotosView.vue'
import VideoPhotosView from '@/views/VideoPhotosView.vue'

const route = useRoute()
const videoOnly = computed(() => route.query.type === 'video')
</script>

<template>
  <VideoPhotosView v-if="videoOnly" />
  <PhotosView v-else />
</template>
