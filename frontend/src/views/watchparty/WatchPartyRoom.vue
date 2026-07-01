<template>
  <div class="room-layout">
    <header class="room-header glass">
      <div class="header-left">
        <router-link to="/reviews" class="btn-back">🚪 Выйти к рецензиям</router-link>
        <h2>📺 Watch Party</h2>
        <span class="room-id">Комната: {{ roomId }}</span>
      </div>
      
      <!-- Центральная кнопка для вызова панели выбора -->
      <div class="header-center" style="display:flex; justify-content:center; flex:1;">
        <button class="btn-text" @click="toggleMediaSelector" style="font-size:14px; font-weight: 500; color: rgba(255,255,255,0.7); display:flex; align-items:center; gap:6px; background:transparent; border:none; cursor:pointer;">
          Выбрать другое
          <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" :style="{ transform: showMediaSelector ? 'rotate(180deg)' : 'rotate(0deg)', transition: 'transform 0.2s' }"><polyline points="6 9 12 15 18 9"></polyline></svg>
        </button>
      </div>

      <div class="header-right" style="display:flex; gap:12px; align-items:center;">
        <button class="btn-ghost" @click="isTheaterMode = !isTheaterMode" style="padding:8px 12px; font-size:13px; display:flex; align-items:center; gap:6px;">
          <svg v-if="!isTheaterMode" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="3" width="20" height="14" rx="2" ry="2"></rect><line x1="8" y1="21" x2="16" y2="21"></line><line x1="12" y1="17" x2="12" y2="21"></line></svg>
          <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="4" y="4" width="16" height="16" rx="2" ry="2"></rect><rect x="9" y="9" width="6" height="6"></rect><line x1="9" y1="1" x2="9" y2="4"></line><line x1="15" y1="1" x2="15" y2="4"></line><line x1="9" y1="20" x2="9" y2="23"></line><line x1="15" y1="20" x2="15" y2="23"></line><line x1="20" y1="9" x2="23" y2="9"></line><line x1="20" y1="14" x2="23" y2="14"></line><line x1="1" y1="9" x2="4" y2="9"></line><line x1="1" y1="14" x2="4" y2="14"></line></svg>
          {{ isTheaterMode ? 'Свернуть плеер' : 'Широкий экран' }}
        </button>
        <button class="btn-copy" @click="copyLink">🔗 Копировать ссылку</button>
      </div>
    </header>

    <!-- Выпадающая панель выбора медиа -->
    <transition name="slide-down">
      <div v-if="showMediaSelector" class="media-selector-drawer glass">
        <div class="drawer-header flex justify-between items-center mb-4">
          <div class="tabs flex gap-2">
            <button class="drawer-tab" :class="{active: selectorTab === 'personal'}" @click="selectorTab = 'personal'">Личные</button>
            <button class="drawer-tab" :class="{active: selectorTab === 'group'}" @click="selectorTab = 'group'">Групповые</button>
          </div>
          <button class="btn-ghost" @click="showMediaSelector = false" style="padding:4px 8px;">✕ Закрыть</button>
        </div>
        
        <div class="drawer-content">
          <!-- Личные рецензии -->
          <div v-if="selectorTab === 'personal'" class="drawer-grid">
            <div v-if="reviewsStore.loading" class="spinner-small" style="margin:20px auto;"></div>
            <div v-else-if="reviewsStore.reviews.length === 0" class="text-center text-muted" style="grid-column: 1/-1;">Список пуст</div>
            <div v-else v-for="rev in reviewsStore.reviews" :key="rev.id" class="media-item-card" @click="selectMediaToWatch(rev)">
              <img :src="rev.poster_url || '/placeholder.png'" class="media-item-poster" />
              <div class="media-item-info">
                <div class="media-item-title">{{ rev.title }}</div>
                <div class="media-item-type">{{ rev.content_type }}</div>
              </div>
            </div>
          </div>
          
          <!-- Групповые списки -->
          <div v-if="selectorTab === 'group'" class="drawer-group-view">
            <div v-if="groupsStore.loading" class="spinner-small" style="margin:20px auto;"></div>
            <div v-else-if="groupsStore.groups.length === 0" class="text-center text-muted">Нет групп</div>
            <template v-else>
              <!-- Слайдер групп -->
              <div class="groups-slider mb-4">
                <button 
                  v-for="group in groupsStore.groups" 
                  :key="group.id" 
                  class="group-pill"
                  :class="{ active: selectedGroupId === group.id }"
                  @click="toggleGroupSelection(group.id)"
                >
                  👥 {{ group.name }}
                </button>
              </div>
              <!-- Список тайтлов в выбранной группе -->
              <div v-if="selectedGroupId" class="drawer-grid">
                <div v-if="!selectedGroupItems || selectedGroupItems.length === 0" class="text-center text-muted" style="grid-column: 1/-1;">В этой группе нет тайтлов</div>
                <div v-else v-for="item in selectedGroupItems" :key="item.id" class="media-item-card" @click="selectMediaToWatch(item)">
                  <img :src="item.poster_url || '/placeholder.png'" class="media-item-poster" />
                  <div class="media-item-info">
                    <div class="media-item-title">{{ item.title }}</div>
                    <div class="media-item-type">{{ item.content_type }}</div>
                  </div>
                </div>
              </div>
              <div v-else class="text-center text-muted mt-6 mb-6">
                👆 Выберите группу, чтобы посмотреть её тайтлы
              </div>
            </template>
          </div>
        </div>
      </div>
    </transition>

    <div class="room-content" :class="{ 'theater-mode': isTheaterMode }">
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
        <div class="anime-controls-panel glass animate-fade-in" v-if="!roomState.error && hasAnimeMetadata && (roomState.videoType === 'kodik' || roomState.videoType === 'alloha' || roomState.videoType === 'direct')">
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

        
        <!-- Инфо для Theater Mode -->
        <div class="theater-media-info glass animate-fade-in mt-4 p-4 flex gap-4" v-if="isTheaterMode && hasAnimeMetadata && !isLoadingShikimoriDetails">
          <img v-if="shikimoriDetails?.image?.original" :src="'https://shikimori.one' + shikimoriDetails.image.original" class="w-[120px] rounded-md shadow-lg flex-shrink-0" style="align-self: flex-start;">
          <div v-else class="w-[120px] h-[160px] rounded-md bg-white/5 flex items-center justify-center text-2xl flex-shrink-0">📺</div>
          <div class="flex-1 min-w-0">
            <h2 class="text-xl font-bold mb-1">{{ metadataTitle }}</h2>
            <div class="text-sm text-gray-400 mb-3">{{ metadataTitleEn }}</div>
            <div class="flex flex-wrap gap-1 mb-3">
              <span class="badge bg-violet-500/20 text-violet-300 text-xs px-2 py-0.5 rounded">{{ shikimoriDetails?.kind?.toUpperCase() }}</span>
              <span class="badge bg-green-500/20 text-green-300 text-xs px-2 py-0.5 rounded">★ {{ shikimoriDetails?.score }}</span>
              <span v-if="shikimoriDetails?.episodes" class="badge bg-gray-500/20 text-gray-300 text-xs px-2 py-0.5 rounded">{{ shikimoriDetails?.episodes }} эп.</span>
            </div>
          </div>
        </div>

        <!-- Информация о медиа (Только описание и персонажи) -->
        <div class="media-info-panel glass animate-fade-in mt-4" v-if="hasAnimeMetadata">
          <div v-if="isLoadingShikimoriDetails" class="loading-state">
            <div class="spinner-small"></div>
            <span>Загрузка информации об аниме...</span>
          </div>
          <div v-else-if="shikimoriDetails" class="media-info-content">
            <div class="media-description text-sm" v-html="shikimoriDetails.description_html || shikimoriDetails.description"></div>
            
            <div class="media-characters-section mt-4" v-if="shikimoriMainCharacters.length > 0">
              <button class="btn btn-outline btn-sm" @click="showCharacters = !showCharacters">
                {{ showCharacters ? 'Скрыть персонажей' : 'Показать персонажей' }}
              </button>
              <div v-if="showCharacters" class="characters-slider mt-3">
                <div v-for="role in shikimoriMainCharacters" :key="role.character.id" class="character-card">
                  <img :src="`https://shikimori.io${role.character.image.original}`" class="character-img" />
                  <div class="character-info">
                    <div class="character-name" :title="role.character.russian || role.character.name">{{ role.character.russian || role.character.name }}</div>
                    <div class="character-role">{{ role.roles_russian[0] || role.roles[0] }}</div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <aside class="side-panel glass">
        <div class="participants-section">
          <h3>Участники ({{ roomState.participants.length }})</h3>
          <ul class="participant-list custom-scroll">
            <li v-for="p in roomState.participants" :key="p.user_id" class="participant-item">
              <span class="status-dot"></span>
              <span class="name">{{ p.username }} <span v-if="p.is_owner" style="color:var(--text-muted); font-size:0.85em;">(Host)</span></span>
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

        <!-- Информация об аниме в сайдбаре -->
        <div class="sidebar-media-info mt-4" v-if="hasAnimeMetadata && shikimoriDetails">
          <img :src="`https://shikimori.io${shikimoriDetails.image?.original}`" class="sidebar-media-poster" alt="Постер" />
          <div class="sidebar-media-details">
            <h2 class="sidebar-media-title" :class="{'expanded': isTitleExpanded}">{{ shikimoriDetails.russian || shikimoriDetails.name }}</h2>
            <button v-if="isTitleLong" @click="isTitleExpanded = !isTitleExpanded" class="btn-text-expand">
              {{ isTitleExpanded ? 'Свернуть' : 'Развернуть' }}
            </button>
            <div class="sidebar-media-rating mt-2">
              <span class="rating-star" style="color: #fbbf24; font-size: 16px;">★</span>
              <span style="font-weight: bold; font-size: 16px; margin-left: 4px;">{{ shikimoriDetails.score }}</span>
              <span style="color: rgba(255,255,255,0.5); font-size: 12px; margin-left: 6px;">({{ shikimoriVotes }} оценок)</span>
            </div>
          </div>
        </div>
        <div class="sidebar-media-info mt-4" v-else-if="!hasAnimeMetadata && roomState.videoUrl">
          <div class="sidebar-media-poster-placeholder">
            🎬
          </div>
          <div class="sidebar-media-details">
            <h2 class="sidebar-media-title">Видео</h2>
            <a :href="roomState.videoUrl" target="_blank" class="sidebar-media-link mt-1 inline-block">
              🔗 Ссылка на видео
            </a>
          </div>
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
import { ref, computed, watch, onMounted } from 'vue'
import { useRoute } from 'vue-router'



