<template>
  <div class="profile-page">
    
    <!-- Сплит-лейаут настроек -->
    <div class="profile-settings-layout">
      
      <!-- Сайдбар настроек -->
      <AppSidebar
        :items="profileMenuItems"
        v-model="activeSection"
        v-model:expanded="isSidebarExpanded"
        :collapsible="true"
        :is-local="true"
      />

      <!-- Оверлей для мобильного сайдбара -->
      <div v-if="isSidebarExpanded" class="sidebar-overlay" @click="isSidebarExpanded = false"></div>

      <!-- Главная область контента -->
      <main class="profile-content-pane">
        
        <!-- Кнопка открытия мобильного меню настроек -->
        <button class="mobile-menu-toggle-btn md:hidden" @click="isSidebarExpanded = true" title="Открыть меню настроек">
          ☰ Настройки
        </button>

        <!-- Hero баннер (общий для всех вкладок) -->
        <div class="profile-hero-card glass mb-6">
          <div class="hero-inner-flex">
            <!-- Аватар с инициалом -->
            <div class="hero-avatar-circle">
              {{ userInitial }}
            </div>
            
            <div class="hero-profile-meta">
              <h1 class="hero-username">{{ auth.user?.username || auth.user?.email }}</h1>
              <p class="hero-email">{{ auth.user?.email }}</p>
              
              <!-- Копируемый ID -->
              <div class="hero-id-badge" v-if="auth.user?.id">
                <span>ID: <code>{{ auth.user?.id }}</code></span>
                <button class="hero-copy-btn" @click="copyUserID">
                  {{ copied ? 'Скопировано!' : '📋 Копировать ID' }}
                </button>
              </div>
            </div>

            <!-- Метрики-плашки -->
            <div class="hero-metrics-row">
              <div class="metric-pill">
                <span class="metric-value">{{ friends.length }}</span>
                <span class="metric-label">друзей</span>
              </div>
              <div class="metric-pill">
                <span class="metric-value">1</span>
                <span class="metric-label">сессия</span>
              </div>
              <div class="metric-pill">
                <span class="metric-value">{{ invites.length }}</span>
                <span class="metric-label">инвайтов</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Контент выбранной вкладки (с легким fade) -->
        <Transition name="fade" mode="out-in">
          <div :key="activeSection">
            
            <!-- Раздел: Профиль (Основная информация) -->
            <section v-if="activeSection === 'profile'" class="profile-card glass">
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

            <!-- Раздел: Безопасность (Смена пароля + Активные сессии) -->
            <div v-else-if="activeSection === 'security'" class="flex flex-col gap-6">
              
              <!-- Смена пароля -->
              <section class="profile-card glass">
                <div class="card-header">
                  <h2 class="card-title">Смена пароля</h2>
                  <p class="card-desc">Введите текущий и новый пароль для подтверждения</p>
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

              <!-- Активные сессии -->
              <section class="profile-card glass">
                <div class="card-header">
                  <h2 class="card-title">Активные сессии</h2>
                  <p class="card-desc">Контроль подключений к вашему аккаунту</p>
                </div>
                <div class="security-info">
                  <div class="security-item">
                    <span class="security-icon">🔑</span>
                    <div>
                      <div style="font-size:14px;font-weight:600;color:var(--text-primary);">Текущая сессия в Redis</div>
                      <div style="font-size:12.5px;color:var(--text-secondary);margin-top:4px;line-height:1.4;">
                        Ваши токены авторизации хранятся в защищенной базе Redis. При выходе из приложения или окончании срока действия токена сессия аннулируется автоматически.
                      </div>
                    </div>
                  </div>
                </div>
              </section>
            </div>

            <!-- Раздел: Друзья -->
            <section v-else-if="activeSection === 'friends'" class="profile-card glass">
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

              <!-- Строка поиска -->
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

              <!-- Результаты поиска -->
              <div v-if="searchResults.length > 0" class="search-results-list">
                <div v-for="user in searchResults" :key="user.id" class="search-result-item flex items-center justify-between">
                  <div class="flex items-center gap-3">
                    <div class="user-avatar-sm">{{ user.username.charAt(0).toUpperCase() }}</div>
                    <div>
                      <div class="user-name-text">{{ user.username }}</div>
                      <div style="font-size:11px;color:var(--text-muted)">ID: {{ user.id }}</div>
                    </div>
                  </div>
                  <button class="btn btn-outline text-xs py-1.5 px-3" @click="sendFriendRequest(user)">
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

              <!-- Оповещения -->
              <div v-if="friendsMsg.text" :class="['alert-' + friendsMsg.type]" style="margin-top: 10px;">
                {{ friendsMsg.text }}
              </div>

              <!-- Вкладка: Список -->
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
                        <div style="font-size:11px;color:var(--text-secondary)">{{ friend.email }}</div>
                      </div>
                    </div>
                    <button
                      class="btn btn-ghost"
                      style="padding: 6px 12px; font-size:12px; color:var(--dusty-rose); border-color:rgba(201,112,100,0.15)"
                      @click="deleteFriend(friend)"
                    >
                      Удалить
                    </button>
                  </div>
                </div>
              </div>

              <!-- Вкладка: Заявки -->
              <div v-else-if="activeFriendsTab === 'requests'" class="friends-tab-content mt-2">
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
                        style="padding: 6px 12px; font-size:12px; background:var(--matcha); box-shadow:none;"
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

            <!-- Раздел: Интеграции -->
            <section v-else-if="activeSection === 'integrations'" class="profile-card glass">
              <div class="card-header">
                <h2 class="card-title">Интеграции</h2>
                <p class="card-desc">Подключите внешние профили для синхронизации</p>
              </div>
              <div class="security-info">
                <div class="security-item flex justify-between items-center" style="width:100%">
                  <div class="flex items-center gap-3">
                    <span class="security-icon" style="background:var(--primary); color:var(--espresso); display:flex; align-items:center; justify-content:center; width:34px; height:34px; border-radius:8px; font-weight:800; font-family:var(--font-mono)">S</span>
                    <div>
                      <div style="font-size:14px;font-weight:700;color:var(--text-primary);">Shikimori</div>
                      <div style="font-size:12.5px;color:var(--text-secondary);margin-top:2px;">
                        <template v-if="auth.user?.shikimori_user_id">
                          Привязан профиль: <span class="font-bold text-emerald-400">{{ shikimoriNickname || 'Загрузка...' }}</span>
                        </template>
                        <template v-else>
                          Синхронизация списков фильмов и аниме
                        </template>
                      </div>
                    </div>
                  </div>
                  <button v-if="auth.user?.shikimori_user_id" class="btn btn-ghost" style="color:var(--dusty-rose); border-color:rgba(201,112,100,0.15)" @click="handleUnlinkShikimori">
                    Отвязать
                  </button>
                  <button v-else class="btn btn-outline" @click="linkShikimori">
                    Подключить
                  </button>
                </div>
              </div>
            </section>

            <!-- Раздел: Инвайт-коды -->
            <section v-else-if="activeSection === 'invites'" class="profile-card glass">
              <div class="card-header">
                <h2 class="card-title">Инвайт-коды</h2>
                <p class="card-desc">Создавайте пригласительные коды для ваших друзей</p>
              </div>
              
              <div style="display: flex; flex-direction: column; gap: 16px;">
                <div>
                  <button class="btn btn-primary" @click="generateInvite" :disabled="inviteLoading">
                    <span v-if="inviteLoading" class="spinner"></span>
                    <span>Создать новый инвайт-код</span>
                  </button>
                </div>

                <div v-if="invites.length === 0" class="empty-tab-text">
                  У вас пока нет инвайт-кодов
                </div>
                <div v-else class="friends-list">
                  <div v-for="inv in invites" :key="inv.id" class="friend-item flex items-center justify-between">
                    <div>
                      <div style="font-size:14px;font-family:var(--font-mono);font-weight:bold;color:var(--text-primary);">{{ inv.code }}</div>
                      <div style="font-size:11px;color:var(--text-muted);margin-top:4px;">
                        Статус: 
                        <span v-if="inv.used_by" style="color:var(--mocha); font-weight:700;">
                          Использован (ID: {{ inv.used_by }})
                        </span>
                        <span v-else style="color:var(--matcha); font-weight:700;">
                          Активен (Свободен)
                        </span>
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </section>

          </div>
        </Transition>

      </main>

    </div>
  </div>
