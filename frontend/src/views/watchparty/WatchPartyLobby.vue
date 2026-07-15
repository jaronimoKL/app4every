<template>
  <div class="lobby-container flex flex-col">
    <!-- Header with Back Button -->
    <div class="w-full max-w-[900px] mb-6 flex justify-between items-center px-4 md:px-0">
      <button class="btn btn-ghost flex items-center gap-2 text-sm" @click="$router.back()">
        <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round" style="width:14px; height:14px;">
          <line x1="19" y1="12" x2="5" y2="12"></line>
          <polyline points="12 19 5 12 12 5"></polyline>
        </svg>
        Назад
      </button>
    </div>

    <div class="lobby-content">
      <!-- Left side: Active Rooms -->
      <div class="active-rooms-panel panel-card glass">
        <h2 class="panel-title flex items-center gap-2">
          <span class="online-dot-large"></span>
          Активные комнаты
        </h2>
        <div class="divider">
          <span>Друзья и участники групп</span>
        </div>
        
        <div v-if="loading && activeRooms.length === 0" class="text-center py-8 text-muted">
          <div class="spinner mx-auto mb-2" style="width:24px;height:24px;"></div>
          Загрузка...
        </div>
        <div v-else-if="activeRooms.length === 0" class="text-center py-8 text-muted text-sm">
          Нет активных комнат
        </div>
        <div v-else class="rooms-list flex flex-col gap-3 mt-3 overflow-y-auto max-h-[400px] pr-2 custom-scrollbar">
          <div 
            v-for="room in activeRooms" 
            :key="room.room_id"
            class="active-room-card glass glass-hover p-4 rounded-lg flex justify-between items-center cursor-pointer"
            @click="joinSpecificRoom(room.room_id)"
          >
            <div class="room-info">
              <div class="room-title flex items-center gap-2">
                <span class="online-dot"></span>
                <span class="font-semibold text-primary">Комната {{ room.room_id.substring(0, 4) }}</span>
              </div>
              <div class="room-meta text-xs text-muted mt-1.5 flex items-center gap-2">
                <span class="flex items-center gap-1">👥 {{ room.participants?.length || 1 }}</span>
                <span v-if="room.shikimori_id || room.aniliberty_alias" style="color: var(--primary);">• Смотрят аниме</span>
              </div>
            </div>
            <button class="btn btn-outline btn-sm" @click.stop="joinSpecificRoom(room.room_id)">
              Войти
            </button>
          </div>
        </div>
      </div>

      <!-- Right side: Create/Join -->
      <div class="lobby-card panel-card glass">
        <h1 class="title">📺 Watch Party</h1>
        
        <div class="section create-room">
          <h3>Создать комнату</h3>
          <label class="form-label">Ссылка на видео (YouTube, Rutube, .mp4):</label>
          <input 
            v-model="newRoomUrl" 
            type="text" 
            placeholder="https://youtube.com/watch?v=..." 
            class="form-input"
            @keyup.enter="createRoom"
          />
          <button @click="createRoom" class="btn btn-primary mt-2" :disabled="!newRoomUrl">
            🎬 Создать комнату
          </button>
        </div>

        <div class="divider">
          <span>или войти по коду</span>
        </div>

        <div class="section join-room">
          <label class="form-label">Код комнаты:</label>
          <div class="join-row">
            <input 
              v-model="joinRoomId" 
              type="text" 
              placeholder="abc123" 
              class="form-input"
              @keyup.enter="joinRoom"
            />
            <button @click="joinRoom" class="btn btn-outline" :disabled="!joinRoomId">
              Войти →
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useGroupsStore } from '@/stores/groups'
import { friendsApi, watchpartyApi } from '@/services/api'

const router = useRouter()
const authStore = useAuthStore()
const groupsStore = useGroupsStore()

const newRoomUrl = ref('')
const joinRoomId = ref('')
const activeRooms = ref([])
const loading = ref(true)

let pollingInterval = null

onMounted(async () => {
  await fetchActiveRooms()
  pollingInterval = setInterval(fetchActiveRooms, 10000) // 10 seconds polling
})

