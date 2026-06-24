<template>
  <div class="modal-overlay" @click.self="$emit('close')">
    <div class="modal-box glass animate-fade-in" style="max-width: 600px; width: 100%;">
      <div class="modal-header">
        <h3 class="modal-title">📺 Выберите озвучку и серию</h3>
        <button class="modal-close" @click="$emit('close')">✕</button>
      </div>

      <div class="modal-body" style="padding: 20px; max-height: 75vh; overflow-y: auto;">
        <div v-if="loading" class="flex flex-col items-center justify-center py-12 text-secondary">
          <div class="spinner mb-4"></div>
          <div>Загрузка озвучек и серий...</div>
        </div>

        <div v-else-if="error" class="text-center py-6 text-red-400">
          ⚠️ {{ error }}
        </div>

        <div v-else-if="!shikimoriId && legacyEpisodes.length === 0" class="text-center py-6 text-secondary">
          Эпизоды не найдены
        </div>

        <div v-else-if="shikimoriId && translations.length === 0" class="text-center py-6 text-secondary">
          Не найдено видео-материалов в базе Kodik
        </div>

        <div v-else class="flex flex-col gap-5">
          <!-- БЛОК 1: Выбор озвучки (для Kodik) -->
          <div v-if="shikimoriId" class="translation-section">
            <h4 class="section-title">🎙️ Выберите перевод / озвучку:</h4>
            
            <!-- Озвучка -->
            <div v-if="voiceTranslations.length > 0" class="translation-group mb-4">
              <span class="group-label">Голосовая озвучка:</span>
              <div class="translation-grid mt-2">
                <button
                  v-for="t in voiceTranslations"
                  :key="t.id"
                  class="trans-btn"
                  :class="{ 'active': selectedTranslation && selectedTranslation.id === t.id }"
                  @click="selectTranslation(t)"
                >
                  <span class="trans-icon">🎙️</span>
                  <span class="trans-name">{{ t.translation.title }}</span>
                  <span class="trans-badge">{{ t.last_episode }} сер.</span>
                </button>
              </div>
            </div>

            <!-- Субтитры -->
            <div v-if="subTranslations.length > 0" class="translation-group">
              <span class="group-label">Субтитры:</span>
              <div class="translation-grid mt-2">
                <button
                  v-for="t in subTranslations"
                  :key="t.id"
                  class="trans-btn sub-type"
                  :class="{ 'active': selectedTranslation && selectedTranslation.id === t.id }"
                  @click="selectTranslation(t)"
                >
                  <span class="trans-icon">📝</span>
                  <span class="trans-name">{{ t.translation.title }}</span>
                  <span class="trans-badge">{{ t.last_episode }} сер.</span>
                </button>
              </div>
            </div>
          </div>

          <!-- БЛОК 2: Поиск по номеру серии (если серий много) -->
          <div v-if="episodesList.length > 12" class="search-section">
            <input 
              v-model="searchQuery" 
              type="text" 
              class="form-input search-input" 
              placeholder="Поиск по номеру серии..." 
            />
          </div>

          <!-- БЛОК 3: Выбор эпизода -->
          <div class="episodes-section">
            <h4 class="section-title">🎬 Эпизод:</h4>
            <div class="episodes-grid mt-2">
              <button 
                v-for="ep in filteredEpisodes" 
                :key="ep.number"
                class="ep-grid-btn"
                @click="onEpisodeClick(ep)"
              >
                <div class="ep-num">{{ ep.number }}</div>
                <div class="ep-desc">серия</div>
              </button>
            </div>
          </div>

          <!-- Легаси поток Anilibria (если есть alias и нет shikimoriId) -->
          <div v-if="!shikimoriId && legacyEpisodes.length > 0" class="legacy-section">
            <h4 class="section-title">⚡ Поток Anilibria:</h4>
            <div class="episodes-grid mt-2">
              <button 
                v-for="ep in filteredLegacyEpisodes" 
                :key="ep.id"
                class="ep-grid-btn"
                @click="selectLegacyHls(ep.hls_1080 || ep.hls_720 || ep.hls_480 || ep.hls)"
              >
                <div class="ep-num">{{ ep.ordinal ?? ep.number }}</div>
                <div class="ep-desc">серия</div>
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'

const props = defineProps({
  alias: String,
  shikimoriId: [String, Number]
})

const emit = defineEmits(['close', 'select'])

