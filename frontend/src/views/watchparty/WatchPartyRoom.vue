<template>
  <div class="room-layout">
    <header class="room-header glass">
      <div class="header-left">
        <router-link to="/reviews" class="btn-back">🚪 Выйти к рецензиям</router-link>
        <h2>📺 Watch Party</h2>
        <span class="room-id">Комната: {{ roomId }}</span>
      </div>
      <div class="header-right">
        <button class="btn-copy" @click="copyLink">🔗 Копировать ссылку</button>
      </div>
    </header>

    <div class="room-content">
      <div class="main-area">
        <div class="player-wrapper">
          <!-- Warnings -->
          <div v-if="roomState.error" class="error-banner">
            {{ roomState.error }}
          </div>
          
          <div v-if="detectVideoType(roomState.videoUrl) === 'unknown' && roomState.videoUrl" class="warning-banner">
            blob:-ссылки или неизвестные форматы не поддерживаются для синхронизации.
          </div>

          <!-- Player Components -->
          <YouTubePlayer
            v-if="roomState.videoType === 'youtube' && !roomState.error"
            ref="playerRef"
            :url="roomState.videoUrl"
            @local-play="onLocalPlay"
            @local-pause="onLocalPause"
            @local-seek="onLocalSeek"
          />
          <DirectVideoPlayer
            v-else-if="roomState.videoType === 'direct' && !roomState.error"
            ref="playerRef"
            :url="roomState.videoUrl"
            @local-play="onLocalPlay"
            @local-pause="onLocalPause"
            @local-seek="onLocalSeek"
          />
          <RutubePlayer
            v-else-if="roomState.videoType === 'rutube' && !roomState.error"
            ref="playerRef"
            :url="roomState.videoUrl"
            @local-play="onLocalPlay"
            @local-pause="onLocalPause"
            @local-seek="onLocalSeek"
          />
          <KodikVideoPlayer
            v-else-if="roomState.videoType === 'kodik' && !roomState.error"
            ref="playerRef"
            :url="roomState.videoUrl"
            @local-play="onLocalPlay"
            @local-pause="onLocalPause"
            @local-seek="onLocalSeek"
            @local-episode-change="onLocalEpisodeChange"
          />
          <AllohaVideoPlayer
            v-else-if="roomState.videoType === 'alloha' && !roomState.error"
            ref="playerRef"
            :url="roomState.videoUrl"
            @local-play="onLocalPlay"
            @local-pause="onLocalPause"
            @local-seek="onLocalSeek"
          />
          <div v-else-if="!roomState.error" class="empty-player">
            Видео не выбрано
          </div>
        </div>

        <!-- Панель переключения источников (доступна владельцу для аниме) -->
        <div class="player-selector glass" v-if="roomState.isOwner && !roomState.error && hasAnimeMetadata">
          <span class="selector-label">📺 Плеер:</span>
          <div class="selector-buttons">
            <button 
              class="selector-btn" 
              :class="{ active: roomState.videoType === 'kodik' }"
              @click="switchPlayer('kodik')"
            >
              🎬 Kodik (Все озвучки)
            </button>
            <button 
              class="selector-btn" 
              :class="{ active: roomState.videoType === 'alloha' }"
              @click="switchPlayer('alloha')"
            >
              🍿 Alloha (Зеркало)
            </button>
            <button 
              class="selector-btn" 
              :class="{ active: roomState.videoType === 'direct' }"
              @click="switchPlayer('direct')"
              :disabled="loadingDirect"
            >
              ⚡ Anilibria {{ loadingDirect ? '(Загрузка...)' : '' }}
            </button>
          </div>
        </div>

        <!-- Панель управления сериями и озвучками -->
        <div class="anime-controls-panel glass animate-fade-in" v-if="!roomState.error && hasAnimeMetadata && (roomState.videoType === 'kodik' || roomState.videoType === 'alloha')">
          <div class="panel-header">
            <h3 class="panel-title">⭐ Управление озвучкой и сериями</h3>
            <span v-if="!roomState.isOwner" class="badge-host-controlled">🔒 Управляет создатель комнаты</span>
          </div>

          <div class="panel-body">
            <!-- Состояние загрузки -->
            <div v-if="loadingTranslations" class="loading-state">
              <div class="spinner-small"></div>
              <span>Загрузка доступных озвучек...</span>
            </div>
            <div v-else-if="translationsError" class="error-state">
              ⚠️ {{ translationsError }}
            </div>
            <template v-else>
              <!-- Выбор перевода -->
              <div class="control-row">
                <span class="control-label">🎙️ Озвучка / Перевод:</span>
                <div class="translations-scroll">
                  <button
                    v-for="t in roomTranslations"
                    :key="t.id"
                    class="trans-pill-btn"
                    :class="{ 
                      active: currentTranslationId == t.translation.id,
                      disabled: !roomState.isOwner
                    }"
                    :disabled="!roomState.isOwner"
                    @click="onTranslationSelect(t)"
                  >
                    {{ t.translation.title }}
                  </button>
                </div>
              </div>

              <!-- Выбор серии -->
              <div class="control-row mt-3">
                <span class="control-label">🎬 Выбор серии (всего {{ episodesCountForActiveTranslation }}):</span>
                <div class="episodes-scroll-grid">
                  <button
                    v-for="ep in episodesListForActiveTranslation"
                    :key="ep"
                    class="ep-pill-btn"
                    :class="{ 
                      active: currentEpisode == ep,
                      disabled: !roomState.isOwner
                    }"
                    :disabled="!roomState.isOwner"
                    @click="onEpisodeSelect(ep)"
                  >
                    {{ ep }}
                  </button>
                </div>
              </div>
            </template>
          </div>
        </div>

        <div class="url-control glass" v-if="roomState.isOwner && !roomState.error">
          <input v-model="editUrl" type="text" placeholder="Новая ссылка на видео..." class="url-input" />
          <button @click="updateUrl" class="btn-change">Сменить видео</button>
        </div>
        <div class="url-display glass" v-else-if="!roomState.error">
          Текущее видео: {{ roomState.videoUrl || 'Не выбрано' }}
        </div>
      </div>

      <aside class="side-panel glass">
        <div class="participants-section">
          <h3>Участники ({{ roomState.participants.length }})</h3>
          <ul class="participant-list">
            <li v-for="p in roomState.participants" :key="p.user_id" class="participant-item">
              <span class="status-dot"></span>
              <span class="name">{{ p.username }} <span v-if="p.is_owner">(Host)</span></span>
              <button 
                v-if="roomState.isOwner && !p.is_owner" 
                class="btn-kick" 
                @click="kickUser(p.user_id)"
                title="Выгнать"
              >
                ✕
              </button>
            </li>
          </ul>
        </div>

        <!-- Knock Requests (Owner only) -->
        <div v-if="roomState.isOwner && roomState.knockRequests.length > 0" class="knock-section">
          <h3>Запросы на вход</h3>
          <div v-for="req in roomState.knockRequests" :key="req.user_id" class="knock-card">
            <div>🚪 <strong>{{ req.username }}</strong> хочет войти</div>
            <div class="knock-actions">
              <button class="btn-admit" @click="admitUser(req.user_id)">Впустить</button>
              <button class="btn-reject" @click="rejectUser(req.user_id)">Отказ</button>
            </div>
          </div>
        </div>
      </aside>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, computed, watch } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useWatchParty } from '@/composables/useWatchParty'
