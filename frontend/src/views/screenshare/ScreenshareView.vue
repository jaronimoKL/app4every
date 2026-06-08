<template>
  <div class="lobby-page">
    <!-- ══ НАВБАР ══ -->
    <nav class="lobby-nav glass">
      <div class="lobby-nav-inner flex items-center gap-3">
        <RouterLink to="/dashboard" class="btn btn-ghost" style="padding:7px 12px;font-size:13px;">← Назад</RouterLink>
        <div class="nav-sep"></div>
        <span style="font-size:18px;">📹</span>
        <span style="font-weight:700;font-size:15px;margin-right:8px;">Видеочат и Экран</span>
      </div>
    </nav>

    <!-- ══ ОСНОВНОЙ КОНТЕНТ ══ -->
    <div class="lobby-container max-w-[640px] mx-auto px-6 py-10">
      
      <!-- Предупреждение о небезопасном контексте (HTTP) -->
      <div v-if="isInsecureContext" class="insecure-context-banner glass p-4 mb-6 rounded-xl border border-red-500/30 flex items-start gap-3">
        <span style="font-size:24px;line-height:1;">⚠️</span>
        <div>
          <h4 style="font-weight:700;font-size:13px;color:#fca5a5;">Небезопасное окружение (HTTP)</h4>
          <p style="font-size:12px;color:var(--text-secondary);margin-top:2px;line-height:1.5;">
            Браузер блокирует функции WebRTC и захват экрана/микрофона на внешних IP-адресах без HTTPS по соображениям безопасности. 
            Пожалуйста, перейдите на <strong>localhost</strong> или настройте <strong>HTTPS</strong>. 
            Для тестирования в локальной сети вы можете включить флаг в Chrome: <code style="background:rgba(255,255,255,0.08);padding:1px 4px;border-radius:4px;word-break:break-all;">chrome://flags/#unsafely-treat-insecure-origin-as-secure</code> и добавить адрес проекта в список доверенных.
          </p>
        </div>
      </div>

      <div class="welcome-header text-center mb-8">
        <h1 class="welcome-title gradient-text inline-block">Демонстрация экрана & Чат</h1>
        <p class="welcome-desc mt-2">Создайте приватную комнату, поделитесь ссылкой с другом и общайтесь напрямую через P2P WebRTC.</p>
      </div>

      <!-- ══ СОЗДАТЬ ИЛИ ВОЙТИ ══ -->
      <div class="actions-grid flex flex-col gap-6">
        
        <!-- Создать комнату -->
        <div class="glass-card p-6 rounded-2xl border border-[var(--border)]">
          <h3 class="card-heading flex items-center gap-2 mb-4">
            <span class="icon">✨</span>
            <span>Создать новую комнату</span>
          </h3>
          <p class="card-text mb-5">Система сгенерирует уникальный идентификатор сессии. Вы сможете мгновенно поделиться ссылкой с другом.</p>
          <div class="flex gap-3">
            <input 
              v-model="newRoomId" 
              type="text" 
              class="form-input flex-1" 
              placeholder="Имя/ID комнаты..." 
              style="padding: 10px 14px; font-size: 14px;"
            />
            <button class="btn btn-primary px-6" @click="createAndJoin">
              Создать и войти
            </button>
          </div>
        </div>

        <!-- Войти в существующую -->
        <div class="glass-card p-6 rounded-2xl border border-[var(--border)]">
          <h3 class="card-heading flex items-center gap-2 mb-4">
            <span class="icon">🔑</span>
            <span>Войти в существующую комнату</span>
          </h3>
          <div class="flex gap-3">
            <input 
              v-model="targetRoomId" 
              type="text" 
              class="form-input flex-1" 
              placeholder="Введите код комнаты..." 
              style="padding: 10px 14px; font-size: 14px;"
              @keyup.enter="joinTargetRoom"
            />
            <button class="btn btn-ghost px-6 border border-[var(--border)]" :disabled="!targetRoomId.trim()" @click="joinTargetRoom">
              Войти
            </button>
          </div>
        </div>

        <!-- Пригласить друзей -->
        <div class="glass-card p-6 rounded-2xl border border-[var(--border)]">
          <h3 class="card-heading flex items-center gap-2 mb-4">
            <span class="icon">👥</span>
            <span>Позвать друга</span>
          </h3>
          
          <!-- Поиск -->
          <div class="search-box mb-4">
            <input 
              v-model="searchQuery" 
              type="text" 
              class="form-input w-full" 
              placeholder="🔍 Имя друга для поиска..." 
              style="padding: 8px 12px; font-size: 13px;"
              @input="searchUsers"
            />
          </div>

          <!-- Списки -->
          <div class="friends-list max-h-[200px] overflow-y-auto pr-1 flex flex-col gap-2">
            <!-- Результаты поиска -->
            <div v-if="searchQuery.trim()" class="search-results-section">
              <div class="list-label mb-2">Найдено пользователей:</div>
              <div v-if="searchResults.length === 0" class="no-items">Пользователи не найдены</div>
              <div v-for="user in searchResults" :key="user.id" class="friend-row flex justify-between items-center p-2 rounded-lg">
                <span class="friend-name">{{ user.username }} <span style="font-size:11px;color:var(--text-muted);">#{{ user.id }}</span></span>
                <button class="btn btn-xs btn-primary" @click="copyInviteForUser(user)">
                  Копировать ссылку
                </button>
              </div>
            </div>

            <!-- Список друзей -->
            <div v-else class="friends-section">
              <div class="list-label mb-2">Мои друзья:</div>
              <div v-if="friends.length === 0" class="no-items">У вас пока нет друзей в списке</div>
              <div v-for="friend in friends" :key="friend.id" class="friend-row flex justify-between items-center p-2 rounded-lg">
                <span class="friend-name">🟢 {{ friend.username }}</span>
                <button class="btn btn-xs btn-primary" @click="copyInviteForUser(friend)">
                  Копировать ссылку
                </button>
              </div>
            </div>
          </div>
          <div v-if="copySuccess" class="success-alert mt-3">
            ✓ Ссылка скопирована! Отправьте её другу в чат.
          </div>
        </div>

        <!-- О подключении (Справка) -->
        <div class="help-box p-4 rounded-xl">
          <h4 style="font-weight:700;font-size:12.5px;color:var(--text-primary);margin-bottom:6px;">ℹ Как работает подключение:</h4>
          <ul style="font-size:12px;color:var(--text-secondary);list-style:disc;padding-left:16px;line-height:1.6;">
            <li><strong>STUN</strong>: сервер определяет ваш публичный IP для прямого P2P соединения.</li>
            <li>Если сеть блокирует P2P (строгий NAT, VPN), требуется TURN-сервер (ретранслятор).</li>
            <li>Все медиаданные (картинка, звук) и текст идут <strong>напрямую между браузерами</strong>. Сервер не имеет доступа к вашему трафику.</li>
            <li>Для 4K качества необходимы: монитор 4K, выбор "Весь экран" при трансляции и браузер Chrome/Edge.</li>
          </ul>
        </div>

      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { friendsApi } from '@/services/api'

