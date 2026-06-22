<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-box glass" style="max-width: 500px; width: 100%;">
      <div class="modal-header">
        <h3 class="modal-title">Выберите эпизод</h3>
        <button class="modal-close" @click="$emit('close')">✕</button>
      </div>

      <div class="modal-body" style="padding: 20px;">
        <div v-if="loading" class="text-center py-6 text-secondary">
          Загрузка эпизодов...
        </div>
        <div v-else-if="episodes.length === 0" class="text-center py-6 text-secondary">
          Эпизоды не найдены
        </div>
        <div v-else class="flex flex-col gap-4">
          
          <!-- Поиск по номеру серии, если больше 26 эпизодов -->
          <div v-if="episodes.length > 26" class="form-group mb-2">
            <input 
              v-model="searchQuery" 
              type="text" 
              class="form-input" 
              placeholder="Поиск по номеру серии..." 
            />
          </div>

          <div class="episodes-list">
            <button 
              v-for="ep in filteredEpisodes" 
              :key="ep.id"
              class="ep-row-btn"
              :class="{ 'active': selectedEpisode && selectedEpisode.id === ep.id }"
              @click="selectEpisode(ep)"
            >
              <div class="ep-num">Эпизод {{ ep.ordinal ?? ep.number }}</div>
              <div class="ep-name">{{ ep.name || ep.name_english || `Серия ${ep.ordinal ?? ep.number}` }}</div>
            </button>
          </div>

          <!-- Выбор качества / плеера если эпизод выбран -->
          <div v-if="selectedEpisode" class="mt-2 pt-4 border-t border-white/10">
            <!-- Кнопка для запуска Kodik плеера (приоритетный вариант с озвучками) -->
            <div v-if="externalPlayerUrl" class="mb-4">
              <button 
                class="btn btn-primary w-full py-3"
                style="justify-content: center;"
                @click="selectKodik(selectedEpisode.ordinal ?? selectedEpisode.number)"
              >
                🎬 Смотреть через Kodik (Все озвучки)
              </button>
            </div>

            <h4 class="font-bold text-sm mb-3 text-white">Прямой поток (Anilibria):</h4>
            <div class="flex flex-wrap gap-2">
              <button v-if="selectedEpisode.hls" class="btn btn-outline btn-sm" @click="selectQuality(selectedEpisode.hls)">По умолчанию</button>
              <button v-if="selectedEpisode.hls_1080" class="btn btn-outline btn-sm" @click="selectQuality(selectedEpisode.hls_1080)">1080p</button>
              <button v-if="selectedEpisode.hls_720" class="btn btn-outline btn-sm" @click="selectQuality(selectedEpisode.hls_720)">720p</button>
              <button v-if="selectedEpisode.hls_480" class="btn btn-outline btn-sm" @click="selectQuality(selectedEpisode.hls_480)">480p</button>
              
              <!-- Запасной вариант -->
              <template v-if="!selectedEpisode.hls && !selectedEpisode.hls_1080 && !selectedEpisode.hls_720 && !selectedEpisode.hls_480">
                <button 
                  v-for="(val, key) in hlsLinks(selectedEpisode)" 
                  :key="key"
                  class="btn btn-outline btn-sm"
                  @click="selectQuality(val)"
                >
                  {{ formatKey(key) }}
                </button>
              </template>
            </div>
            
            <div v-if="!externalPlayerUrl && !selectedEpisode.hls && !selectedEpisode.hls_1080 && !selectedEpisode.hls_720 && !selectedEpisode.hls_480 && Object.keys(hlsLinks(selectedEpisode)).length === 0" class="text-sm text-red-400 mt-2">
              К сожалению, видео-ссылок не найдено.
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, nextTick } from 'vue'
import { useAuthStore } from '@/stores/auth'

const props = defineProps({
  alias: String
})

const emit = defineEmits(['close', 'select'])

const episodes = ref([])
const externalPlayerUrl = ref('')
const loading = ref(true)
const selectedEpisode = ref(null)
const searchQuery = ref('')

onMounted(async () => {
  if (props.alias) {
    const auth = useAuthStore()
    try {
      const res = await fetch(`/api/v1/reviews/integrations/aniliberty/episodes/${props.alias}`, {
        headers: {
          'Authorization': `Bearer ${auth.accessToken}`
        }
      })
      const data = await res.json()
      
      // Запоминаем ссылку на внешний плеер Kodik из релиза
      externalPlayerUrl.value = data.external_player || ''
      
      // Загружаем список эпизодов
      episodes.value = data.episodes || []
      episodes.value.sort((a, b) => parseFloat(a.number) - parseFloat(b.number))
    } catch (e) {
      console.error("Failed to fetch AniLiberty episodes", e)
    }
  }
  loading.value = false
})