import YouTubePlayer from '@/components/watchparty/YouTubePlayer.vue'
import DirectVideoPlayer from '@/components/watchparty/DirectVideoPlayer.vue'
import RutubePlayer from '@/components/watchparty/RutubePlayer.vue'
import KodikVideoPlayer from '@/components/watchparty/KodikVideoPlayer.vue'
import AllohaVideoPlayer from '@/components/watchparty/AllohaVideoPlayer.vue'

const route = useRoute()
const roomId = route.params.roomId
const auth = useAuthStore()

const { 
  roomState, 
  playerRef, 
  connect, 
  onLocalPlay, 
  onLocalPause, 
  onLocalSeek, 
  changeVideo, 
  admitUser, 
  rejectUser, 
  kickUser 
} = useWatchParty()

const editUrl = ref('')
const loadingDirect = ref(false)

// Метаданные аниме, извлекаемые из sessionStorage или URL
const currentShikimoriId = ref(sessionStorage.getItem(`wp_shikimori_${roomId}`) || '')
const currentAlias = ref(sessionStorage.getItem(`wp_alias_${roomId}`) || '')

const hasAnimeMetadata = computed(() => {
  return currentShikimoriId.value !== ''
})

// Кастомное управление озвучками и сериями для Kodik
const roomTranslations = ref([])
const loadingTranslations = ref(false)
const translationsError = ref(null)

