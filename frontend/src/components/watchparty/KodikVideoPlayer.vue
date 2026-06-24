<template>
  <div class="kodik-player-container-wrapper">
    <div class="kodik-player-container">
      <iframe
        ref="iframeRef"
        :src="cleanUrl"
        class="kodik-iframe"
        frameborder="0"
        allowfullscreen
        allow="autoplay; encrypted-media"
      ></iframe>
    </div>
    
    <div class="mirror-selector-bar glass">
      <span class="mirror-label">🔗 Зеркало плеера (домен):</span>
      <select v-model="selectedMirror" class="mirror-select">
        <option v-for="mirror in mirrors" :key="mirror" :value="mirror">
          {{ mirror }}
        </option>
      </select>
      <span class="mirror-tip">⚠️ Смените зеркало, если плеер не загружается из-за блокировок провайдера</span>
    </div>
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

const emit = defineEmits(['local-play', 'local-pause', 'local-seek', 'local-episode-change'])

const iframeRef = ref(null)
let isSyncing = false
let lastTime = 0
let currentEpisode = null

// Список известных зеркал Kodik (в порядке надёжности)
const mirrors = [
  'kodikplayer.com',  // Основной — всегда работает
  'kodikonline.com',
  'gbase.online',
  'gbase.site',
  'baseofplay.club',
  'kinosko.pro',
  'anivod.pro',
  'aniqit.com',
  'kodik.info',       // Часто заблокирован провайдерами — в конце
]

// Зеркала, которые заведомо не работают — сбрасываем сохранённый выбор
const BROKEN_MIRRORS = ['kodik.info', 'aniqit.com']

// Получаем начальный домен из ссылки или из localStorage
function getInitialMirror() {
  const saved = localStorage.getItem('kodik_mirror')
  if (saved && mirrors.includes(saved) && !BROKEN_MIRRORS.includes(saved)) {
    return saved
  }
  // Если сохранено сломанное зеркало — удаляем
  if (saved && BROKEN_MIRRORS.includes(saved)) {
    localStorage.removeItem('kodik_mirror')
  }
  
  if (props.url) {
    try {
      let tempUrl = props.url
      if (tempUrl.startsWith('//')) {
        tempUrl = 'https:' + tempUrl
      }
      const parsed = new URL(tempUrl)
      if (parsed.hostname) {
        // Если домен из ссылки есть в списке и не сломан — используем его
        const matched = mirrors.find(m => parsed.hostname.includes(m) && !BROKEN_MIRRORS.includes(m))
        if (matched) return matched
      }
    } catch (e) {}
  }
  
  return 'kodikplayer.com' // Надёжный дефолт
}

const selectedMirror = ref(getInitialMirror())

// Гарантируем правильный протокол для iframe src и подменяем домен на выбранное зеркало
const cleanUrl = computed(() => {
  if (!props.url) return ''
  let url = props.url
  if (url.startsWith('//')) {
    url = window.location.protocol + url
  }
  
  try {
    const urlObj = new URL(url)
    urlObj.hostname = selectedMirror.value
    return urlObj.toString()
  } catch (e) {
    return url
  }
})

// Сохраняем выбор зеркала при его изменении
watch(selectedMirror, (newVal) => {
  localStorage.setItem('kodik_mirror', newVal)
})

// Отправка сообщений в плеер Kodik
function sendCommand(method, args = {}) {
  if (iframeRef.value && iframeRef.value.contentWindow) {
    try {
      // 1. Стандартный формат объекта
      iframeRef.value.contentWindow.postMessage({
        key: 'kodik_player_api',
        value: {
          method,
          ...args
        }
      }, '*')

      // 2. Альтернативные строковые сообщения для старых зеркал / плееров
      if (method === 'seek') {
        const val = args.value !== undefined ? args.value : args.time
        if (val !== undefined) {
          iframeRef.value.contentWindow.postMessage(`setCurrentTime=${val}`, '*')
          iframeRef.value.contentWindow.postMessage(`seek=${val}`, '*')
        }
      } else if (method === 'play') {
        iframeRef.value.contentWindow.postMessage('play', '*')
      } else if (method === 'pause') {
        iframeRef.value.contentWindow.postMessage('pause', '*')
      }
    } catch (e) {
      console.error('Failed to send postMessage to Kodik iframe', e)
    }
  }
}

