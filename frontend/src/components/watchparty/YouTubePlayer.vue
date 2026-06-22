<template>
  <div class="youtube-player-container">
    <div id="yt-player"></div>
    <div v-if="!isApiReady" class="player-loading">
      Загрузка YouTube плеера...
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'

const props = defineProps({
  url: {
    type: String,
    required: true
  }
})

const emit = defineEmits(['local-play', 'local-pause', 'local-seek'])

const isApiReady = ref(false)
let player = null
let isSyncing = false
let driftCheckInterval = null

function extractVideoId(url) {
  const match = url.match(/(?:youtu\.be\/|youtube\.com\/(?:embed\/|v\/|watch\?v=|watch\?.+&v=))([^&?]+)/)
  return match ? match[1] : null
}

function initYouTube() {
  const videoId = extractVideoId(props.url)
  if (!videoId) return

  player = new window.YT.Player('yt-player', {
    videoId: videoId,
    width: '100%',
    height: '100%',
    playerVars: {
      controls: 1,
      rel: 0,
      modestbranding: 1,
      playsinline: 1
    },
    events: {
      onReady: () => {
        isApiReady.value = true
        startDriftCheck()
      },
      onStateChange: onPlayerStateChange
    }
  })
}

function onPlayerStateChange(event) {
  if (isSyncing) return

  const t = player.getCurrentTime()
  if (event.data === window.YT.PlayerState.PLAYING) {
    emit('local-play', t)
  } else if (event.data === window.YT.PlayerState.PAUSED) {
    emit('local-pause', t)
  }
}

// Sync methods called by useWatchParty via playerRef
async function syncPlay(time) {
  if (!player || !player.playVideo) return
  isSyncing = true
  
  await seekIfNeeded(time)
  
  // Try to play. If autoplay policy blocks it, we might need to mute first.
  try {
    player.playVideo()
  } catch (e) {
    player.mute()
    player.playVideo()
  }
  
  setTimeout(() => { isSyncing = false }, 300)
}

async function syncPause(time) {
  if (!player || !player.pauseVideo) return
  isSyncing = true
  
  player.pauseVideo()
  await seekIfNeeded(time)
  
  setTimeout(() => { isSyncing = false }, 300)
}

async function syncSeek(time) {
  if (!player || !player.seekTo) return
  isSyncing = true
  player.seekTo(time, true)
  setTimeout(() => { isSyncing = false }, 300)
}

async function seekIfNeeded(time) {
  const current = player.getCurrentTime()
  if (Math.abs(current - time) > 2) {
    player.seekTo(time, true)
  }
}

function startDriftCheck() {
  // Periodically emit current time if playing, so we can detect drift? 
  // Watch party sync usually driven by events, but if someone seeks, it triggers pause/play.
  // We can just rely on the user dragging the scrubber, which triggers state change.
}

watch(() => props.url, (newUrl) => {
  const newId = extractVideoId(newUrl)
  if (player && player.loadVideoById && newId) {
    isSyncing = true
    player.loadVideoById(newId)
    player.pauseVideo()
    setTimeout(() => { isSyncing = false }, 500)
  }
})

onMounted(() => {
  // Load YouTube IFrame API if not loaded
  if (!window.YT) {
    const tag = document.createElement('script')
    tag.src = 'https://www.youtube.com/iframe_api'
    const firstScriptTag = document.getElementsByTagName('script')[0]
    firstScriptTag.parentNode.insertBefore(tag, firstScriptTag)
    
    window.onYouTubeIframeAPIReady = () => {
      initYouTube()
    }
  } else if (window.YT && window.YT.Player) {
    initYouTube()
  }
})

onUnmounted(() => {
  if (driftCheckInterval) clearInterval(driftCheckInterval)
  if (player && player.destroy) {
    player.destroy()
  }
})

defineExpose({
  syncPlay,
  syncPause,
  syncSeek
})
</script>

<style scoped>
.youtube-player-container {
  width: 100%;
  height: 100%;
  position: relative;
  background: #000;
  border-radius: 12px;
  overflow: hidden;
}

#yt-player {
  width: 100%;
  height: 100%;
}

.player-loading {
  position: absolute;
  top: 0; left: 0; right: 0; bottom: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  color: rgba(255, 255, 255, 0.7);
  font-family: 'Inter', sans-serif;
}
</style>
