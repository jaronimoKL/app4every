<template>
  <div class="lobby-container">
    <div class="lobby-card">
      <h1 class="title">📺 Совместный просмотр</h1>
      
      <div class="section create-room">
        <h3>Создать комнату</h3>
        <label>Ссылка на видео (YouTube, Rutube, .mp4):</label>
        <input 
          v-model="newRoomUrl" 
          type="text" 
          placeholder="https://youtube.com/watch?v=..." 
          class="input-field"
        />
        <button @click="createRoom" class="btn primary-btn" :disabled="!newRoomUrl">
          🎬 Создать комнату
        </button>
      </div>

      <div class="divider">
        <span>или войти по коду</span>
      </div>

      <div class="section join-room">
        <label>Код комнаты:</label>
        <div class="join-row">
          <input 
            v-model="joinRoomId" 
            type="text" 
            placeholder="abc123" 
            class="input-field"
          />
          <button @click="joinRoom" class="btn secondary-btn" :disabled="!joinRoomId">
            Войти →
          </button>
        </div>
      </div>

      <div v-if="activeRooms.length > 0" class="active-rooms-section mt-6">
        <div class="divider">
          <span>Активные комнаты друзей</span>
        </div>
        <div class="rooms-list flex flex-col gap-3 mt-3">
          <div 
            v-for="room in activeRooms" 
            :key="room.room_id"
            class="active-room-card glass p-3 rounded-lg flex justify-between items-center"
          >
            <div class="room-info">
              <div class="room-title flex items-center gap-2">
                <span class="online-dot"></span>
                <span class="font-semibold">Комната {{ room.room_id.substring(0, 4) }}</span>
              </div>
              <div class="room-meta text-xs text-gray-400 mt-1">
                Участников: {{ room.participants?.length || 1 }} 
                <span v-if="room.shikimori_id || room.aniliberty_alias"> • Смотрят аниме</span>
              </div>
            </div>
            <button class="btn primary-btn btn-sm" @click="joinSpecificRoom(room.room_id)">
              Присоединиться
            </button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
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

onMounted(async () => {
  await fetchActiveRooms()
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
.online-dot {
  width: 8px;
  height: 8px;
  background-color: #4ade80;
  border-radius: 50%;
  display: inline-block;
  box-shadow: 0 0 8px rgba(74, 222, 128, 0.5);
}
.btn-sm {
  padding: 8px 16px;
  font-size: 0.85rem;
}
.lobby-container {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: calc(100vh - 60px);
  background: var(--bg-color, #1a1a1a);
  color: #fff;
  padding: 20px;
}

.lobby-card {
  background: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 16px;
  padding: 30px;
  width: 100%;
  max-width: 400px;
  box-shadow: 0 10px 30px rgba(0, 0, 0, 0.3);
}

.title {
  margin-top: 0;
  margin-bottom: 24px;
  font-size: 1.5rem;
  text-align: center;
  color: #e0e0e0;
}

.section {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

h3 {
  margin: 0;
  font-size: 1.1rem;
  color: #fff;
}

label {
  font-size: 0.9rem;
  color: #aaa;
}

.input-field {
  background: rgba(0, 0, 0, 0.2);
  border: 1px solid rgba(255, 255, 255, 0.1);
  padding: 12px;
  border-radius: 8px;
  color: #fff;
  font-size: 1rem;
  transition: border-color 0.2s;
}

.input-field:focus {
  outline: none;
  border-color: #3b82f6;
}

.btn {
  padding: 12px;
  border: none;
  border-radius: 8px;
  font-size: 1rem;
  font-weight: 600;
  cursor: pointer;
  transition: opacity 0.2s, transform 0.1s;
}

.btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn:active:not(:disabled) {
  transform: scale(0.98);
}

.primary-btn {
  background: #3b82f6;
  color: #fff;
}

.secondary-btn {
  background: rgba(255, 255, 255, 0.1);
  color: #fff;
  white-space: nowrap;
}

.join-row {
  display: flex;
  gap: 8px;
}

.join-row .input-field {
  flex: 1;
}

.divider {
  display: flex;
  align-items: center;
  text-align: center;
  margin: 24px 0;
  color: #666;
  font-size: 0.9rem;
}

.divider::before,
.divider::after {
  content: '';
  flex: 1;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
}

.divider span {
  padding: 0 10px;
}
</style>
