<template>
  <div class="direct-player-container">
    <div v-if="playerError" class="hls-error-banner">
      {{ playerError }}
    </div>
    <video
      ref="videoRef"
      class="html5-video"
      controls
      @play="onPlay"
      @pause="onPause"
      @seeked="onSeeked"
    ></video>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import Hls from 'hls.js'

const props = defineProps({
  url: {
    type: String,
    required: true
  }
})

const emit = defineEmits(['local-play', 'local-pause', 'local-seek'])

const videoRef = ref(null)
const playerError = ref('')
let isSyncing = false
let hlsInstance = null

function setupPlayer() {
  if (!videoRef.value) return
  if (hlsInstance) {
    hlsInstance.destroy()
    hlsInstance = null
  }

  // Strip Anilibria ad parameters and proxy the stream to bypass 403 Forbidden hotlink blocks
  let cleanUrl = props.url
  if (cleanUrl.includes('libria')) {
    cleanUrl = cleanUrl.split('?')[0] // remove query string entirely to bypass ads
    cleanUrl = `/api/v1/reviews/integrations/aniliberty/proxy?url=${encodeURIComponent(cleanUrl)}`
  }

  if (props.url.includes('.m3u8') && Hls.isSupported()) {
    hlsInstance = new Hls({
      debug: false,
      enableWorker: true
    })
    hlsInstance.on(Hls.Events.ERROR, function (event, data) {
      if (data.fatal) {
        playerError.value = `HLS Ошибка: ${data.type} - ${data.details}`
        switch (data.type) {
          case Hls.ErrorTypes.NETWORK_ERROR:
            hlsInstance.startLoad()
            break
          case Hls.ErrorTypes.MEDIA_ERROR:
            hlsInstance.recoverMediaError()
            break
          default:
            hlsInstance.destroy()
            break
        }
      }
    })
    hlsInstance.loadSource(cleanUrl)
    hlsInstance.attachMedia(videoRef.value)
  } else {
    // Native HLS support (Safari) or direct MP4/WebM
    videoRef.value.src = cleanUrl
  }
}

onMounted(() => {
  setupPlayer()
})

watch(() => props.url, () => {
  setupPlayer()
})

onUnmounted(() => {
  if (hlsInstance) {
    hlsInstance.destroy()
  }
})

function onPlay() {
  if (isSyncing) return
  emit('local-play', videoRef.value.currentTime)
}

function onPause() {
  if (isSyncing) return
  // If we just seeked, pause might be triggered. The HTML5 video emits pause before seeking sometimes.
  emit('local-pause', videoRef.value.currentTime)
}

function onSeeked() {
  if (isSyncing) return
  emit('local-seek', videoRef.value.currentTime)
}

function syncPlay(time) {
  if (!videoRef.value) return
  isSyncing = true
  seekIfNeeded(time)
  videoRef.value.play().catch(e => {
    videoRef.value.muted = true
    videoRef.value.play()
  })
  setTimeout(() => { isSyncing = false }, 300)
}

function syncPause(time) {
  if (!videoRef.value) return
  isSyncing = true
  videoRef.value.pause()
  seekIfNeeded(time)
  setTimeout(() => { isSyncing = false }, 300)
}

function syncSeek(time) {
  if (!videoRef.value) return
  isSyncing = true
  videoRef.value.currentTime = time
  setTimeout(() => { isSyncing = false }, 300)
}

function seekIfNeeded(time) {
  const current = videoRef.value.currentTime
  if (Math.abs(current - time) > 2) {
    videoRef.value.currentTime = time
  }
}

defineExpose({
  syncPlay,
  syncPause,
  syncSeek
})
</script>

<style scoped>
.direct-player-container {
  width: 100%;
  height: 100%;
  background: #000;
  border-radius: 12px;
  overflow: hidden;
  display: flex;
  align-items: center;
  justify-content: center;
}

.html5-video {
  width: 100%;
  height: 100%;
  outline: none;
}

.hls-error-banner {
  position: absolute;
  top: 10px; left: 10px; right: 10px;
  background: rgba(239, 68, 68, 0.9);
  color: white;
  padding: 10px;
  border-radius: 8px;
  font-size: 14px;
  text-align: center;
  z-index: 10;
}
</style>