async function fetchRoomTranslations() {
  if (!currentShikimoriId.value) return
  loadingTranslations.value = true
  translationsError.value = null
  try {
    const token = auth.accessToken
    const res = await fetch(`/api/v1/reviews/integrations/kodik/search?shikimori_id=${currentShikimoriId.value}`, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    })
    if (!res.ok) {
      throw new Error('Не удалось загрузить список озвучек для этой комнаты')
    }
    const data = await res.json()
    roomTranslations.value = data.results || []
  } catch (e) {
    console.error(e)
    translationsError.value = e.message
  } finally {
    loadingTranslations.value = false
  }
}

watch(currentShikimoriId, (newId) => {
  if (newId) {
    fetchRoomTranslations()
  }
})

const currentTranslationId = computed(() => {
  if (!roomState.videoUrl) return ''
  try {
    let url = roomState.videoUrl
    if (url.startsWith('//')) {
      url = window.location.protocol + url
    }
    const urlObj = new URL(url)
    return urlObj.searchParams.get('translation_id') || ''
  } catch (e) {
    return ''
  }
})

const activeTranslationData = computed(() => {
  if (roomTranslations.value.length === 0) return null
  if (currentTranslationId.value) {
    const found = roomTranslations.value.find(t => t.translation.id == currentTranslationId.value)
    if (found) return found
  }
  return roomTranslations.value[0]
})

const episodesCountForActiveTranslation = computed(() => {
  if (!activeTranslationData.value) return 0
  return activeTranslationData.value.last_episode || activeTranslationData.value.episodes_count || 1
})

const episodesListForActiveTranslation = computed(() => {
  if (!activeTranslationData.value) return []
  
  const list = []
  const seasons = activeTranslationData.value.seasons || {}
  
  const sortedSeasonKeys = Object.keys(seasons).sort((a, b) => parseInt(a, 10) - parseInt(b, 10))
  
  for (const sKey of sortedSeasonKeys) {
    const season = seasons[sKey]
    const episodes = season.episodes || {}
    const sortedEpisodeKeys = Object.keys(episodes).sort((a, b) => parseInt(a, 10) - parseInt(b, 10))
    for (const epKey of sortedEpisodeKeys) {
      list.push(parseInt(epKey, 10))
    }
  }
  
  if (list.length === 0) {
    const count = episodesCountForActiveTranslation.value
    for (let i = 1; i <= count; i++) {
      list.push(i)
    }
  }
  
  return list
})