// Слушатель сообщений от плеера Kodik
function onMessage(event) {
  let data
  try {
    data = typeof event.data === 'string' ? JSON.parse(event.data) : event.data
  } catch (e) {
    return
  }

  if (!data || !data.key) return

  switch (data.key) {
    case 'kodik_player_time_update':
      if (typeof data.value === 'number') {
        lastTime = data.value
      }
      break

    case 'kodik_player_play':
      if (isSyncing) return
      emit('local-play', lastTime)
      break

    case 'kodik_player_pause':
      if (isSyncing) return
      emit('local-pause', lastTime)
      break

    case 'kodik_player_current_episode':
      if (data.value && data.value !== currentEpisode) {
        currentEpisode = data.value
        if (!isSyncing) {
          emit('local-episode-change', data.value)
        }
      }
      break
  }
}

// Методы синхронизации для родительского компонента (WatchPartyRoom.vue)
function syncPlay(time) {
  isSyncing = true
  seekIfNeeded(time)
  sendCommand('play')
  setTimeout(() => { isSyncing = false }, 400)
}

function syncPause(time) {
  isSyncing = true
  sendCommand('pause')
  seekIfNeeded(time)
  setTimeout(() => { isSyncing = false }, 400)
}

// Позиционирование
function syncSeek(time) {
  isSyncing = true
  sendCommand('seek', { time: time, value: time })
  lastTime = time
  setTimeout(() => { isSyncing = false }, 400)
}

function seekIfNeeded(time) {
  if (Math.abs(lastTime - time) > 2) {
    sendCommand('seek', { time: time, value: time })
    lastTime = time
  }
}

// Отслеживание смены URL
watch(() => props.url, () => {
  isSyncing = true
  try {
    const urlObj = new URL(cleanUrl.value)
    const ep = urlObj.searchParams.get('episode')
    if (ep) {
      currentEpisode = parseInt(ep, 10)
    }
  } catch (e) {}
  setTimeout(() => { isSyncing = false }, 800)
})

onMounted(() => {
  window.addEventListener('message', onMessage)
  try {
    const urlObj = new URL(cleanUrl.value)
    const ep = urlObj.searchParams.get('episode')
    if (ep) {
      currentEpisode = parseInt(ep, 10)
    }
  } catch (e) {}
})

onUnmounted(() => {
  window.removeEventListener('message', onMessage)
})

defineExpose({
  syncPlay,
  syncPause,
  syncSeek
})
</script>

<style scoped>
.kodik-player-container-wrapper {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  gap: 12px;
}

.kodik-player-container {
  flex: 1;
  background: #000;
  border-radius: 12px;
  overflow: hidden;
  position: relative;
  min-height: 400px;
}

.kodik-iframe {
  width: 100%;
  height: 100%;
  border: 0;
  position: absolute;
  top: 0;
  left: 0;
}

.mirror-selector-bar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 12px;
  padding: 10px 16px;
  border-radius: 12px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  background: rgba(255, 255, 255, 0.02);
  font-size: 0.85rem;
}

.mirror-label {
  color: var(--text-secondary, #94a3b8);
  font-weight: 600;
}

.mirror-select {
  background: rgba(0, 0, 0, 0.4);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: #fff;
  padding: 6px 12px;
  border-radius: 6px;
  outline: none;
  cursor: pointer;
  font-size: 0.85rem;
  transition: all 0.2s ease;
}
.mirror-select:hover {
  border-color: rgba(255, 255, 255, 0.2);
}
.mirror-select:focus {
  border-color: rgba(99, 102, 241, 0.5);
}

.mirror-tip {
  color: var(--text-muted, #475569);
  font-size: 0.75rem;
  margin-left: auto;
}

@media (max-width: 650px) {
  .mirror-tip {
    margin-left: 0;
    width: 100%;
  }
}
</style>