const loading = ref(true)
const error = ref(null)
const searchQuery = ref('')

// Переменные для Kodik
const translations = ref([])
const selectedTranslation = ref(null)

// Переменные для легаси Anilibria
const legacyEpisodes = ref([])
const externalPlayerUrl = ref('')

// Разделение переводов по типу
const voiceTranslations = computed(() => {
  return translations.value.filter(t => t.translation.type === 'voice')
})

const subTranslations = computed(() => {
  return translations.value.filter(t => t.translation.type === 'subtitles')
})

// Парсинг серий для активной озвучки Kodik
const episodesList = computed(() => {
  if (!selectedTranslation.value) return []
  const list = []
  const seasons = selectedTranslation.value.seasons || {}
  
  const sortedSeasonKeys = Object.keys(seasons).sort((a, b) => parseInt(a, 10) - parseInt(b, 10))
  
  for (const sKey of sortedSeasonKeys) {
    const season = seasons[sKey]
    const episodes = season.episodes || {}
    const sortedEpisodeKeys = Object.keys(episodes).sort((a, b) => parseInt(a, 10) - parseInt(b, 10))
    for (const epKey of sortedEpisodeKeys) {
      list.push({
        number: parseInt(epKey, 10),
        season: parseInt(sKey, 10),
        link: episodes[epKey]
      })
    }
  }
  
  if (list.length === 0) {
    const count = selectedTranslation.value.last_episode || selectedTranslation.value.episodes_count || 1
    for (let i = 1; i <= count; i++) {
      list.push({
        number: i,
        season: 1,
        link: ''
      })
    }
  }
  
  return list
})

const filteredEpisodes = computed(() => {
  if (searchQuery.value) {
    const q = searchQuery.value.trim().toLowerCase()
    return episodesList.value.filter(ep => ep.number.toString().includes(q))
  }
  return episodesList.value
})

const filteredLegacyEpisodes = computed(() => {
  if (searchQuery.value) {
    const q = searchQuery.value.trim().toLowerCase()
    return legacyEpisodes.value.filter(ep => {
      const num = (ep.ordinal ?? ep.number) || ''
      return num.toString().includes(q)
    })
  }
  return legacyEpisodes.value
})

onMounted(async () => {
  const auth = useAuthStore()
  loading.value = true
  
  if (props.shikimoriId) {
    // 1. Поиск озвучек через Kodik
    try {
      const res = await fetch(`/api/v1/reviews/integrations/kodik/search?shikimori_id=${props.shikimoriId}`, {
        headers: {
          'Authorization': `Bearer ${auth.accessToken}`
        }
      })
      if (!res.ok) {
        throw new Error('Не удалось загрузить базу озвучек')
      }
      const data = await res.json()
      translations.value = data.results || []
      
      if (translations.value.length > 0) {
        // По умолчанию выбираем первую озвучку
        selectedTranslation.value = voiceTranslations.value[0] || translations.value[0]
      }
    } catch (e) {
      console.error(e)
      error.value = e.message
    }
  } else if (props.alias) {
    // 2. Легаси поиск через AniLiberty
    try {
      const res = await fetch(`/api/v1/reviews/integrations/aniliberty/episodes/${props.alias}`, {
        headers: {
          'Authorization': `Bearer ${auth.accessToken}`
        }
      })
      const data = await res.json()
      externalPlayerUrl.value = data.external_player || ''
      legacyEpisodes.value = data.episodes || []
      legacyEpisodes.value.sort((a, b) => parseFloat(a.number) - parseFloat(b.number))
    } catch (e) {
      console.error("Failed to fetch AniLiberty episodes", e)
      error.value = "Ошибка при загрузке эпизодов Anilibria"
    }
  } else {
    error.value = "Отсутствует ID аниме для поиска"
  }
  
  loading.value = false
})

function selectTranslation(t) {
  selectedTranslation.value = t
}

function onEpisodeClick(ep) {
  if (!selectedTranslation.value) return
  
  // Строим каноническую ссылку на Kodik для watchparty
  let url = `https://kodik.info/find?shikimori_id=${props.shikimoriId}&episode=${ep.number}&translation_id=${selectedTranslation.value.translation.id}&shikimori=${props.shikimoriId}&only_episode=true&only_translation=true`
  if (props.alias) {
    url += `&alias=${props.alias}`
  }
  
  emit('select', url)
}