function onTranslationSelect(t) {
  if (!roomState.isOwner) return
  
  let targetEp = currentEpisode.value
  const seasons = t.seasons || {}
  let baseLink = ''
  
  // Ищем ссылку на текущий эпизод в новом переводе
  for (const sKey in seasons) {
    if (seasons[sKey].episodes && seasons[sKey].episodes[targetEp]) {
      baseLink = seasons[sKey].episodes[targetEp]
      break
    }
  }
  
  // Если эпизод не нашёлся — берём первый доступный эпизод
  if (!baseLink) {
    for (const sKey in seasons) {
      if (seasons[sKey].episodes) {
        const sortedKeys = Object.keys(seasons[sKey].episodes).sort((a, b) => parseInt(a, 10) - parseInt(b, 10))
        if (sortedKeys.length > 0) {
          targetEp = parseInt(sortedKeys[0], 10)
          baseLink = seasons[sKey].episodes[sortedKeys[0]]
          break
        }
      }
    }
  }

  // Если всё ещё нет ссылки — используем serial-уровень link
  if (!baseLink) {
    baseLink = t.link || ''
  }

  const shId = currentShikimoriId.value
  const alias = currentAlias.value
  
  if (roomState.videoType === 'alloha') {
    const activeMirror = localStorage.getItem('alloha_mirror') || 'api.alloha.live'
    const url = `https://${activeMirror}/?token=df2ef76e33055d72f107f90c885068&shikimori=${shId}&episode=${targetEp}&translation=${encodeURIComponent(t.translation.title)}&translation_id=${t.translation.id}`
    changeVideo(url, 'alloha')
  } else {
    // Kodik: используем прямую ссылку из API без лишних параметров
    if (baseLink.startsWith('//')) {
      baseLink = 'https:' + baseLink
    }
    // Принудительно переписываем домен на рабочий миррор
    try {
      const urlObj = new URL(baseLink)
      urlObj.hostname = 'kodikplayer.com'
      // Добавляем метаданные для гостей (плеер Kodik игнорирует неизвестные параметры)
      if (shId) urlObj.searchParams.set('shikimori', String(shId))
      if (alias) urlObj.searchParams.set('alias', alias)
      changeVideo(urlObj.toString(), 'kodik')
    } catch (e) {
      changeVideo(baseLink, 'kodik')
    }
  }
}

function onEpisodeSelect(episodeNum) {
  if (!roomState.isOwner) return
  
  const shId = currentShikimoriId.value
  const alias = currentAlias.value
  const translationId = currentTranslationId.value || (activeTranslationData.value ? activeTranslationData.value.translation.id : '')

  if (roomState.videoType === 'alloha') {
    const activeTrans = activeTranslationData.value
    const transTitle = activeTrans ? activeTrans.translation.title : ''
    const activeMirror = localStorage.getItem('alloha_mirror') || 'api.alloha.live'
    
    let url = `https://${activeMirror}/?token=df2ef76e33055d72f107f90c885068&shikimori=${shId}&episode=${episodeNum}`
    if (transTitle) url += `&translation=${encodeURIComponent(transTitle)}`
    if (translationId) url += `&translation_id=${translationId}`
    changeVideo(url, 'alloha')
  } else {
    // Kodik: ищем прямую ссылку на эпизод в данных API
    let baseLink = ''
    if (activeTranslationData.value?.seasons) {
      for (const sKey in activeTranslationData.value.seasons) {
        const epLink = activeTranslationData.value.seasons[sKey]?.episodes?.[episodeNum]
        if (epLink) {
          baseLink = epLink
          break
        }
      }
    }
    
    // Если конкретная ссылка не нашлась — используем serial-уровень ссылку с параметром episode
    if (!baseLink) {
      baseLink = activeTranslationData.value?.link
        ? activeTranslationData.value.link
        : `//kodikplayer.com/serial/find?shikimori_id=${shId}&episode=${episodeNum}&translation_id=${translationId}`
    }

    if (baseLink.startsWith('//')) {
      baseLink = 'https:' + baseLink
    }

    try {
      const urlObj = new URL(baseLink)
      urlObj.hostname = 'kodikplayer.com'
      // Добавляем метаданные для гостей
      if (shId) urlObj.searchParams.set('shikimori', String(shId))
      if (alias) urlObj.searchParams.set('alias', alias)
      changeVideo(urlObj.toString(), 'kodik')
    } catch (e) {
      changeVideo(baseLink, 'kodik')
    }
  }
}

// Извлекаем номер текущей серии из URL
const currentEpisode = computed(() => {
  if (!roomState.videoUrl) return 1
  try {
    let url = roomState.videoUrl
    if (url.startsWith('//')) {
      url = window.location.protocol + url
    }
    const urlObj = new URL(url)
    return parseInt(urlObj.searchParams.get('episode') || '1', 10)
  } catch (e) {
    return 1
  }
})