const router = useRouter()
const authStore = useAuthStore()

const isInsecureContext = ref(!window.isSecureContext)
const newRoomId = ref('')
const targetRoomId = ref('')
const searchQuery = ref('')

const friends = ref([])
const searchResults = ref([])
const copySuccess = ref(false)

onMounted(async () => {
  // Generate random room ID UUID candidate
  newRoomId.value = generateUUID().substring(0, 8)

  try {
    const data = await friendsApi.getFriends(authStore.accessToken)
    friends.value = data || []
  } catch (err) {
    console.error('Failed to load friends:', err)
  }
})

function generateUUID() {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }
  return Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15)
}

function createAndJoin() {
  const roomId = newRoomId.value.trim()
  if (!roomId) return
  router.push(`/screenshare/room/${roomId}`)
}

function joinTargetRoom() {
  const roomId = targetRoomId.value.trim()
  if (!roomId) return
  router.push(`/screenshare/room/${roomId}`)
}

async function searchUsers() {
  const q = searchQuery.value.trim()
  if (!q) {
    searchResults.value = []
    return
  }
  try {
    // Используем REST хендлер поиска
    const res = await fetch(`/api/v1/users/search?q=${encodeURIComponent(q)}`, {
      headers: {
        'Authorization': `Bearer ${authStore.accessToken}`
      }
    })
    if (res.ok) {
      searchResults.value = await res.json()
    }
  } catch (err) {
    console.error(err)
  }
}

function copyInviteForUser(user) {
  const roomId = newRoomId.value.trim()
  if (!roomId) return

  const inviteUrl = `${window.location.origin}/screenshare/room/${roomId}`
  navigator.clipboard.writeText(inviteUrl).then(() => {
    copySuccess.value = true
    setTimeout(() => {
      copySuccess.value = false
    }, 3000)
  })
}
</script>

<style scoped>
.lobby-page {
  min-height: 100vh;
  background: var(--bg-base);
  color: var(--text-primary);
}

/* Навбар */
.lobby-nav {
  border-radius: 0;
  border: none;
  border-bottom: 1px solid var(--border);
}
.lobby-nav-inner {
  padding: 11px 24px;
}
.nav-sep {
  width: 1px;
  height: 20px;
  background: var(--border);
}

/* Заголовки */
.welcome-title {
  font-size: clamp(24px, 3.5vw, 36px);
  font-weight: 800;
  letter-spacing: -0.02em;
}
.welcome-desc {
  font-size: 14.5px;
  color: var(--text-secondary);
  line-height: 1.5;
}

/* Карточки */
.glass-card {
  background: rgba(255, 255, 255, 0.02);
  backdrop-filter: blur(10px);
  box-shadow: 0 8px 32px 0 rgba(0, 0, 0, 0.2);
}
.card-heading {
  font-size: 15px;
  font-weight: 700;
}
.card-text {
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.5;
}

/* Списки */
.list-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.no-items {
  font-size: 12px;
  color: var(--text-muted);
  text-align: center;
  padding: 12px 0;
  font-style: italic;
}
.friend-row {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--border);
}
.friend-name {
  font-size: 13px;
  font-weight: 500;
}

/* Справка */
.help-box {
  background: rgba(99, 102, 241, 0.06);
  border: 1px solid rgba(99, 102, 241, 0.15);
}

.success-alert {
  font-size: 12.5px;
  color: #4ade80;
  background: rgba(74, 222, 128, 0.1);
  padding: 8px 12px;
  border-radius: 8px;
  border: 1px solid rgba(74, 222, 128, 0.2);
}
</style>
