<template>
  <div class="profile-page">
    <!-- Навбар -->
    <nav class="navbar glass">
      <div class="navbar-inner">
        <div class="flex items-center gap-3">
          <RouterLink to="/dashboard" class="btn btn-ghost" style="padding:7px 12px;font-size:13px;gap:6px;">
            ← Назад
          </RouterLink>
          <div class="nav-sep"></div>
          <div class="logo-mark" style="width:32px;height:32px;font-size:14px;">⬡</div>
          <span style="font-weight:700;font-size:15px;">Профиль</span>
        </div>
        <button @click="auth.logout()" class="btn btn-ghost" style="padding:7px 14px;font-size:13px;">
          Выйти
        </button>
      </div>
    </nav>

    <main class="profile-main">
      <div class="profile-header">
        <div class="profile-avatar">{{ userInitial }}</div>
        <div>
          <div class="profile-name">{{ auth.user?.username || auth.user?.email }}</div>
          <div style="font-size:13px;color:var(--text-secondary);">{{ auth.user?.email }}</div>
          <!-- Мой ID -->
          <div class="profile-id-badge glass" v-if="auth.user?.id">
            <span>ID: <code>{{ auth.user?.id }}</code></span>
            <button class="copy-id-btn" @click="copyUserID" title="Скопировать ID">
              {{ copied ? 'Скопировано!' : '📋 Копировать' }}
            </button>
          </div>
        </div>
      </div>

      <!-- ── Секция 1: Основная информация ── -->
      <section class="profile-card glass">
        <div class="card-header">
          <h2 class="card-title">Основная информация</h2>
          <p class="card-desc">Изменение логина и адреса электронной почты</p>
        </div>

        <div v-if="profileMsg.text" :class="['alert-' + profileMsg.type]">
          <span>{{ profileMsg.type === 'success' ? '✓' : '⚠' }}</span>
          {{ profileMsg.text }}
        </div>

        <form @submit.prevent="handleProfileUpdate" class="card-form">
          <div class="form-group">
            <label class="form-label" for="p-username">Логин</label>
            <input
              id="p-username"
              v-model="profileForm.username"
              type="text"
              class="form-input"
              placeholder="your_username"
            />
          </div>
          <div class="form-group">
            <label class="form-label" for="p-email">Email</label>
            <input
              id="p-email"
              v-model="profileForm.email"
              type="email"
              class="form-input"
              placeholder="you@example.com"
            />
          </div>
          <div class="form-actions">
            <button type="submit" class="btn btn-primary" :disabled="profileLoading">
              <span v-if="profileLoading" class="spinner"></span>
              <span>{{ profileLoading ? 'Сохраняем...' : 'Сохранить изменения' }}</span>
            </button>
          </div>
        </form>
      </section>

      <!-- ── Секция 2: Смена пароля ── -->
      <section class="profile-card glass">
        <div class="card-header">
          <h2 class="card-title">Смена пароля</h2>
          <p class="card-desc">Введите текущий пароль для подтверждения</p>
        </div>

        <div v-if="passwordMsg.text" :class="['alert-' + passwordMsg.type]">
          <span>{{ passwordMsg.type === 'success' ? '✓' : '⚠' }}</span>
          {{ passwordMsg.text }}
        </div>

        <form @submit.prevent="handlePasswordChange" class="card-form">
          <div class="form-group">
            <label class="form-label" for="p-current">Текущий пароль</label>
            <input
              id="p-current"
              v-model="passwordForm.current"
              type="password"
              class="form-input"
              :class="{ error: passwordErrors.current }"
              placeholder="••••••••"
              autocomplete="current-password"
            />
            <span v-if="passwordErrors.current" class="field-error">{{ passwordErrors.current }}</span>
          </div>
          <div class="form-group">
            <label class="form-label" for="p-new">Новый пароль</label>
            <input
              id="p-new"
              v-model="passwordForm.newPass"
              type="password"
              class="form-input"
              :class="{ error: passwordErrors.newPass }"
              placeholder="Минимум 8 символов"
              autocomplete="new-password"
            />
            <span v-if="passwordErrors.newPass" class="field-error">{{ passwordErrors.newPass }}</span>
          </div>
          <div class="form-group">
            <label class="form-label" for="p-confirm">Подтверждение нового пароля</label>
            <input
              id="p-confirm"
              v-model="passwordForm.confirm"
              type="password"
              class="form-input"
              :class="{ error: passwordErrors.confirm }"
              placeholder="Повторите новый пароль"
              autocomplete="new-password"
            />
            <span v-if="passwordErrors.confirm" class="field-error">{{ passwordErrors.confirm }}</span>
          </div>
          <div class="form-actions">
            <button type="submit" class="btn btn-primary" :disabled="passwordLoading">
              <span v-if="passwordLoading" class="spinner"></span>
              <span>{{ passwordLoading ? 'Меняем...' : 'Изменить пароль' }}</span>
            </button>
          </div>
        </form>
      </section>

      <!-- ── Секция: Друзья ── -->
      <section class="profile-card glass">
        <div class="card-header flex justify-between items-center" style="border-bottom:none; padding-bottom:0; flex-wrap: wrap; gap: 12px;">
          <div>
            <h2 class="card-title">Друзья</h2>
            <p class="card-desc">Поиск пользователей и управление списком друзей</p>
          </div>
          <!-- Вкладки "Друзья" / "Заявки" -->
          <div class="friends-tabs flex gap-1">
            <button
              class="tab-btn"
              :class="{ active: activeFriendsTab === 'list' }"
              @click="activeFriendsTab = 'list'"
            >
              Друзья ({{ friends.length }})
            </button>
            <button
              class="tab-btn"
              :class="{ active: activeFriendsTab === 'requests' }"
              @click="activeFriendsTab = 'requests'"
            >
              Заявки
              <span v-if="requests.length > 0" class="req-badge">{{ requests.length }}</span>
            </button>
          </div>
        </div>

        <hr style="border:none; border-top:1px solid var(--border); margin:4px 0 16px 0;" />

        <!-- ── Строка поиска пользователей ── -->
        <div class="friends-search-container flex gap-2">
          <input
            v-model="searchQuery"
            type="text"
            class="form-input"
            placeholder="Введите логин или ID..."
            @input="handleSearch"
          />
          <button v-if="searchQuery" class="btn btn-ghost" @click="clearSearch" style="padding: 10px 14px;">✕</button>
        </div>

        <!-- Результаты поиска пользователей -->
        <div v-if="searchResults.length > 0" class="search-results-list">
          <div v-for="user in searchResults" :key="user.id" class="search-result-item flex items-center justify-between">
            <div class="flex items-center gap-3">
              <div class="user-avatar-sm">{{ user.username.charAt(0).toUpperCase() }}</div>
              <div>
                <div class="user-name-text">{{ user.username }}</div>
                <div style="font-size:11px;color:var(--text-muted)">ID: {{ user.id }}</div>
              </div>
            </div>
            <button class="btn btn-outline" style="padding: 6px 12px; font-size:12px;" @click="sendFriendRequest(user)">
              Добавить
            </button>
          </div>
        </div>
        <div v-else-if="searchQuery.trim() && !searchLoading" class="no-results-text">
          Ничего не найдено
        </div>
        <div v-else-if="searchLoading" class="flex justify-center p-2">
          <div class="spinner"></div>
        </div>

        <!-- Оповещения по друзьям -->
        <div v-if="friendsMsg.text" :class="['alert-' + friendsMsg.type]" style="margin-top: 10px;">
          {{ friendsMsg.text }}
        </div>

        <!-- ── Вкладка 1: Список друзей ── -->
        <div v-if="activeFriendsTab === 'list'" class="friends-tab-content mt-2">
          <div v-if="friends.length === 0" class="empty-tab-text">
            Список друзей пуст
          </div>
          <div v-else class="friends-list">
            <div v-for="friend in friends" :key="friend.id" class="friend-item flex items-center justify-between">
              <div class="flex items-center gap-3">
                <div class="user-avatar-sm">{{ friend.username.charAt(0).toUpperCase() }}</div>
                <div>
                  <div class="user-name-text">{{ friend.username }}</div>
                  <div style="font-size:11px;color:var(--text-muted)">{{ friend.email }}</div>
                </div>
              </div>
              <button
                class="btn btn-ghost"
                style="padding: 6px 12px; font-size:12px; color:#f87171; border-color:rgba(248,113,113,0.15)"
                @click="deleteFriend(friend)"
              >
                Удалить
              </button>
            </div>
          </div>
        </div>

        <!-- ── Вкладка 2: Входящие заявки ── -->
        <div v-if="activeFriendsTab === 'requests'" class="friends-tab-content mt-2">
          <div v-if="requests.length === 0" class="empty-tab-text">
            Нет входящих заявок
          </div>
          <div v-else class="requests-list">
            <div v-for="req in requests" :key="req.id" class="request-item flex items-center justify-between">
              <div class="flex items-center gap-3">
                <div class="user-avatar-sm">{{ req.username.charAt(0).toUpperCase() }}</div>
                <div class="user-name-text">{{ req.username }}</div>
              </div>
              <div class="flex gap-2">
                <button
                  class="btn btn-primary"
                  style="padding: 6px 12px; font-size:12px; background:linear-gradient(135deg, #10b981, #059669); box-shadow:none;"
                  @click="acceptRequest(req)"
                >
                  Принять
                </button>
                <button
                  class="btn btn-ghost"
                  style="padding: 6px 12px; font-size:12px;"
                  @click="declineRequest(req)"
                >
                  Отклонить
                </button>
              </div>
            </div>
          </div>
        </div>
      </section>

      <!-- ── Секция 3: Безопасность ── -->
      <section class="profile-card glass">
        <div class="card-header">
          <h2 class="card-title">Безопасность</h2>
          <p class="card-desc">Управление сессиями и аккаунтом</p>
        </div>
        <div class="security-info">
          <div class="security-item">
            <span class="security-icon">🔑</span>
            <div>
              <div style="font-size:14px;font-weight:500;">Сессии</div>
              <div style="font-size:12px;color:var(--text-muted);">Refresh-токены хранятся в Redis. При выходе — сессия удаляется.</div>
            </div>
          </div>
          <div class="security-item">
            <span class="security-icon">🍪</span>
            <div>
              <div style="font-size:14px;font-weight:500;">HttpOnly Cookie</div>
              <div style="font-size:12px;color:var(--text-muted);">Refresh-токен недоступен JS — защита от XSS.</div>
            </div>
          </div>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup>