function selectLegacyHls(url) {
  if (!url) return
  emit('select', url)
}
</script>

<style scoped>
.modal-overlay {
  position: fixed; inset: 0;
  background: rgba(0,0,0,0.6); backdrop-filter: blur(8px);
  display: flex; align-items: center; justify-content: center;
  z-index: 1000; padding: 20px;
}

.modal-box {
  border-radius: 20px;
  max-height: 85vh;
  display: flex; flex-direction: column;
  overflow: hidden;
  background: rgba(22, 22, 38, 0.7);
  backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.08);
  box-shadow: 0 20px 40px rgba(0,0,0,0.4);
}

.modal-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 20px 24px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  flex-shrink: 0;
}

.modal-title { font-size: 1.15rem; font-weight: 700; color: #fff; }

.modal-close {
  width: 32px; height: 32px; border-radius: 50%;
  background: rgba(255, 255, 255, 0.05); border: none;
  color: rgba(255, 255, 255, 0.6); cursor: pointer; font-size: 14px;
  display: flex; align-items: center; justify-content: center;
  transition: all 0.2s ease;
}
.modal-close:hover {
  background: rgba(255, 255, 255, 0.12);
  color: #fff;
  transform: rotate(90deg);
}

.section-title {
  font-size: 0.9rem;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.9);
  margin-bottom: 8px;
}

.group-label {
  font-size: 0.8rem;
  color: rgba(255, 255, 255, 0.5);
  font-weight: 500;
}

/* Озвучки */
.translation-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(170px, 1fr));
  gap: 10px;
  max-height: 180px;
  overflow-y: auto;
  padding-right: 4px;
}

.translation-grid::-webkit-scrollbar { width: 4px; }
.translation-grid::-webkit-scrollbar-thumb { background: rgba(255, 255, 255, 0.1); border-radius: 2px; }

.trans-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
  text-align: left;
}
.trans-btn:hover {
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.15);
  transform: translateY(-1px);
}
.trans-btn.active {
  background: rgba(99, 102, 241, 0.15);
  border-color: rgba(99, 102, 241, 0.6);
  box-shadow: 0 0 12px rgba(99, 102, 241, 0.2);
}
.trans-name {
  font-size: 0.8rem;
  font-weight: 600;
  color: #fff;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  flex: 1;
}
.trans-badge {
  font-size: 0.7rem;
  color: rgba(255, 255, 255, 0.4);
  background: rgba(0,0,0,0.2);
  padding: 2px 6px;
  border-radius: 4px;
}
.trans-icon { font-size: 0.95rem; }

/* Эпизоды */
.episodes-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(70px, 1fr));
  gap: 8px;
  max-height: 240px;
  overflow-y: auto;
  padding-right: 4px;
}
.episodes-grid::-webkit-scrollbar { width: 4px; }
.episodes-grid::-webkit-scrollbar-thumb { background: rgba(255, 255, 255, 0.1); border-radius: 2px; }

.ep-grid-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 10px 6px;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s ease;
}
.ep-grid-btn:hover {
  background: rgba(99, 102, 241, 0.1);
  border-color: rgba(99, 102, 241, 0.4);
  transform: scale(1.05);
}
.ep-num { font-size: 1rem; font-weight: 700; color: #fff; }
.ep-desc { font-size: 0.65rem; color: rgba(255, 255, 255, 0.4); text-transform: uppercase; letter-spacing: 0.5px; }

.form-input {
  background: rgba(0, 0, 0, 0.25);
  border: 1px solid rgba(255, 255, 255, 0.08);
  color: #fff;
  padding: 10px 14px;
  border-radius: 10px;
  font-size: 0.85rem;
  width: 100%;
  outline: none;
  transition: all 0.2s ease;
}
.form-input:focus {
  border-color: rgba(99, 102, 241, 0.5);
  background: rgba(0, 0, 0, 0.4);
}

.spinner {
  width: 36px; height: 36px;
  border: 3px solid rgba(255, 255, 255, 0.1);
  border-top-color: rgba(99, 102, 241, 1);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

@keyframes fadeIn {
  from { opacity: 0; transform: scale(0.97); }
  to { opacity: 1; transform: scale(1); }
}
.animate-fade-in {
  animation: fadeIn 0.25s cubic-bezier(0.16, 1, 0.3, 1) forwards;
}
</style>