// Парсинг метаданных из URL (позволяет гостям получать ID аниме и алиас из ссылки хоста)
function extractMetadataFromUrl(url) {
  if (!url) return
  try {
    let cleanUrl = url
    if (cleanUrl.startsWith('//')) {
      cleanUrl = window.location.protocol + cleanUrl
    }
    const urlObj = new URL(cleanUrl)
    const shId = urlObj.searchParams.get('shikimori') || urlObj.searchParams.get('shikimori_id')
    const al = urlObj.searchParams.get('alias')
    
    if (shId && !currentShikimoriId.value) {
      currentShikimoriId.value = shId
    }
    if (al && !currentAlias.value) {
      currentAlias.value = al
    }
  } catch (e) {}
}

watch(() => roomState.videoUrl, (newUrl) => {
  extractMetadataFromUrl(newUrl)
})

function detectVideoType(url) {
  if (!url) return ''
  if (/youtube\.com|youtu\.be/.test(url)) return 'youtube'
  if (/rutube\.ru/.test(url)) return 'rutube'
  if (/kodik|aniqit/i.test(url)) return 'kodik'
  if (/alloha/i.test(url)) return 'alloha'
  if (/\.(mp4|webm|ogg|m3u8)(\?|$)/i.test(url)) return 'direct'
  return 'unknown'
}

function updateUrl() {
  if (!editUrl.value) return
  const vType = detectVideoType(editUrl.value)
  changeVideo(editUrl.value, vType)
  editUrl.value = ''
}

function onLocalEpisodeChange(episodeNum) {
  if (!roomState.isOwner) return
  try {
    let url = roomState.videoUrl
    if (!url) return

    if (url.startsWith('//')) {
      url = window.location.protocol + url
    }

    const urlObj = new URL(url)
    urlObj.searchParams.set('episode', episodeNum)
    const newUrl = urlObj.toString()

    if (newUrl !== roomState.videoUrl) {
      changeVideo(newUrl, detectVideoType(newUrl))
    }
  } catch (e) {
    console.error('Failed to change episode via owner action', e)
  }
}

// Метод переключения плеера хостом
async function switchPlayer(type) {
  if (!roomState.isOwner) return
  
  const shId = currentShikimoriId.value
  const alias = currentAlias.value
  const ep = currentEpisode.value
  
  if (type === 'kodik') {
    const translationId = currentTranslationId.value || (activeTranslationData.value ? activeTranslationData.value.translation.id : '')
    let baseLink = ''
    
    // Ищем прямую ссылку на эпизод в данных API
    if (activeTranslationData.value?.seasons) {
      for (const sKey in activeTranslationData.value.seasons) {
        const epLink = activeTranslationData.value.seasons[sKey]?.episodes?.[ep]
        if (epLink) {
          baseLink = epLink
          break
        }
      }
    }
    
    // Если нет — берём serial-уровень ссылку с параметром episode
    if (!baseLink) {
      baseLink = activeTranslationData.value?.link || `//kodikplayer.com/serial/find?shikimori_id=${shId}&episode=${ep}&translation_id=${translationId}`
    }
    
    if (baseLink.startsWith('//')) {
      baseLink = 'https:' + baseLink
    }
    
    try {
      const urlObj = new URL(baseLink)
      urlObj.hostname = 'kodikplayer.com'
      if (shId) urlObj.searchParams.set('shikimori', String(shId))
      if (alias) urlObj.searchParams.set('alias', alias)
      changeVideo(urlObj.toString(), 'kodik')
    } catch (e) {
      changeVideo(baseLink, 'kodik')
    }
  } else if (type === 'alloha') {
    const translationId = currentTranslationId.value || (activeTranslationData.value ? activeTranslationData.value.translation.id : '')
    const activeTrans = activeTranslationData.value
    const transTitle = activeTrans ? activeTrans.translation.title : ''
    const activeMirror = localStorage.getItem('alloha_mirror') || 'api.alloha.live'
    
    let url = `https://${activeMirror}/?token=df2ef76e33055d72f107f90c885068&shikimori=${shId}&episode=${ep}`
    if (transTitle) url += `&translation=${encodeURIComponent(transTitle)}`
    if (translationId) url += `&translation_id=${translationId}`
    changeVideo(url, 'alloha')
  } else if (type === 'direct') {
    loadingDirect.value = true
    try {
      // Запрашиваем список серий из Anilibria для данного алиаса
      const token = auth.accessToken
      const res = await fetch(`/api/v1/reviews/integrations/aniliberty/episodes/${alias}`, {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })
      const data = await res.json()
      const episodes = data.episodes || []
      
      // Ищем эпизод, соответствующий текущему номеру серии
      const foundEp = episodes.find(e => (e.ordinal ?? e.number) == ep)
      if (foundEp) {
        const hlsUrl = foundEp.hls_1080 || foundEp.hls_720 || foundEp.hls_480 || foundEp.hls
        if (hlsUrl) {
          // Сохраняем метаданные в параметрах прямого URL
          const urlObj = new URL(hlsUrl)
          urlObj.searchParams.set('shikimori', shId)
          urlObj.searchParams.set('alias', alias)
          urlObj.searchParams.set('episode', ep)
          
          changeVideo(urlObj.toString(), 'direct')
        } else {
          alert('Прямой HLS-поток для этой серии не найден.')
        }
      } else {
        alert(`Серия ${ep} не найдена в базе Anilibria.`)
      }
    } catch (e) {
      console.error('Failed to switch to direct player', e)
      alert('Не удалось переключиться на прямой плеер.')
    } finally {
      loadingDirect.value = false
    }
  }
}