async function syncProgressWithList(episodeNum) {
  if (!shikimoriDetails.value || !auth.isAuthenticated) return
  const shikiId = shikimoriDetails.value.id
  
  // Try to find in personal reviews
  if (reviewsStore.reviews.length === 0) {
    await reviewsStore.fetchReviews()
  }
  const review = reviewsStore.reviews.find(r => r.shikimori_id === shikiId)
  if (review && review.current_episode < episodeNum) {
    console.log(`Syncing progress to ${episodeNum} for review ${review.id}`)
    const payload = {
      title: review.title,
      content_type: review.content_type,
      status: review.status,
      rating: review.rating,
      notes: review.notes,
      poster_url: review.poster_url,
      shikimori_id: review.shikimori_id,
      description: review.description,
      episodes_total: review.episodes_total,
      current_episode: episodeNum,
      aniliberty_alias: review.aniliberty_alias,
      shikimori_score: review.shikimori_score
    }
    await reviewsStore.updateReview(review.id, payload)
  }
}


import { useReviewsStore } from '@/stores/reviews'
import { useGroupsStore } from '@/stores/groups'
import { useWatchParty } from '@/composables/useWatchParty'
import { useAuthStore } from '@/stores/auth'
import YouTubePlayer from '@/components/watchparty/YouTubePlayer.vue'
import DirectVideoPlayer from '@/components/watchparty/DirectVideoPlayer.vue'
import RutubePlayer from '@/components/watchparty/RutubePlayer.vue'
import KodikVideoPlayer from '@/components/watchparty/KodikVideoPlayer.vue'
import AllohaVideoPlayer from '@/components/watchparty/AllohaVideoPlayer.vue'

