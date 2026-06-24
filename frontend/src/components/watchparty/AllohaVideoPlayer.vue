<template>
  <div class="alloha-player-container">
    <iframe
      :src="cleanUrl"
      class="alloha-iframe"
      frameborder="0"
      allowfullscreen
      allow="autoplay; encrypted-media"
    ></iframe>
  </div>
</template>

<script setup>
import { computed } from 'vue'

const props = defineProps({
  url: {
    type: String,
    required: true
  }
})

const cleanUrl = computed(() => {
  if (!props.url) return ''
  let url = props.url
  if (url.startsWith('//')) {
    url = window.location.protocol + url
  }
  return url
})

// Пустые методы заглушки для совместимости с интерфейсом useWatchParty
function syncPlay() {}
function syncPause() {}
function syncSeek() {}

defineExpose({
  syncPlay,
  syncPause,
  syncSeek
})
</script>

<style scoped>
.alloha-player-container {
  width: 100%;
  height: 100%;
  background: #000;
  border-radius: 12px;
  overflow: hidden;
  position: relative;
  min-height: 400px;
}

.alloha-iframe {
  width: 100%;
  height: 100%;
  border: 0;
  position: absolute;
  top: 0;
  left: 0;
}
</style>