function copyLink() {
  const url = window.location.origin + `/watch/room/${roomId}`
  navigator.clipboard.writeText(url)
}

onMounted(() => {
  const token = auth.accessToken
  connect(roomId, token)

  if (currentShikimoriId.value) {
    fetchRoomTranslations()
  }

  // Check if we just created it with an initial URL
  const initialUrl = sessionStorage.getItem(`wp_url_${roomId}`)
  if (initialUrl) {
    sessionStorage.removeItem(`wp_url_${roomId}`)
    extractMetadataFromUrl(initialUrl)
    // We wait a bit for connection to establish and check if we are owner
    setTimeout(() => {
      if (roomState.isOwner) {
        changeVideo(initialUrl, detectVideoType(initialUrl))
      }
    }, 1000)
  }
})
</script>

<style scoped>
.room-layout {
  display: flex;
  flex-direction: column;
  height: calc(100vh - 60px);
  background: var(--bg-color, #121212);
  color: #fff;
}

.room-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 24px;
  background: rgba(255, 255, 255, 0.02);
  backdrop-filter: blur(20px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  z-index: 10;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 16px;
}

.btn-back {
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.08);
  color: #fff;
  text-decoration: none;
  padding: 8px 16px;
  border-radius: var(--radius-md, 12px);
  font-size: 0.9rem;
  font-weight: 500;
  transition: all 0.2s ease;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.btn-back:hover {
  background: rgba(255, 255, 255, 0.12);
  border-color: rgba(255, 255, 255, 0.15);
  transform: translateY(-1px);
}

.header-left h2 {
  margin: 0;
  font-size: 1.2rem;
}

.room-id {
  background: rgba(0, 0, 0, 0.3);
  padding: 4px 8px;
  border-radius: 4px;
  font-family: monospace;
  color: #aaa;
}

.btn-copy {
  background: rgba(255, 255, 255, 0.06);
  border: 1px solid rgba(255, 255, 255, 0.08);
  padding: 8px 16px;
  border-radius: var(--radius-md, 12px);
  color: #fff;
  cursor: pointer;
  font-size: 0.9rem;
  font-weight: 500;
  transition: all 0.2s ease;
}

.btn-copy:hover {
  background: rgba(255, 255, 255, 0.12);
  border-color: rgba(255, 255, 255, 0.15);
  transform: translateY(-1px);
}

.room-content {
  display: flex;
  flex: 1;
  overflow: hidden;
}

.main-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  padding: 24px;
  gap: 16px;
  overflow-y: auto;
}