const route = useRoute()
const roomId = route.params.roomId
const auth = useAuthStore()
const reviewsStore = useReviewsStore()
const groupsStore = useGroupsStore()

const {
  roomState,
  playerRef,
  connect,
  onLocalPlay,
  onLocalPause,
  onLocalSeek,
  changeVideo,
  kickUser,
  admitUser,
  rejectUser,
  updateMetadata
} = useWatchParty()

const editUrl = ref('')
const showMediaSelector = ref(false)
const selectorTab = ref('personal')
const selectedGroupId = ref(null)

function toggleMediaSelector() {
  console.log("toggleMediaSelector clicked", showMediaSelector.value);
  showMediaSelector.value = !showMediaSelector.value;
  if (showMediaSelector.value) {
    if (reviewsStore.reviews.length === 0 && !reviewsStore.loading) {
      reviewsStore.fetchReviews();
    }
    if (groupsStore.groups.length === 0 && !groupsStore.loading) {
      groupsStore.fetchGroups();
    }
  }
}

function toggleGroupSelection(id) {
  if (selectedGroupId.value === id) {
    selectedGroupId.value = null // Скрываем, если нажали повторно
  } else {
    selectedGroupId.value = id
  }
}

const selectedGroupItems = computed(() => {
  if (!selectedGroupId.value) return []
  const group = groupsStore.groups.find(g => g.id === selectedGroupId.value)
  return group ? group.items : []
})