import { ref, computed, reactive, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { friendsApi } from '@/services/api'

const auth = useAuthStore()

const userInitial = computed(() =>
  (auth.user?.username || auth.user?.email || '?').charAt(0).toUpperCase()
)

// ── Форма профиля ──
const profileForm = reactive({
  username: '',
  email: '',
})
const profileLoading = ref(false)
const profileMsg = reactive({ text: '', type: 'success' })

// ── Форма пароля ──
const passwordForm = reactive({ current: '', newPass: '', confirm: '' })
const passwordErrors = reactive({ current: '', newPass: '', confirm: '' })
const passwordLoading = ref(false)
const passwordMsg = reactive({ text: '', type: 'success' })

// ── Состояние друзей ──
const activeFriendsTab = ref('list')
const searchQuery = ref('')
const searchLoading = ref(false)
const searchResults = ref([])
const friends = ref([])
const requests = ref([])
const friendsMsg = reactive({ text: '', type: 'success' })
const copied = ref(false)

onMounted(async () => {
  // Заполняем форму текущими данными пользователя
  profileForm.username = auth.user?.username || ''
  profileForm.email = auth.user?.email || ''
  await loadFriendsData()
})

async function handleProfileUpdate() {
  profileMsg.text = ''
  if (!profileForm.username || !profileForm.email) return

  profileLoading.value = true
  const result = await auth.updateProfile(profileForm.username, profileForm.email)
  profileLoading.value = false

  if (result.success) {
    profileMsg.text = 'Данные успешно обновлены'
    profileMsg.type = 'success'
  } else {
    const status = result.error?.status
    profileMsg.type = 'error'
    profileMsg.text = status === 409
      ? 'Этот email или логин уже занят'
      : 'Ошибка обновления. Попробуйте позже.'
  }

  // Скрыть сообщение через 4 секунды
  setTimeout(() => { profileMsg.text = '' }, 4000)
}

function validatePassword() {
  passwordErrors.current = ''
  passwordErrors.newPass = ''
  passwordErrors.confirm = ''
  let ok = true
  if (!passwordForm.current) { passwordErrors.current = 'Введите текущий пароль'; ok = false }
  if (passwordForm.newPass.length < 8) { passwordErrors.newPass = 'Минимум 8 символов'; ok = false }
  if (passwordForm.newPass !== passwordForm.confirm) { passwordErrors.confirm = 'Пароли не совпадают'; ok = false }
  return ok
}

async function handlePasswordChange() {
  passwordMsg.text = ''
  if (!validatePassword()) return

  passwordLoading.value = true
  const result = await auth.changePassword(passwordForm.current, passwordForm.newPass)
  passwordLoading.value = false

  if (result.success) {
    passwordMsg.text = 'Пароль успешно изменён'
    passwordMsg.type = 'success'
    passwordForm.current = ''
    passwordForm.newPass = ''
    passwordForm.confirm = ''
  } else {
    passwordMsg.type = 'error'
    const status = result.error?.status
    passwordMsg.text = status === 401
      ? 'Текущий пароль неверен'
      : 'Ошибка изменения пароля. Попробуйте позже.'
  }

  setTimeout(() => { passwordMsg.text = '' }, 4000)
}

async function loadFriendsData() {
  try {
    const token = auth.accessToken
    if (!token) return
    friends.value = await friendsApi.getFriends(token)
    requests.value = await friendsApi.getRequests(token)
  } catch (err) {
    console.error('Failed to load friends data', err)
  }
}

async function copyUserID() {
  if (auth.user?.id) {
    try {
      await navigator.clipboard.writeText(auth.user.id.toString())
      copied.value = true
      setTimeout(() => { copied.value = false }, 2000)
    } catch (err) {
      console.error(err)
    }
  }
}

let searchTimeout = null
function handleSearch() {
  if (searchTimeout) clearTimeout(searchTimeout)
  const q = searchQuery.value.trim()
  if (!q) {
    searchResults.value = []
    return
  }
  searchLoading.value = true
  searchTimeout = setTimeout(async () => {
    try {
      searchResults.value = await friendsApi.searchUsers(q, auth.accessToken)
    } catch (err) {
      console.error(err)
    } finally {
      searchLoading.value = false
    }
  }, 300)
}

function clearSearch() {
  searchQuery.value = ''
  searchResults.value = []
}

async function sendFriendRequest(targetUser) {
  try {
    await friendsApi.sendRequest(targetUser.id.toString(), auth.accessToken)
    showFriendsMsg(`Заявка отправлена пользователю ${targetUser.username}`, 'success')
    clearSearch()
    await loadFriendsData()
  } catch (err) {
    const msg = err.message || 'Ошибка отправки заявки'
    showFriendsMsg(msg === 'friendship already exists or pending' ? 'Заявка уже отправлена или вы уже в друзьях' : msg, 'error')
  }
}

async function acceptRequest(senderUser) {
  try {
    await friendsApi.acceptRequest(senderUser.id, auth.accessToken)
    showFriendsMsg(`Заявка от ${senderUser.username} принята`, 'success')
    await loadFriendsData()
  } catch (err) {
    showFriendsMsg('Ошибка при принятии заявки', 'error')
  }
}

async function declineRequest(senderUser) {
  try {
    await friendsApi.declineRequest(senderUser.id, auth.accessToken)
    showFriendsMsg(`Заявка от ${senderUser.username} отклонена`, 'success')
    await loadFriendsData()
  } catch (err) {
    showFriendsMsg('Ошибка при отклонении заявки', 'error')
  }
}

async function deleteFriend(friendUser) {
  if (!confirm(`Удалить ${friendUser.username} из друзей?`)) return
  try {
    await friendsApi.deleteFriend(friendUser.id, auth.accessToken)
    showFriendsMsg(`${friendUser.username} удален из друзей`, 'success')
    await loadFriendsData()
  } catch (err) {
    showFriendsMsg('Ошибка при удалении друга', 'error')
  }
}

function showFriendsMsg(text, type) {
  friendsMsg.text = text
  friendsMsg.type = type
  setTimeout(() => { friendsMsg.text = '' }, 4000)
}
</script>

<style scoped>
.profile-page {
  min-height: 100vh;
  background: var(--bg-base);
}

/* Навбар */
.navbar {
  position: sticky;
  top: 0;
  z-index: 100;
  border-bottom: 1px solid var(--border);
  border-radius: 0;
  border-left: none;
  border-right: none;
  border-top: none;
}
.navbar-inner {
  max-width: 800px;
  margin: 0 auto;
  padding: 12px 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.nav-sep {
  width: 1px;
  height: 20px;
  background: var(--border);
}

/* Основной контент */
.profile-main {
  max-width: 680px;
  margin: 0 auto;
  padding: 40px 24px 80px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

/* Шапка профиля */
.profile-header {
  display: flex;
  align-items: center;
  gap: 20px;
  padding-bottom: 8px;
}
.profile-avatar {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--primary), var(--violet));
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  font-weight: 700;
  flex-shrink: 0;
  box-shadow: 0 8px 25px var(--primary-glow);
}
.profile-name {
  font-size: 22px;
  font-weight: 700;
  letter-spacing: -0.02em;
}

/* Карточка секции */
.profile-card {
  border-radius: var(--radius-xl);
  padding: 28px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.card-header { border-bottom: 1px solid var(--border); padding-bottom: 16px; }
.card-title {
  font-size: 16px;
  font-weight: 700;
  margin-bottom: 4px;
}
.card-desc { font-size: 13px; color: var(--text-secondary); }
.card-form { display: flex; flex-direction: column; gap: 16px; }
.form-actions { display: flex; justify-content: flex-end; }
.field-error { font-size: 12px; color: #f87171; margin-top: 2px; }

/* Безопасность */
.security-info { display: flex; flex-direction: column; gap: 14px; }
.security-item {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  padding: 12px 14px;
  background: rgba(255,255,255,0.02);
  border-radius: var(--radius-md);
  border: 1px solid var(--border);
}
.security-icon { font-size: 20px; flex-shrink: 0; margin-top: 1px; }

/* Мой ID */
.profile-id-badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 4px 10px;
  border-radius: var(--radius-sm);
  font-size: 12px;
  color: var(--text-secondary);
  border: 1px solid var(--border);
  margin-top: 6px;
}
.profile-id-badge code {
  color: var(--text-primary);
  font-weight: 600;
  font-family: monospace;
}
.copy-id-btn {
  background: transparent;
  border: none;
  color: #a5b4fc;
  cursor: pointer;
  font-weight: 500;
  transition: color 0.15s;
}
.copy-id-btn:hover {
  color: white;
}

/* Друзья */
.friends-tabs {
  background: rgba(255,255,255,0.02);
  border: 1px solid var(--border);
  padding: 3px;
  border-radius: var(--radius-sm);
}
.tab-btn {
  background: transparent;
  border: none;
  color: var(--text-secondary);
  padding: 5px 12px;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  border-radius: 6px;
  transition: all 0.15s;
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.tab-btn.active {
  background: rgba(255,255,255,0.06);
  color: var(--text-primary);
  border: 1px solid var(--border);
}
.req-badge {
  background: #ef4444;
  color: white;
  font-size: 10px;
  font-weight: bold;
  padding: 1px 6px;
  border-radius: 10px;
}
.search-results-list {
  background: rgba(255,255,255,0.01);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  margin-top: 8px;
  padding: 4px;
}
.search-result-item {
  padding: 10px 14px;
  border-bottom: 1px solid rgba(255,255,255,0.03);
}
.search-result-item:last-child {
  border-bottom: none;
}
.user-avatar-sm {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--primary), var(--violet));
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 700;
  color: white;
}
.user-name-text {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}
.no-results-text, .empty-tab-text {
  font-size: 13px;
  color: var(--text-muted);
  text-align: center;
  padding: 16px;
}
.friends-list, .requests-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}
.friend-item, .request-item {
  padding: 12px 14px;
  background: rgba(255,255,255,0.02);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
}
</style>