</template>

<script setup>
import { ref, computed, reactive, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { friendsApi } from '@/services/api'
import { useSidebarStore } from '@/stores/sidebar'
import AppSidebar from '@/components/AppSidebar.vue'

const auth = useAuthStore()
const sidebarStore = useSidebarStore()

const userInitial = computed(() =>
  (auth.user?.username || auth.user?.email || '?').charAt(0).toUpperCase()
)

// ── Навигация настроек ──
const activeSection = ref('profile')
const isSidebarExpanded = ref(window.innerWidth > 768)

const profileMenuItems = computed(() => [
  { key: 'profile', title: 'Профиль', icon: '👤' },
  { key: 'security', title: 'Безопасность', icon: '🔒' },
  { key: 'friends', title: 'Друзья', icon: '👥', badge: requests.value.length > 0 ? requests.value.length : null },
  { key: 'integrations', title: 'Интеграции', icon: '🔌' },
  { key: 'invites', title: 'Инвайты', icon: '🎟️' },
])

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

// ── Инвайт-коды ──
const invites = ref([])
const inviteLoading = ref(false)

// ── Shikimori ──
const shikimoriNickname = ref('')

onMounted(async () => {
  // Ensure the global navigation sidebar is collapsed to vertical rail while on profile
  sidebarStore.collapseGlobal()

  if (auth.user) {
    profileForm.username = auth.user.username || ''
    profileForm.email = auth.user.email || ''
    
    if (auth.user.shikimori_user_id) {
      loadShikimoriProfile()
    }
  }

  await loadFriendsData()
  await loadInvites()
})

async function loadShikimoriProfile() {
  const data = await auth.fetchShikimoriWhoami()
  if (data && data.nickname) {
    shikimoriNickname.value = data.nickname
  }
}

async function handleUnlinkShikimori() {
  if (confirm('Вы уверены, что хотите отвязать аккаунт Shikimori?')) {
    const success = await auth.unlinkShikimori()
    if (success) {
      shikimoriNickname.value = ''
    }
  }
}

async function loadInvites() {
  try {
    const res = await fetch('/api/v1/auth/invites', {
      headers: { 'Authorization': `Bearer ${auth.accessToken}` }
    })
    if (res.ok) {
      invites.value = await res.json()
    }
  } catch (err) {
    console.error(err)
  }
}

async function generateInvite() {
  inviteLoading.value = true
  try {
    const res = await fetch('/api/v1/auth/invites', {
      method: 'POST',
      headers: { 'Authorization': `Bearer ${auth.accessToken}` }
    })
    if (res.ok) {
      await loadInvites()
    }
  } catch (err) {
    console.error(err)
  } finally {
    inviteLoading.value = false
  }
}

async function linkShikimori() {
  try {
    const res = await fetch('/api/v1/auth/shikimori/login', {
      headers: { 'Authorization': `Bearer ${auth.accessToken}` }
    })
    if (res.ok) {
      const data = await res.json()
      window.location.href = data.url
    } else {
      console.error('Failed to get Shikimori auth URL')
    }
  } catch (err) {
    console.error(err)
  }
}

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
  padding: 24px 24px 80px 24px;
}