function selectMediaToWatch(item) {
  if (!roomState.isOwner) {
    alert('Только хост комнаты может переключать видео!')
    return
  }
  
  const shId = item.shikimori_id
  const alias = item.aniliberty_alias
  
  if (!shId && !alias && (!item.links || item.links.length === 0)) {
    alert('У этого тайтла нет привязок для воспроизведения.')
    return
  }

  // Строим URL как в ReviewView /handleWatchTogether
  let urlStr = ''
  if (shId && alias) {
    urlStr = `https://anilibria.top/api/v1/anime/player?shikimori=${shId}&alias=${alias}&episode=1`
  } else if (shId) {
    urlStr = `https://kodik.info/find-player?shikimoriID=${shId}&types=anime,anime-serial`
  } else if (item.links && item.links.length > 0) {
    urlStr = item.links[0].url
  } else {
    urlStr = `https://kodik.info/find-player?title=${encodeURIComponent(item.title)}`
  }

  changeVideo(urlStr, 'kodik')
  showMediaSelector.value = false
}

const isTheaterMode = ref(false)
const isTitleExpanded = ref(false)
const isTitleLong = computed(() => {
  if (!shikimoriDetails.value) return false
  const title = shikimoriDetails.value.russian || shikimoriDetails.value.name
  return title && title.length > 35 // rough heuristic for 2 lines
})

const activeMirror = 'video.kodik.online' // можно использовать aniqit.com или другой
const currentShikimoriId = ref(sessionStorage.getItem(`wp_shikimori_${roomId}`) || '')
const currentAlias = ref(sessionStorage.getItem(`wp_alias_${roomId}`) || '')

const shikimoriDetails = ref(null)
const shikimoriRoles = ref([])
const showCharacters = ref(false)
const isLoadingShikimoriDetails = ref(false)