.player-wrapper {
  flex: 1;
  background: #000;
  border-radius: 12px;
  position: relative;
  min-height: 400px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.5);
  display: flex;
  align-items: center;
  justify-content: center;
}

.empty-player {
  color: #666;
  font-size: 1.2rem;
}

.error-banner {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  background: rgba(239, 68, 68, 0.9);
  color: white;
  padding: 16px 24px;
  border-radius: 8px;
  text-align: center;
  z-index: 10;
}

.warning-banner {
  position: absolute;
  top: 10px;
  left: 50%;
  transform: translateX(-50%);
  background: rgba(245, 158, 11, 0.9);
  color: white;
  padding: 8px 16px;
  border-radius: 8px;
  z-index: 10;
  font-size: 0.9rem;
}

/* Панель выбора плеера */
.player-selector {
  display: flex;
  align-items: center;
  gap: 16px;
  background: rgba(255, 255, 255, 0.02);
  backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.08);
  padding: 12px 20px;
  border-radius: 12px;
}

.selector-label {
  font-size: 0.9rem;
  font-weight: 600;
  color: var(--text-secondary, #94a3b8);
  white-space: nowrap;
}

.selector-buttons {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.selector-btn {
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.08);
  color: #fff;
  padding: 6px 14px;
  border-radius: 8px;
  cursor: pointer;
  font-size: 0.85rem;
  font-weight: 500;
  transition: all 0.2s ease;
}
.selector-btn:hover:not(:disabled) {
  background: rgba(255, 255, 255, 0.1);
  border-color: rgba(255, 255, 255, 0.15);
}
.selector-btn.active {
  background: linear-gradient(135deg, var(--primary), var(--violet));
  border: 1px solid rgba(255, 255, 255, 0.15);
  box-shadow: 0 4px 15px rgba(99, 102, 241, 0.25);
}
.selector-btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.url-control {
  display: flex;
  gap: 8px;
  background: rgba(255, 255, 255, 0.02);
  backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.08);
  padding: 12px;
  border-radius: 12px;
}

.url-input {
  flex: 1;
  background: rgba(0, 0, 0, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.08);
  padding: 8px 12px;
  border-radius: 8px;
  color: #fff;
}
.url-input:focus {
  outline: none;
  border-color: rgba(99, 102, 241, 0.5);
}

.btn-change {
  background: linear-gradient(135deg, var(--primary), var(--violet));
  border: none;
  padding: 8px 16px;
  border-radius: 8px;
  color: #fff;
  cursor: pointer;
  font-weight: 500;
  transition: all 0.2s ease;
}
.btn-change:hover {
  filter: brightness(1.1);
  transform: translateY(-1px);
}

.url-display {
  padding: 12px;
  color: #aaa;
  background: rgba(255, 255, 255, 0.02);
  backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 12px;
  word-break: break-all;
}

.side-panel {
  width: 300px;
  background: rgba(255, 255, 255, 0.01);
  backdrop-filter: blur(20px);
  border-left: 1px solid rgba(255, 255, 255, 0.08);
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}