.profile-settings-layout {
  display: flex;
  max-width: 1200px;
  margin: 0 auto;
  gap: 32px;
  position: relative;
  align-items: start;
}

/* ── Главная область настроек ── */
.profile-content-pane {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

/* Мобильная кнопка меню */
.mobile-menu-toggle-btn {
  align-self: flex-start;
  background: var(--bg-surface);
  border: 1px solid var(--border);
  color: var(--primary);
  font-size: 12px;
  font-weight: 700;
  padding: 6px 14px;
  border-radius: 20px;
  cursor: pointer;
  margin-bottom: 12px;
  transition: all 0.2s;
}
.mobile-menu-toggle-btn:hover {
  border-color: var(--primary);
  background: var(--btn-ghost-hover-bg);
}

/* ── Hero баннер ── */
.profile-hero-card {
  padding: 24px;
  border-radius: var(--radius-xl);
  background: var(--bg-surface);
  border: 1px solid var(--border);
}

.hero-inner-flex {
  display: flex;
  align-items: center;
  gap: 24px;
  flex-wrap: wrap;
}

.hero-avatar-circle {
  width: 72px;
  height: 72px;
  border-radius: 50%;
  background: var(--primary);
  color: var(--espresso);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32px;
  font-weight: 800;
  flex-shrink: 0;
  box-shadow: var(--shadow-md);
}
[data-theme='latte'] .hero-avatar-circle {
  color: var(--latte-foam);
}

.hero-profile-meta {
  flex: 1;
  min-width: 200px;
}

.hero-username {
  font-family: var(--font-sans);
  font-size: 22px;
  font-weight: 800;
  color: var(--text-primary);
  margin-bottom: 2px;
}

.hero-email {
  font-size: 13.5px;
  color: var(--text-muted);
}

.hero-id-badge {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  margin-top: 8px;
  background: var(--form-bg);
  border: 1px solid var(--border);
  border-radius: var(--radius-sm);
  padding: 3px 8px;
  font-size: 11px;
}
.hero-id-badge code {
  font-family: var(--font-mono);
  color: var(--text-primary);
}
.hero-copy-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  font-weight: 600;
  transition: color 0.2s;
  padding: 0 4px;
  font-size: 10px;
  border-radius: var(--radius-sm);
}
.hero-copy-btn:hover {
  color: var(--primary);
}