const shikimoriVotes = computed(() => {
  if (!shikimoriDetails.value || !shikimoriDetails.value.rates_scores_stats) return 0
  return shikimoriDetails.value.rates_scores_stats.reduce((acc, stat) => acc + stat.value, 0)
})

const shikimoriMainCharacters = computed(() => {
  if (!shikimoriRoles.value) return []
  return shikimoriRoles.value.filter(r => r.character)
})

const hasAnimeMetadata = computed(() => {
  return currentShikimoriId.value !== ''
})

watch([() => roomState.isConnected, currentShikimoriId, currentAlias], ([connected, shId, alias]) => {
  if (connected && roomState.isOwner) {
    updateMetadata(shId, alias)
  }
})

// Кастомное управление озвучками и сериями для Kodik
const roomTranslations = ref([])
const loadingTranslations = ref(false)
const translationsError = ref(null)

async function fetchRoomTranslations() {
  loadingTranslations.value = true
  translationsError.value = null
  
  const token = auth.accessToken
  let kodikPromise = Promise.resolve([])
  let anilibriaPromise = Promise.resolve([])
  
  if (currentShikimoriId.value) {
    kodikPromise = fetch(`/api/v1/reviews/integrations/kodik/search?shikimori_id=${currentShikimoriId.value}`, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    })
    .then(res => res.ok ? res.json() : { results: [] })
    .then(data => data.results || [])
    .catch(e => {
      console.error("Failed to fetch Kodik translations", e)
      return []
    })
  }
  
  if (currentAlias.value) {
    anilibriaPromise = fetch(`/api/v1/reviews/integrations/aniliberty/episodes/${currentAlias.value}`, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    })
    .then(res => res.ok ? res.json() : { episodes: [] })
    .then(data => data.episodes || [])
    .catch(e => {
      console.error("Failed to fetch Anilibria episodes", e)
      return []
    })
  }
  
  try {
    const [kodikResults, anilibriaEpisodesList] = await Promise.all([kodikPromise, anilibriaPromise])
    
    // Sort Anilibria episodes
    anilibriaEpisodesList.sort((a, b) => {
      const aNum = parseFloat(a.ordinal ?? a.number) || 0
      const bNum = parseFloat(b.ordinal ?? b.number) || 0
      return aNum - bNum
    })
    
    let syntheticAnilibria = null
    if (anilibriaEpisodesList.length > 0) {
      const episodesMap = {}
      anilibriaEpisodesList.forEach(ep => {
        const epNum = ep.ordinal ?? ep.number
        const hlsUrl = ep.hls_1080 || ep.hls_720 || ep.hls_480 || ep.hls
        if (hlsUrl) {
          episodesMap[String(epNum)] = hlsUrl
        }
      })
      
      syntheticAnilibria = {
        id: 'anilibria_api',
        translation: {
          id: 'anilibria_api',
          title: 'AniLibria (API)',
          type: 'voice'
        },
        last_episode: anilibriaEpisodesList.length,
        episodes_count: anilibriaEpisodesList.length,
        seasons: {
          "1": {
            episodes: episodesMap
          }
        },
        isAnilibriaApi: true
      }
    }
    
    const allTranslations = []
    if (syntheticAnilibria) {
      allTranslations.push(syntheticAnilibria)
    }
    if (kodikResults.length > 0) {
      allTranslations.push(...kodikResults)
    }
    
    roomTranslations.value = allTranslations
  } catch (e) {
    console.error(e)
    translationsError.value = e.message
  } finally {
    loadingTranslations.value = false
  }
}