.participants-section, .knock-section {
  padding: 20px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

h3 {
  margin-top: 0;
  margin-bottom: 16px;
  font-size: 1rem;
  color: #ccc;
}

.participant-list {
  list-style: none;
  padding: 0;
  margin: 0;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.participant-item {
  display: flex;
  align-items: center;
  gap: 10px;
}

.status-dot {
  width: 8px;
  height: 8px;
  background: #10b981;
  border-radius: 50%;
}

.name {
  flex: 1;
  font-size: 0.95rem;
}

.name span {
  color: #f59e0b;
  font-size: 0.8rem;
}

.btn-kick {
  background: none;
  border: none;
  color: #ef4444;
  cursor: pointer;
  opacity: 0.5;
  transition: opacity 0.2s;
}

.btn-kick:hover {
  opacity: 1;
}

.knock-card {
  background: rgba(245, 158, 11, 0.1);
  border: 1px solid rgba(245, 158, 11, 0.3);
  padding: 12px;
  border-radius: 8px;
  margin-bottom: 12px;
  font-size: 0.9rem;
}

.knock-actions {
  display: flex;
  gap: 8px;
  margin-top: 10px;
}

.btn-admit {
  background: #10b981;
  border: none;
  padding: 6px 12px;
  border-radius: 4px;
  color: white;
  cursor: pointer;
  flex: 1;
}

.btn-reject {
  background: #ef4444;
  border: none;
  padding: 6px 12px;
  border-radius: 4px;
  color: white;
  cursor: pointer;
  flex: 1;
}

/* Glass helper override */
.glass {
  background: rgba(255, 255, 255, 0.02) !important;
  backdrop-filter: blur(24px) !important;
}

/* Стили кастомной панели управления аниме */
.anime-controls-panel {
  padding: 16px 20px;
  border-radius: 12px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  margin-top: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.panel-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  padding-bottom: 8px;
}

.panel-title {
  font-size: 0.95rem;
  font-weight: 700;
  color: #fff;
  margin: 0;
}

.badge-host-controlled {
  font-size: 0.75rem;
  color: rgba(245, 158, 11, 0.95);
  background: rgba(245, 158, 11, 0.1);
  border: 1px solid rgba(245, 158, 11, 0.2);
  padding: 4px 8px;
  border-radius: 6px;
}

.loading-state, .error-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  padding: 14px;
  font-size: 0.85rem;
  color: rgba(255, 255, 255, 0.6);
}

.spinner-small {
  width: 16px;
  height: 16px;
  border: 2px solid rgba(255, 255, 255, 0.1);
  border-top-color: var(--primary, #6366f1);
  border-radius: 50%;
  animation: spin 1s linear infinite;
}

.control-row {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.control-label {
  font-size: 0.8rem;
  color: rgba(255, 255, 255, 0.5);
  font-weight: 600;
}

.translations-scroll {
  display: flex;
  gap: 8px;
  overflow-x: auto;
  padding-bottom: 6px;
}
.translations-scroll::-webkit-scrollbar {
  height: 4px;
}
.translations-scroll::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 2px;
}

.trans-pill-btn {
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.05);
  color: rgba(255, 255, 255, 0.8);
  padding: 6px 12px;
  border-radius: 8px;
  cursor: pointer;
  font-size: 0.8rem;
  font-weight: 600;
  white-space: nowrap;
  transition: all 0.2s ease;
}
.trans-pill-btn:not(.disabled):hover {
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.15);
  color: #fff;
}
.trans-pill-btn.active {
  background: rgba(99, 102, 241, 0.15);
  border-color: rgba(99, 102, 241, 0.6);
  color: #fff;
  box-shadow: 0 0 10px rgba(99, 102, 241, 0.15);
}
.trans-pill-btn.disabled {
  cursor: default;
  opacity: 0.85;
}

.episodes-scroll-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(40px, 1fr));
  gap: 6px;
  max-height: 120px;
  overflow-y: auto;
  padding-right: 4px;
}
.episodes-scroll-grid::-webkit-scrollbar {
  width: 4px;
}
.episodes-scroll-grid::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.1);
  border-radius: 2px;
}

.ep-pill-btn {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.05);
  color: #fff;
  padding: 6px 4px;
  border-radius: 6px;
  cursor: pointer;
  font-size: 0.8rem;
  font-weight: 700;
  text-align: center;
  transition: all 0.15s ease;
}
.ep-pill-btn:not(.disabled):hover {
  background: rgba(99, 102, 241, 0.12);
  border-color: rgba(99, 102, 241, 0.4);
  transform: translateY(-1px);
}
.ep-pill-btn.active {
  background: rgba(99, 102, 241, 0.2);
  border-color: rgba(99, 102, 241, 0.7);
  color: #fff;
}
.ep-pill-btn.disabled {
  cursor: default;
  opacity: 0.85;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

@keyframes fadeIn {
  from { opacity: 0; transform: translateY(4px); }
  to { opacity: 1; transform: translateY(0); }
}
.animate-fade-in {
  animation: fadeIn 0.2s ease-out forwards;
}
</style>
