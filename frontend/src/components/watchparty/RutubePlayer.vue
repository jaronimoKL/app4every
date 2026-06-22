<template>
  <div class="rutube-player-container">
    <iframe
      ref="iframeRef"
      :src="embedUrl"
      frameborder="0"
      allow="clipboard-write; autoplay"
      webkitAllowFullScreen
      mozallowfullscreen
      allowFullScreen
      class="rutube-iframe"
    ></iframe>
    <!-- We capture clicks if needed or just rely on postMessage -->
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'

const props = defineProps({
  url: {
    type: String,
    required: true
  }
})

const emit = defineEmits(['local-play', 'local-pause', 'local-seek'])

const iframeRef = ref(null)
let isSyncing = false

function extractRutubeId(url) {
  const match = url.match(/rutube\.ru\/video\/([a-zA-Z0-9]+)/)
  if (match) return match[1]
  const embedMatch = url.match(/rutube\.ru\/play\/embed\/([a-zA-Z0-9]+)/)
  return embedMatch ? embedMatch[1] : null
}

const embedUrl = computed(() => {
  const id = extractRutubeId(props.url)
  return id ? `https://rutube.ru/play/embed/${id}` : props.url
})

function handleMessage(event) {
  if (!event.data || typeof event.data !== 'string') return
  try {
    const data = JSON.parse(event.data)
    if (data.type === 'player:changeState') {
      if (isSyncing) return
      if (data.data.state === 'playing') {
        emit('local-play', data.data.currentTime || 0)
      } else if (data.data.state === 'paused') {
        emit('local-pause', data.data.currentTime || 0)
      }
    }
    if (data.type === 'player:currentTime') {
      // Periodic time updates, could use for drift detection
    }
  } catch (e) {
    // ignore
  }
}

onMounted(() => {
  window.addEventListener('message', handleMessage)
})

onUnmounted(() => {
  window.removeEventListener('message', handleMessage)
})

function sendCommand(type, data = {}) {
  if (iframeRef.value && iframeRef.value.contentWindow) {
    iframeRef.value.contentWindow.postMessage(JSON.stringify({ type, data }), '*')
  }
}

function syncPlay(time) {
  isSyncing = true
  sendCommand('player:setCurrentTime', { time })
  sendCommand('player:play', {})
  setTimeout(() => { isSyncing = false }, 500)
}

function syncPause(time) {
  isSyncing = true
  sendCommand('player:pause', {})
  sendCommand('player:setCurrentTime', { time })
  setTimeout(() => { isSyncing = false }, 500)
}

function syncSeek(time) {
  isSyncing = true
  sendCommand('player:setCurrentTime', { time })
  setTimeout(() => { isSyncing = false }, 500)
}

defineExpose({
  syncPlay,
  syncPause,
  syncSeek
})
</script>

<style scoped>
.rutube-player-container {
  width: 100%;
  height: 100%;
  background: #000;
  border-radius: 12px;
  overflow: hidden;
}

.rutube-iframe {
  width: 100%;
  height: 100%;
  border: none;
}
</style>