const filteredEpisodes = computed(() => {
  if (episodes.value.length > 26 && searchQuery.value) {
    const q = searchQuery.value.trim().toLowerCase()
    return episodes.value.filter(ep => {
      const num = (ep.ordinal ?? ep.number) || ''
      return num.toString().includes(q)
    })
  }
  return episodes.value
})

function selectEpisode(ep) {
  selectedEpisode.value = ep
  
  // Авто-прокрутка вниз к выбору качества
  nextTick(() => {
    const modalBox = document.querySelector('.modal-body')
    if (modalBox) {
      modalBox.scrollTop = modalBox.scrollHeight
    }
  })
}

function selectQuality(url) {
  emit('select', url)
}

function selectKodik(episodeNum) {
  let url = externalPlayerUrl.value
  if (!url) return
  
  // Приводим к абсолютному протоколу
  if (url.startsWith('//')) {
    url = window.location.protocol + url
  }
  
  // Добавляем параметр episode
  const separator = url.includes('?') ? '&' : '?'
  url = `${url}${separator}episode=${episodeNum}`
  
  emit('select', url)
}

function hlsLinks(ep) {
  const links = {}
  for (const [key, val] of Object.entries(ep)) {
    if (typeof val === 'string' && val.includes('.m3u8')) {
      if (!['hls', 'hls_1080', 'hls_720', 'hls_480'].includes(key)) {
        links[key] = val
      }
    }
  }
  if (Object.keys(links).length === 0) {
    if (ep.url && typeof ep.url === 'string') links['Ссылка'] = ep.url;
    if (ep.video && typeof ep.video === 'string') links['Видео'] = ep.video;
    if (ep.source && typeof ep.source === 'string') links['Источник'] = ep.source;
    if (ep.player && typeof ep.player === 'string') links['Плеер'] = ep.player;
  }
  return links
}

function formatKey(key) {
  return key.replace(/_/g, ' ').toUpperCase()
}
</script>

<style scoped>
/* ══ Модал ══ */
.modal-overlay {
  position: fixed; inset: 0;
  background: rgba(0,0,0,0.6); backdrop-filter: blur(6px);
  display: flex; align-items: center; justify-content: center;
  z-index: 1000; padding: 20px;
}
.modal-box {
  border-radius: var(--radius-xl);
  width: 100%; max-width: 520px;
  max-height: 90vh;
  display: flex; flex-direction: column;
  overflow: hidden;
  background: rgba(17, 24, 39, 0.75); /* --bg-elevated but transparent */
  backdrop-filter: blur(12px);
  border: 1px solid var(--border);
}
.modal-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 18px 22px 14px;
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
  border-top-left-radius: var(--radius-xl);
  border-top-right-radius: var(--radius-xl);
}
.modal-title { font-size: 17px; font-weight: 700; }
.modal-close {
  width: 30px; height: 30px; border-radius: 50%;
  background: rgba(255,255,255,0.06); border: none;
  color: var(--text-secondary); cursor: pointer; font-size: 14px;
  display: flex; align-items: center; justify-content: center;
  transition: background 0.15s;
}
.modal-close:hover { background: rgba(255,255,255,0.12); }

.modal-body {
  overflow-y: auto; padding: 18px 22px;
  display: flex; flex-direction: column; gap: 16px;
}
.modal-body::-webkit-scrollbar { width: 4px; }
.modal-body::-webkit-scrollbar-thumb { background: var(--border); border-radius: 2px; }

.episodes-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 300px;
  overflow-y: auto;
  padding-right: 8px;
}
.episodes-list::-webkit-scrollbar {
  width: 6px;
}
.episodes-list::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 4px;
}
.episodes-list::-webkit-scrollbar-track {
  background: transparent;
}

.ep-row-btn {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  text-align: left;
  padding: 10px 14px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid transparent;
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s ease;
  width: 100%;
}
.ep-row-btn:hover {
  background: rgba(255, 255, 255, 0.1);
}
.ep-row-btn.active {
  background: rgba(99, 102, 241, 0.15);
  border-color: rgba(99, 102, 241, 0.5);
}
.ep-num {
  font-size: 12px;
  color: var(--text-muted);
  margin-bottom: 4px;
  font-weight: 500;
}
.ep-name {
  font-size: 14px;
  font-weight: 600;
  color: #fff;
  line-height: 1.3;
}

.btn-sm {
  padding: 6px 12px;
  font-size: 12px;
  border-radius: var(--radius-sm);
}

.w-full {
  width: 100%;
}
</style>