watch([currentShikimoriId, currentAlias], async () => {
  fetchRoomTranslations()

  if (currentShikimoriId.value) sessionStorage.setItem(`wp_shikimori_${roomId}`, currentShikimoriId.value)
  if (currentAlias.value) sessionStorage.setItem(`wp_alias_${roomId}`, currentAlias.value)

  if (currentShikimoriId.value && hasAnimeMetadata.value) {
    isLoadingShikimoriDetails.value = true
    try {
      const res = await fetch(`https://shikimori.io/api/animes/${currentShikimoriId.value}`)
      if (res.ok) shikimoriDetails.value = await res.json()
      
      const rolesRes = await fetch(`https://shikimori.io/api/animes/${currentShikimoriId.value}/roles`)
      if (rolesRes.ok) shikimoriRoles.value = await rolesRes.json()
    } catch (e) {
      console.error('Failed to load shikimori metadata', e)
    } finally {
      isLoadingShikimoriDetails.value = false
    }
  } else {
    shikimoriDetails.value = null
    shikimoriRoles.value = []
    showCharacters.value = false
  }
}, { immediate: true })

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

  // Если всё ещё нет ссылки — используем serial-уровень link
  if (!baseLink) {
    baseLink = t.link || ''
  }

  const shId = currentShikimoriId.value
  const alias = currentAlias.value
  
  if (t.isAnilibriaApi) {
    if (baseLink) {
      try {
        const urlObj = new URL(baseLink)
        urlObj.searchParams.set('shikimori', shId)
        urlObj.searchParams.set('alias', alias)
        urlObj.searchParams.set('episode', String(targetEp))
        urlObj.searchParams.set('translation_id', 'anilibria_api')
        changeVideo(urlObj.toString(), 'direct')
      } catch (e) {
        changeVideo(baseLink, 'direct')
      }
    } else {
      alert('Поток Anilibria не найден для этой серии')
    }
  } else if (roomState.videoType === 'alloha') {
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
      if (t.translation.id) urlObj.searchParams.set('translation_id', String(t.translation.id))
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

  const activeTrans = activeTranslationData.value
  
  if (activeTrans?.isAnilibriaApi) {
    const baseLink = activeTrans.seasons["1"]?.episodes?.[episodeNum]
    if (baseLink) {
      try {
        const urlObj = new URL(baseLink)
        urlObj.searchParams.set('shikimori', shId)
        urlObj.searchParams.set('alias', alias)
        urlObj.searchParams.set('episode', String(episodeNum))
        urlObj.searchParams.set('translation_id', 'anilibria_api')
        changeVideo(urlObj.toString(), 'direct')
      } catch (e) {
        changeVideo(baseLink, 'direct')
      }
    } else {
      alert(`Серия ${episodeNum} не найдена в Anilibria API`)
    }
  } else if (roomState.videoType === 'alloha') {
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
      if (translationId) urlObj.searchParams.set('translation_id', String(translationId))
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
      if (translationId) urlObj.searchParams.set('translation_id', String(translationId))
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
    const anilibriaTrans = roomTranslations.value.find(t => t.isAnilibriaApi)
    const epLink = anilibriaTrans?.seasons?.["1"]?.episodes?.[ep]
    
    if (epLink) {
      try {
        const urlObj = new URL(epLink)
        urlObj.searchParams.set('shikimori', shId)
        urlObj.searchParams.set('alias', alias)
        urlObj.searchParams.set('episode', String(ep))
        urlObj.searchParams.set('translation_id', 'anilibria_api')
        changeVideo(urlObj.toString(), 'direct')
      } catch (e) {
        changeVideo(epLink, 'direct')
      }
    } else {
      loadingDirect.value = true
      try {
        const token = auth.accessToken
        const res = await fetch(`/api/v1/reviews/integrations/aniliberty/episodes/${alias}`, {
          headers: {
            'Authorization': `Bearer ${token}`
          }
        })
        const data = await res.json()
        const episodes = data.episodes || []
        
        const foundEp = episodes.find(e => (e.ordinal ?? e.number) == ep)
        if (foundEp) {
          const hlsUrl = foundEp.hls_1080 || foundEp.hls_720 || foundEp.hls_480 || foundEp.hls
          if (hlsUrl) {
            const urlObj = new URL(hlsUrl)
            urlObj.searchParams.set('shikimori', shId)
            urlObj.searchParams.set('alias', alias)
            urlObj.searchParams.set('episode', String(ep))
            urlObj.searchParams.set('translation_id', 'anilibria_api')
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
  position: relative;
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

/* Стили для панели информации о медиа */
.media-info-panel {
  padding: 20px;
  border-radius: 12px;
  border: 1px solid rgba(255, 255, 255, 0.08);
}
.media-poster-placeholder {
  width: 120px;
  height: 170px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 8px;
  font-size: 40px;
  flex-shrink: 0;
}
.media-poster {
  width: 140px;
  height: auto;
  border-radius: 8px;
  object-fit: cover;
  box-shadow: 0 4px 12px rgba(0,0,0,0.3);
  flex-shrink: 0;
}
.media-title {
  font-size: 1.4rem;
  font-weight: 700;
  margin: 0;
  color: #fff;
}
.media-description {
  color: rgba(255, 255, 255, 0.8);
  line-height: 1.5;
  max-height: 200px;
  overflow-y: auto;
  padding-right: 8px;
}
.media-description::-webkit-scrollbar {
  width: 4px;
}
.media-description::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 2px;
}

.characters-slider {
  display: flex;
  gap: 12px;
  overflow-x: auto;
  padding-bottom: 8px;
  scroll-snap-type: x mandatory;
}
.characters-slider::-webkit-scrollbar {
  height: 6px;
}
.characters-slider::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 3px;
}
.character-card {
  display: flex;
  background: rgba(255, 255, 255, 0.04);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 8px;
  overflow: hidden;
  align-items: center;
  flex: 0 0 160px; /* fixed width for horizontal scroll */
  scroll-snap-align: start;
}
.character-img {
  width: 40px;
  height: 56px;
  object-fit: cover;
  flex-shrink: 0;
}
.character-info {
  padding: 6px;
  display: flex;
  flex-direction: column;
  justify-content: center;
  overflow: hidden;
}
.character-name {
  font-size: 0.75rem;
  font-weight: 600;
  color: #fff;
  line-height: 1.2;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.character-role {
  font-size: 0.65rem;
  color: rgba(255, 255, 255, 0.5);
  margin-top: 4px;
}

/* Стили для сайдбара и theater-mode */
.participant-list.custom-scroll {
  max-height: 160px; /* ~4-5 participants */
  overflow-y: auto;
  padding-right: 4px;
}
.participant-list.custom-scroll::-webkit-scrollbar {
  width: 4px;
}
.participant-list.custom-scroll::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 2px;
}

.sidebar-media-info {
  display: flex;
  flex-direction: column;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.05);
  border-radius: 8px;
  padding: 12px;
}
.sidebar-media-poster {
  width: 100%;
  border-radius: 6px;
  object-fit: cover;
  box-shadow: 0 4px 12px rgba(0,0,0,0.3);
  margin-bottom: 12px;
}
.sidebar-media-poster-placeholder {
  width: 100%;
  height: 180px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 6px;
  display: flex;
  justify-content: center;
  align-items: center;
  font-size: 40px;
  margin-bottom: 12px;
}
.sidebar-media-title {
  font-size: 1.1rem;
  font-weight: 700;
  margin: 0;
  color: #fff;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}
.sidebar-media-title.expanded {
  -webkit-line-clamp: unset;
  display: block;
}
.btn-text-expand {
  background: none;
  border: none;
  color: #60a5fa;
  font-size: 0.8rem;
  cursor: pointer;
  padding: 0;
  margin-top: 4px;
}
.sidebar-media-link {
  color: #60a5fa;
  text-decoration: none;
  font-size: 0.85rem;
}

.room-content.theater-mode {
  flex-direction: column;
  overflow-y: auto;
}
.room-content.theater-mode .main-area {
  overflow-y: visible;
  padding: 16px 24px;
}
.room-content.theater-mode .player-wrapper {
  flex: none;
  height: 75vh;
  max-height: calc(100vh - 120px);
  min-height: unset;
}
.room-content.theater-mode .side-panel {
  width: 100%;
  border-left: none;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  flex-direction: row;
  flex-wrap: wrap;
}
.room-content.theater-mode .side-panel .participants-section {
  flex: 1;
  min-width: 300px;
}
.room-content.theater-mode .sidebar-media-info {
  display: none !important;
}
.room-content.theater-mode .sidebar-media-poster {
  width: 120px;
  margin-bottom: 0;
}
.room-content.theater-mode .sidebar-media-poster-placeholder {
  width: 120px;
  height: 160px;
  margin-bottom: 0;
}

.btn-sm {
  padding: 6px 12px;
  font-size: 0.8rem;
}

/* --- Стили для выпадающей панели медиа (Media Selector) --- */
.media-selector-drawer {
  position: absolute;
  top: 70px;
  left: 50%;
  transform: translateX(-50%);
  width: 90%;
  max-width: 800px;
  max-height: 70vh;
  background: rgba(20, 20, 20, 0.95);
  backdrop-filter: blur(20px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  z-index: 100;
  display: flex;
  flex-direction: column;
  box-shadow: 0 10px 40px rgba(0,0,0,0.5);
  padding: 20px;
}

.slide-down-enter-active, .slide-down-leave-active {
  transition: all 0.3s cubic-bezier(0.25, 0.8, 0.25, 1);
}
.slide-down-enter-from, .slide-down-leave-to {
  opacity: 0;
  transform: translate(-50%, -20px);
}

.drawer-tab {
  background: rgba(255, 255, 255, 0.05);
  border: none;
  color: rgba(255, 255, 255, 0.6);
  padding: 8px 16px;
  border-radius: 12px;
  font-size: 0.9rem;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}
.drawer-tab:hover {
  background: rgba(255, 255, 255, 0.1);
}
.drawer-tab.active {
  background: var(--primary-color, #60a5fa);
  color: #fff;
}

.drawer-content {
  overflow-y: auto;
  flex: 1;
  padding-right: 8px;
}
.drawer-content::-webkit-scrollbar {
  width: 6px;
}
.drawer-content::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 3px;
}

.drawer-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, 1fr));
  gap: 16px;
}

.media-item-card {
  display: flex;
  flex-direction: column;
  background: rgba(255, 255, 255, 0.03);
  border-radius: 8px;
  overflow: hidden;
  cursor: pointer;
  transition: all 0.2s;
  border: 1px solid transparent;
}
.media-item-card:hover {
  transform: translateY(-2px);
  background: rgba(255, 255, 255, 0.08);
  border-color: rgba(255, 255, 255, 0.2);
}

.media-item-poster {
  width: 100%;
  aspect-ratio: 2/3;
  object-fit: cover;
}

.media-item-info {
  padding: 8px;
  flex: 1;
  display: flex;
  flex-direction: column;
}

.media-item-title {
  font-size: 0.8rem;
  font-weight: 600;
  color: #fff;
  line-height: 1.2;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
  margin-bottom: 4px;
}

.media-item-type {
  font-size: 0.7rem;
  color: rgba(255, 255, 255, 0.5);
  margin-top: auto;
  text-transform: capitalize;
}

.groups-slider {
  display: flex;
  gap: 10px;
  overflow-x: auto;
  padding-bottom: 8px;
}
.groups-slider::-webkit-scrollbar {
  height: 4px;
}
.groups-slider::-webkit-scrollbar-thumb {
  background: rgba(255, 255, 255, 0.2);
  border-radius: 2px;
}

.group-pill {
  background: rgba(255, 255, 255, 0.08);
  border: 1px solid rgba(255, 255, 255, 0.1);
  padding: 8px 16px;
  border-radius: 20px;
  color: #fff;
  font-size: 0.85rem;
  font-weight: 500;
  cursor: pointer;
  white-space: nowrap;
  transition: all 0.2s;
}
.group-pill:hover {
  background: rgba(255, 255, 255, 0.15);
}
.group-pill.active {
  background: #10b981;
  border-color: #34d399;
  color: #fff;
}
</style>
