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
          <div v-else-if="!roomState.error" class="empty-player">
            Видео не выбрано
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
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useWatchParty } from '@/composables/useWatchParty'
import YouTubePlayer from '@/components/watchparty/YouTubePlayer.vue'
import DirectVideoPlayer from '@/components/watchparty/DirectVideoPlayer.vue'
import RutubePlayer from '@/components/watchparty/RutubePlayer.vue'
import KodikVideoPlayer from '@/components/watchparty/KodikVideoPlayer.vue'

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

function detectVideoType(url) {
  if (!url) return ''
  if (/youtube\.com|youtu\.be/.test(url)) return 'youtube'
  if (/rutube\.ru/.test(url)) return 'rutube'
  if (/kodik|aniqit/i.test(url)) return 'kodik'
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

    // Добавляем схему протокола, если ссылка относительная
    if (url.startsWith('//')) {
      url = window.location.protocol + url
    }

    const urlObj = new URL(url)
    urlObj.searchParams.set('episode', episodeNum)
    const newUrl = urlObj.toString()

    // Меняем видео для всей комнаты только если оно действительно изменилось
    if (newUrl !== roomState.videoUrl) {
      changeVideo(newUrl, 'kodik')
    }
  } catch (e) {
    console.error('Failed to change episode via owner action', e)
  }
}

function copyLink() {
  const url = window.location.origin + `/watch/room/${roomId}`
  navigator.clipboard.writeText(url)
}

onMounted(() => {
  const token = auth.accessToken
  connect(roomId, token)

  // Check if we just created it with an initial URL
  const initialUrl = sessionStorage.getItem(`wp_url_${roomId}`)
  if (initialUrl) {
    sessionStorage.removeItem(`wp_url_${roomId}`)
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
</style>
