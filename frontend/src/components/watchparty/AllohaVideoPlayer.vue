<template>
  <div class="alloha-player-container-wrapper">
    <div class="alloha-player-container">
      <iframe
        ref="iframeRef"
        :src="cleanUrl"
        class="alloha-iframe"
        frameborder="0"
        allowfullscreen
        allow="autoplay; encrypted-media"
      ></iframe>
    </div>
    
    <div class="mirror-selector-bar glass">
      <span class="mirror-label">🔗 Зеркало Alloha:</span>
      <select v-model="selectedMirror" class="mirror-select">
        <option v-for="mirror in mirrors" :key="mirror" :value="mirror">
          {{ mirror }}
        </option>
      </select>
      <span class="mirror-tip">⚠️ Смените зеркало, если плеер Alloha выдает ошибку SSL или не загружается</span>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch } from 'vue'

const props = defineProps({
  url: {
    type: String,
    required: true
  }
})

const mirrors = [
  'api.alloha.live',
  'api.allohavn.tv',
  'api.alloha.tv',
  'api.allohacdn.com'
]

function getInitialMirror() {
  const saved = localStorage.getItem('alloha_mirror')
  if (saved && mirrors.includes(saved)) {
    return saved
  }
  
  if (props.url) {
    try {
      let tempUrl = props.url
      if (tempUrl.startsWith('//')) {
        tempUrl = 'https:' + tempUrl
      }
      const parsed = new URL(tempUrl)
      if (parsed.hostname && parsed.hostname !== 'api.alloha.tv') {
        const matched = mirrors.find(m => parsed.hostname.includes(m))
        if (matched) return matched
      }
    } catch (e) {}
  }
  
  return 'api.alloha.live' // Default to working mirror
}

const selectedMirror = ref(getInitialMirror())

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

watch(selectedMirror, (newVal) => {
  localStorage.setItem('alloha_mirror', newVal)
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
.alloha-player-container-wrapper {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  gap: 12px;
}

.alloha-player-container {
  flex: 1;
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