/* Метрики */
.hero-metrics-row {
  display: flex;
  gap: 12px;
}

.metric-pill {
  background: var(--form-bg);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  padding: 10px 16px;
  min-width: 80px;
  text-align: center;
  display: flex;
  flex-direction: column;
}

.metric-value {
  font-family: var(--font-mono);
  font-size: 16px;
  font-weight: 800;
  color: var(--primary);
}

.metric-label {
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  color: var(--text-muted);
  margin-top: 2px;
}

/* ── Настройки карточек ── */
.profile-card {
  border-radius: var(--radius-xl);
  padding: 24px;
  background: var(--bg-surface);
  border: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.card-header {
  border-bottom: 1px solid var(--border);
  padding-bottom: 14px;
}

.card-title {
  font-size: 16px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 2px;
}

.card-desc {
  font-size: 12.5px;
  color: var(--text-muted);
}

.card-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-actions {
  display: flex;
  justify-content: flex-end;
}

.field-error {
  font-size: 11px;
  color: var(--dusty-rose);
  margin-top: 2px;
}

/* Друзья */
.friends-tabs {
  background: var(--btn-ghost-bg);
  border: 1px solid var(--border);
  padding: 2px;
  border-radius: 8px;
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
}

.tab-btn.active {
  background: var(--bg-surface);
  color: var(--primary);
  box-shadow: var(--shadow-sm);
}

.req-badge {
  background: var(--dusty-rose);
  color: var(--latte-foam);
  font-size: 9px;
  font-weight: 700;
  padding: 1px 5px;
  border-radius: 10px;
}

.friends-search-container {
  margin-bottom: 8px;
}

.search-results-list {
  background: var(--form-bg);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  padding: 4px;
}

.search-result-item {
  padding: 8px 12px;
  border-bottom: 1px solid var(--border);
}
.search-result-item:last-child {
  border-bottom: none;
}

.user-avatar-sm {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  background: var(--btn-ghost-bg);
  border: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 700;
  color: var(--primary);
}

.user-name-text {
  font-size: 13.5px;
  font-weight: 700;
  color: var(--text-primary);
}

.no-results-text,
.empty-tab-text {
  font-size: 12.5px;
  color: var(--text-muted);
  text-align: center;
  padding: 16px;
}

.friends-list,
.requests-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.friend-item,
.request-item {
  padding: 10px 14px;
  background: var(--form-bg);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
}

.security-info {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.security-item {
  display: flex;
  gap: 12px;
  padding: 12px;
  background: var(--form-bg);
  border-radius: var(--radius-md);
  border: 1px solid var(--border);
}

.security-icon {
  font-size: 20px;
}

/* ── Анимация смены табов настроек ── */
.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}
.fade-enter-from {
  opacity: 0;
  transform: translateY(4px);
}
.fade-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}

/* Мобильный оверлей */
.sidebar-overlay {
  display: none;
  position: fixed;
  inset: 0;
  background: rgba(20, 16, 13, 0.6);
  backdrop-filter: blur(8px);
  z-index: 190;
}

/* Адаптив */
@media (max-width: 768px) {
  .sidebar-overlay {
    display: block;
  }

  .hero-inner-flex {
    flex-direction: column;
    align-items: flex-start;
  }
  
  .hero-metrics-row {
    width: 100%;
    margin-top: 12px;
  }

  .metric-pill {
    flex: 1;
  }

  .profile-settings-layout {
    gap: 0;
  }
}
</style>