onUnmounted(() => {
  if (pollingInterval) clearInterval(pollingInterval)
})

async function fetchActiveRooms() {
  try {
    const userIds = new Set()
    
    // Get friends
    try {
      const friends = await friendsApi.getFriends(authStore.accessToken)
      if (friends) {
        friends.forEach(f => userIds.add(f.id))
      }
    } catch (e) {
      console.error('Failed to load friends for active rooms:', e)
    }

    // Get group members
    try {
      if (groupsStore.groups.length === 0) {
        await groupsStore.fetchGroups()
      }
      groupsStore.groups.forEach(g => {
        if (g.members) {
          g.members.forEach(m => userIds.add(m.user_id))
        }
      })
    } catch (e) {
      console.error('Failed to load groups for active rooms:', e)
    }

    if (userIds.size > 0) {
      const rooms = await watchpartyApi.getActiveRooms(Array.from(userIds))
      activeRooms.value = rooms || []
    }
  } catch (err) {
    console.error('Failed to fetch active rooms:', err)
  } finally {
    loading.value = false
  }
}

function generateUUID() {
  return Math.random().toString(36).substring(2, 10)
}

function createRoom() {
  if (!newRoomUrl.value) return
  const roomId = generateUUID()
  sessionStorage.setItem(`wp_url_${roomId}`, newRoomUrl.value)
  router.push(`/watch/room/${roomId}`)
}

function joinRoom() {
  if (!joinRoomId.value) return
  router.push(`/watch/room/${joinRoomId.value}`)
}

function joinSpecificRoom(id) {
  router.push(`/watch/room/${id}`)
}
</script>

<style scoped>
.lobby-container {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: flex-start;
  min-height: calc(100vh - 60px);
  background-color: var(--bg-base);
  color: var(--text-primary);
  padding: 40px 20px;
}

.lobby-content {
  display: flex;
  gap: 30px;
  width: 100%;
  max-width: 900px;
  align-items: flex-start;
}

@media (max-width: 768px) {
  .lobby-content {
    flex-direction: column;
    align-items: center;
    gap: 20px;
  }
}

.panel-card {
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: 30px;
  box-shadow: var(--shadow-md);
}

.active-rooms-panel {
  flex: 1;
  width: 100%;
  min-height: 380px;
}

.lobby-card {
  width: 100%;
  max-width: 400px;
}

.panel-title {
  font-size: 18px;
  font-weight: 700;
  margin-top: 0;
  margin-bottom: 8px;
  color: var(--text-primary);
}

.online-dot-large {
  width: 12px;
  height: 12px;
  background-color: var(--matcha);
  border-radius: 50%;
  display: inline-block;
  box-shadow: 0 0 10px rgba(124, 154, 110, 0.6);
}

.online-dot {
  width: 8px;
  height: 8px;
  background-color: var(--matcha);
  border-radius: 50%;
  display: inline-block;
  box-shadow: 0 0 8px rgba(124, 154, 110, 0.5);
}

.title {
  margin-top: 0;
  margin-bottom: 24px;
  font-size: 22px;
  text-align: center;
  font-weight: 800;
  color: var(--text-primary);
}

.section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 700;
  color: var(--text-primary);
}

.form-label {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-muted);
  margin-bottom: -4px;
}

.btn-sm {
  padding: 6px 12px;
  font-size: 12px;
}

.join-row {
  display: flex;
  gap: 8px;
}

.join-row .form-input {
  flex: 1;
}

.divider {
  display: flex;
  align-items: center;
  text-align: center;
  margin: 24px 0;
  color: var(--text-muted);
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.divider::before,
.divider::after {
  content: '';
  flex: 1;
  border-bottom: 1px solid var(--border);
}

.divider span {
  padding: 0 10px;
}

.custom-scrollbar::-webkit-scrollbar {
  width: 6px;
}
.custom-scrollbar::-webkit-scrollbar-track {
  background: transparent;
}
.custom-scrollbar::-webkit-scrollbar-thumb {
  background: var(--border);
  border-radius: 4px;
}
</style>
