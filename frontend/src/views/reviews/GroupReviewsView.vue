<template>
  <div class="groups-page">

    <!-- ══ НАВБАР ══ -->
    <nav class="rv-nav glass">
      <div class="rv-nav-inner">
        <div class="flex items-center gap-3">
          <RouterLink to="/dashboard" class="btn btn-ghost" style="padding:7px 12px;font-size:13px;">← Назад</RouterLink>
          <div class="nav-sep"></div>
          <span style="font-size:18px;">⭐</span>
          <span style="font-weight:700;font-size:15px;margin-right:8px;">Рецензии</span>
          <div class="nav-sep"></div>
          <RouterLink to="/reviews" class="nav-link-toggle">Личные</RouterLink>
          <RouterLink to="/reviews/groups" class="nav-link-toggle active">Групповые</RouterLink>
        </div>

        <div class="flex items-center gap-3">
          <!-- Бейдж приглашений -->
          <button 
            class="btn btn-ghost flex items-center gap-1.5 relative"
            style="padding: 7px 12px; font-size: 13px;"
            @click="showInvitesModal = true"
          >
            <span>📩</span> Приглашения
            <span v-if="groupsStore.invites.length > 0" class="badge-invites-count">
              {{ groupsStore.invites.length }}
            </span>
          </button>

          <button v-if="activeGroup" class="add-btn" @click="openCreateItem">
            <span>＋</span> Добавить в список
          </button>
        </div>
      </div>
    </nav>

    <!-- ══ ОСНОВНОЙ ЛАЙАУТ ══ -->
    <div class="main-layout">
      
      <!-- ══ БОКОВАЯ ПАНЕЛЬ (СПИСОК ГРУПП) ══ -->
      <aside class="sidebar glass">
        <div class="sidebar-header">
          <h3>Мои Группы</h3>
          <button class="icon-btn" @click="openCreateGroup" title="Создать группу">＋</button>
        </div>

        <div class="groups-list">
          <div v-if="groupsStore.loading && groupsStore.groups.length === 0" class="sidebar-loading">
            <div class="spinner"></div>
          </div>
          <div v-else-if="groupsStore.groups.length === 0" class="sidebar-empty">
            У вас пока нет групп
          </div>
          <div 
            v-for="g in groupsStore.groups" 
            :key="g.id"
            class="group-item"
            :class="{ active: activeGroup?.id === g.id }"
            @click="selectGroup(g.id)"
          >
            <div class="group-name">👥 {{ g.name }}</div>
            <div class="group-meta">Владелец: {{ g.owner_id === authStore.user?.id ? 'Вы' : 'Другой' }}</div>
          </div>
        </div>

        <!-- Информация о выбранной группе -->
        <div v-if="activeGroup" class="active-group-sidebar-info">
          <div class="info-divider"></div>
          <div class="sidebar-subheader flex justify-between items-center">
            <h4>Участники ({{ activeGroup.members?.length || 0 }})</h4>
            <button 
              v-if="activeGroup.owner_id !== authStore.user?.id" 
              class="leave-btn"
              @click="leaveGroupConfirm"
            >
              Выйти
            </button>
            <button 
              v-else 
              class="leave-btn delete-group-btn"
              @click="deleteGroupConfirm"
            >
              Удалить
            </button>
          </div>

          <div class="members-list">
            <div v-for="m in activeGroup.members" :key="m.id" class="member-row">
              <span class="member-dot"></span>
              <span class="member-name">{{ m.username }}</span>
              <span v-if="m.user_id === activeGroup.owner_id" class="owner-badge">Корона</span>
            </div>
          </div>

          <!-- Пригласить в группу -->
          <div class="invite-section">
            <div class="info-divider"></div>
            <h4>Пригласить друга</h4>
            <div class="flex gap-2 mt-2">
              <input 
                v-model="inviteIdentifier" 
                type="text" 
                class="form-input sidebar-input" 
                placeholder="ID или username..."
                @keyup.enter="sendGroupInvite"
              />
              <button class="btn btn-primary" style="padding: 0 12px; font-size:13px;" @click="sendGroupInvite">＋</button>
            </div>
            <div v-if="inviteError" class="error-msg mt-1">{{ inviteError }}</div>
            <div v-if="inviteSuccess" class="success-msg mt-1">Приглашение отправлено!</div>
          </div>
        </div>
      </aside>

      <!-- ══ ГЛАВНЫЙ КОНТЕНТ (СПИСОК WATCHLIST) ══ -->
      <main class="content-area">
        <div v-if="!activeGroup" class="no-active-group">
          <div class="empty-bubble glass">
            <div class="bubble-icon">👥</div>
            <h3>Совместные списки</h3>
            <p>Выберите группу из списка слева или создайте новую, чтобы отслеживать просмотр фильмов и аниме вместе с друзьями в реальном времени!</p>
            <button class="btn btn-primary mt-4" style="padding: 10px 24px;" @click="openCreateGroup">Создать группу</button>
          </div>
        </div>

        <div v-else class="watchlist-area">
          <div class="watchlist-header flex justify-between items-center flex-wrap gap-4">
            <h2>🍿 {{ activeGroup.name }}</h2>
            <div class="flex items-center gap-3">
              <span class="online-indicator flex items-center gap-1.5 text-xs text-green-400">
                <span class="ping-dot"></span> в сети (Live)
              </span>
            </div>
          </div>

          <!-- ══ ТАБЫ СТАТУСОВ ══ -->
          <div class="status-tabs">
            <button
              v-for="tab in tabs"
              :key="tab.key"
              class="status-tab"
              :class="{ active: activeTab === tab.key }"
              @click="activeTab = tab.key"
            >
              <span class="tab-icon">{{ tab.icon }}</span>
              <span class="tab-label">{{ tab.label }}</span>
              <span class="tab-count" :style="{ background: tab.countBg }">
                {{ getTabCount(tab.key) }}
              </span>
            </button>
          </div>

          <!-- ══ ПАНЕЛЬ ПОИСКА И ФИЛЬТРОВ ══ -->
          <div class="filter-panel glass">
            <div class="filter-row flex flex-wrap gap-4 items-center">
              <div class="filter-search flex-1 min-w-[200px]">
                <input
                  v-model="searchQuery"
                  type="text"
                  class="form-input"
                  placeholder="Поиск по названию..."
                  style="padding: 8px 12px; font-size: 13px;"
                />
              </div>
              <div class="filter-types flex items-center gap-4">
                <label class="radio-label">
                  <input type="radio" v-model="selectedContentType" value="all" />
                  <span>Все</span>
                </label>
                <label class="radio-label" v-for="t in contentTypes" :key="t.value">
                  <input type="radio" v-model="selectedContentType" :value="t.value" />
                  <span>{{ t.icon }} {{ t.label }}</span>
                </label>
              </div>
              <button
                v-if="searchQuery || selectedContentType !== 'all'"
                @click="resetFilters"
                class="btn btn-ghost reset-filters-btn"
                style="padding: 7px 12px; font-size: 12px;"
              >
                Сбросить
              </button>
            </div>
          </div>

          <!-- ══ СПИСОК КАРТОЧЕК ══ -->
          <div class="rv-body">
            <div v-if="filteredItems.length === 0" class="empty-state">
              <div style="font-size:48px;margin-bottom:16px;">🔍</div>
              <h3 style="font-weight:700;font-size:16px;">Ничего не найдено</h3>
              <p style="font-size:13px;color:var(--text-muted);">Попробуйте изменить параметры поиска или добавить новые фильмы</p>
            </div>

            <div v-else class="card-grid">
              <div 
                v-for="item in filteredItems" 
                :key="item.id"
                class="rv-card"
              >
                <!-- Постер -->
                <div class="card-poster" :style="posterStyle(item)">
                  <div class="card-poster-overlay"></div>
                  <!-- Тип контента -->
                  <div class="card-type-badge" :style="{ background: typeColor(item.content_type) }">
                    {{ typeIcon(item.content_type) }} {{ typeLabel(item.content_type) }}
                  </div>
                  <!-- Средняя оценка -->
                  <div class="card-rating-holder flex flex-col items-end gap-1" style="align-self: flex-end; position:relative; z-index: 1;">
                    <div class="card-rating tooltip-ratings" v-if="item.average_rating > 0">
                      <span class="rating-star">★</span>
                      <span class="rating-num">{{ item.average_rating.toFixed(1) }}</span>
                      <!-- Тултип с голосами -->
                      <span class="tooltip-text glass">
                        <div style="font-weight:700; margin-bottom:4px; font-size:11px;">Оценки участников:</div>
                        <div v-for="r in item.ratings" :key="r.id" style="font-size:11px;">
                          {{ r.username }}: <span style="color:#fbbf24; font-weight:bold;">{{ r.rating }}★</span>
                        </div>
                      </span>
                    </div>
                  </div>
                </div>

                <!-- Тело карточки -->
                <div class="card-info">
                  <div class="card-title-row flex justify-between items-start gap-2">
                    <div class="card-title">{{ item.title }}</div>
                    <div class="edit-actions" v-if="item.added_by === authStore.user?.id">
                      <button class="small-icon-btn" @click="openEditItem(item)">✏️</button>
                      <button class="small-icon-btn text-red-400" @click="deleteItemConfirm(item)">🗑</button>
                    </div>
                  </div>
                  
                  <div class="author-label">Добавил: {{ item.added_by_username }}</div>

                  <!-- Прогресс серий -->
                  <div class="card-episode-progress mt-1.5 flex items-center gap-1.5 text-xs text-indigo-300 font-semibold">
                    <span>🎬</span> Серия: {{ item.current_episode }} из {{ item.max_episodes }}
                  </div>

                  <!-- Жанры -->
                  <div class="card-genres flex flex-wrap gap-1 mb-2 mt-1" v-if="item.genres && item.genres.length > 0">
                    <span v-for="g in item.genres" :key="g" class="card-genre-pill">
                      {{ g }}
                    </span>
                  </div>

                  <!-- Заметки -->
                  <div class="card-notes mt-2" v-if="item.notes">{{ item.notes }}</div>

                  <!-- Выставление оценки -->
                  <div class="item-personal-vote mt-3">
                    <div class="flex justify-between items-center mb-1">
                      <span style="font-size:11px;color:var(--text-secondary);">Моя оценка:</span>
                      <button 
                        v-if="getMyRating(item) !== null" 
                        class="btn-clear-rating"
                        @click="rateItem(item.id, null)"
                      >
                        Сбросить
                      </button>
                    </div>
                    <div class="rating-stars-row flex gap-1">
                      <button 
                        v-for="star in 10" 
                        :key="star"
                        class="star-pill-btn"
                        :class="{ active: getMyRating(item) >= star }"
                        @click="rateItem(item.id, star)"
                      >
                        {{ star }}
                      </button>
                    </div>
                  </div>

                  <!-- Ссылки -->
                  <div class="item-links-section mt-3">
                    <div class="flex justify-between items-center mb-1">
                      <span style="font-size:11px;color:var(--text-secondary);">Ссылки:</span>
                      <button class="add-link-small-btn" @click="toggleAddLinkInput(item.id)">＋ Добавить</button>
                    </div>

                    <!-- Форма добавления ссылки -->
                    <div v-if="showLinkInputId === item.id" class="add-link-inline mt-1.5 mb-2 flex gap-1">
                      <input v-model="linkForm.label" type="text" placeholder="Метка..." class="form-input text-xs" style="padding:4px 6px;" />
                      <input v-model="linkForm.url" type="url" placeholder="Ссылка..." class="form-input text-xs flex-1" style="padding:4px 6px;" />
                      <button class="btn btn-primary btn-xs" @click="addLink(item.id)">✓</button>
                      <button class="btn btn-ghost btn-xs" @click="showLinkInputId = null">✕</button>
                    </div>

                    <!-- Список ссылок -->
                    <div class="card-links flex flex-col gap-1.5" v-if="item.links && item.links.length > 0">
                      <div v-for="link in item.links" :key="link.id" class="group-link-row flex justify-between items-center">
                        <a 
                          :href="link.url" 
                          target="_blank" 
                          rel="noopener"
                          class="link-pill flex-1 text-left"
                          style="font-size:11px;"
                        >
                          🔗 {{ link.label || 'Ссылка' }} <span style="font-size:9px;color:var(--text-muted);">({{ link.username }})</span>
                        </a>
                        <button 
                          v-if="link.user_id === authStore.user?.id"
                          class="link-del-small" 
                          @click="deleteLink(item.id, link.id)"
                        >
                          ✕
                        </button>
                      </div>
                    </div>
                    <div v-else style="font-size:11px;color:var(--text-muted);font-style:italic;">Ссылок нет</div>
                  </div>

                </div>
              </div>
            </div>
          </div>
        </div>
      </main>
    </div>

    <!-- ══ МОДАЛЬНОЕ ОКНО: Входящие приглашения ══ -->
    <div v-if="showInvitesModal" class="modal-overlay" @click.self="showInvitesModal = false">
      <div class="modal-box glass" style="max-width: 440px;">
        <div class="modal-header">
          <h3 class="modal-title">📨 Приглашения в группы</h3>
          <button class="modal-close" @click="showInvitesModal = false">✕</button>
        </div>
        <div class="modal-body" style="padding: 20px;">
          <div v-if="groupsStore.invites.length === 0" style="text-align:center; padding: 20px 0; color: var(--text-muted);">
            Нет активных приглашений
          </div>
          <div v-else class="invites-list flex flex-col gap-3">
            <div v-for="inv in groupsStore.invites" :key="inv.id" class="invite-card glass p-3 rounded-lg flex flex-col gap-2">
              <div>
                Группа <span style="font-weight:700; color:white;">«{{ inv.group_name }}»</span>
              </div>
              <div style="font-size:12px; color: var(--text-muted);">
                Пригласил: <span style="color:var(--text-secondary);">{{ inv.inviter_username }}</span>
              </div>
              <div class="flex gap-2 justify-end mt-1">
                <button class="btn btn-ghost btn-xs text-red-400" @click="declineInvite(inv.id)">Отклонить</button>
                <button class="btn btn-primary btn-xs" @click="acceptInvite(inv.id)">Принять</button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- ══ МОДАЛЬНОЕ ОКНО: Создание группы ══ -->
    <div v-if="showCreateGroupModal" class="modal-overlay" @click.self="showCreateGroupModal = false">
      <div class="modal-box glass" style="max-width:440px;">
        <div class="modal-header">
          <h3 class="modal-title">👥 Создание группы</h3>
          <button class="modal-close" @click="showCreateGroupModal = false">✕</button>
        </div>
        <div class="modal-body" style="padding: 20px;">
          <div class="form-group">
            <label class="form-label">Название группы *</label>
            <input v-model="groupForm.name" type="text" class="form-input" placeholder="Например: Movie Night" />
          </div>

          <div class="form-group mt-4">
            <label class="form-label">Пригласить друзей</label>
            <div class="friends-invite-checkboxes max-h-[180px] overflow-y-auto mt-2 pr-1">
              <div v-if="friends.length === 0" style="font-size:12px; color:var(--text-muted); text-align:center; padding:10px;">
                У вас нет друзей. Вы сможете пригласить других пользователей позже по ID.
              </div>
              <div v-for="friend in friends" :key="friend.id" class="friend-checkbox-row flex items-center gap-2 py-1.5">
                <input 
                  type="checkbox" 
                  :id="'friend-' + friend.id"
                  :value="friend.id"
                  v-model="groupForm.inviteIds"
                />
                <label :for="'friend-' + friend.id" style="font-size:13px; color:var(--text-primary); cursor:pointer;">
                  {{ friend.username }} <span style="font-size:11px; color:var(--text-muted);">(#{{ friend.id }})</span>
                </label>
              </div>
            </div>
          </div>

          <div class="flex gap-3 justify-end mt-6">
            <button class="btn btn-ghost" @click="showCreateGroupModal = false">Отмена</button>
            <button class="btn btn-primary" :disabled="!groupForm.name" @click="createGroup">Создать</button>
          </div>
        </div>
      </div>
    </div>

    <!-- ══ МОДАЛЬНОЕ ОКНО: Добавить / Редактировать запись в группе ══ -->
    <div v-if="showItemModal" class="modal-overlay" @click.self="closeItemModal">
      <div class="modal-box glass">
        <div class="modal-header">
          <h3 class="modal-title">{{ isEditingItem ? 'Редактировать запись' : 'Добавить запись' }}</h3>
          <button class="modal-close" @click="closeItemModal">✕</button>
        </div>
        <div class="modal-body" style="padding:20px;">
          <!-- Тип контента -->
          <div class="form-group">
            <label class="form-label">Тип</label>
            <div class="type-selector">
              <button
                v-for="t in contentTypes"
                :key="t.value"
                class="type-btn"
                :class="{ active: itemForm.content_type === t.value }"
                @click="itemForm.content_type = t.value"
              >{{ t.icon }} {{ t.label }}</button>
            </div>
          </div>

          <!-- Название -->
          <div class="form-group mt-4">
            <label class="form-label">Название *</label>
            <input v-model="itemForm.title" type="text" class="form-input" placeholder="Введите название..." />
          </div>

          <!-- Прогресс серий (Текущая / Всего) -->
          <div class="grid grid-cols-2 gap-4 mt-4">
            <div class="form-group">
              <label class="form-label">Текущая серия *</label>
              <input v-model.number="itemForm.current_episode" type="number" min="1" class="form-input" placeholder="1" />
            </div>
            <div class="form-group">
              <label class="form-label">Всего серий *</label>
              <input v-model.number="itemForm.max_episodes" type="number" min="1" class="form-input" placeholder="1" />
            </div>
          </div>

          <!-- Статус -->
          <div class="form-group mt-4">
            <label class="form-label">Статус просмотра</label>
            <select v-model="itemForm.status" class="form-input">
              <option value="watching">📺 Смотрю</option>
              <option value="completed">✅ Просмотрено</option>
              <option value="planned">📋 Запланировано</option>
              <option value="dropped">⛔ Брошено</option>
            </select>
          </div>

          <!-- Постер URL -->
          <div class="form-group mt-4">
            <label class="form-label">URL постера</label>
            <input v-model="itemForm.poster_url" type="url" class="form-input" placeholder="https://ссылка-на-картинку..." />
          </div>

          <!-- Жанры -->
          <div class="form-group mt-4">
            <label class="form-label">Жанры (через запятую)</label>
            <input v-model="itemForm.genresString" type="text" class="form-input" placeholder="Аниме, Комедия, Боевик..." />
          </div>

          <!-- Заметки -->
          <div class="form-group mt-4">
            <label class="form-label">Заметки</label>
            <textarea v-model="itemForm.notes" class="form-input" rows="3" placeholder="Ваши мысли..."></textarea>
          </div>

          <div class="flex gap-3 justify-end mt-6">
            <button class="btn btn-ghost" @click="closeItemModal">Отмена</button>
            <button class="btn btn-primary" :disabled="!itemForm.title" @click="saveItem">Сохранить</button>
          </div>
        </div>
      </div>
    </div>

    <!-- Диалоги подтверждения -->
    <div v-if="showDeleteConfirm" class="modal-overlay" @click.self="showDeleteConfirm = false">
      <div class="confirm-box glass">
        <div style="font-size:32px;margin-bottom:12px;">🗑</div>
        <h3 style="font-weight:700;margin-bottom:8px;">Удалить запись?</h3>
        <p style="font-size:13px;color:var(--text-secondary);margin-bottom:20px;">
          Запись «{{ itemToDelete?.title }}» будет удалена из группового списка.
        </p>
        <div class="flex items-center gap-3" style="justify-content:flex-end;">
          <button class="btn btn-ghost" @click="showDeleteConfirm = false">Отмена</button>
          <button class="btn btn-primary" style="background:#ef4444;" @click="confirmDeleteItem">Удалить</button>
        </div>
      </div>
    </div>

    <div v-if="showDeleteGroupConfirm" class="modal-overlay" @click.self="showDeleteGroupConfirm = false">
      <div class="confirm-box glass">
        <div style="font-size:32px;margin-bottom:12px;">⚠️</div>
        <h3 style="font-weight:700;margin-bottom:8px;">Удалить группу?</h3>
        <p style="font-size:13px;color:var(--text-secondary);margin-bottom:20px;">
          Группа «{{ activeGroup?.name }}» и весь совместный список будут безвозвратно удалены.
        </p>
        <div class="flex items-center gap-3" style="justify-content:flex-end;">
          <button class="btn btn-ghost" @click="showDeleteGroupConfirm = false">Отмена</button>
          <button class="btn btn-primary" style="background:#ef4444;" @click="confirmDeleteGroup">Удалить</button>
        </div>
      </div>
    </div>

    <div v-if="showLeaveConfirm" class="modal-overlay" @click.self="showLeaveConfirm = false">
      <div class="confirm-box glass">
        <div style="font-size:32px;margin-bottom:12px;">🚪</div>
        <h3 style="font-weight:700;margin-bottom:8px;">Выйти из группы?</h3>
        <p style="font-size:13px;color:var(--text-secondary);margin-bottom:20px;">
          Вы больше не сможете просматривать совместный список группы «{{ activeGroup?.name }}».
        </p>
        <div class="flex items-center gap-3" style="justify-content:flex-end;">
          <button class="btn btn-ghost" @click="showLeaveConfirm = false">Отмена</button>
          <button class="btn btn-primary" style="background:#ef4444;" @click="confirmLeaveGroup">Выйти</button>
        </div>
      </div>
    </div>

  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, reactive, watch } from 'vue'
import { useGroupsStore } from '@/stores/groups'
import { useAuthStore } from '@/stores/auth'
import { friendsApi } from '@/services/api'

const groupsStore = useGroupsStore()
const authStore = useAuthStore()

// ── Состояние ──
const activeGroupId = ref(null)
const activeTab = ref('watching')
const searchQuery = ref('')
const selectedContentType = ref('all')

const friends = ref([])

// Модалки
const showInvitesModal = ref(false)
const showCreateGroupModal = ref(false)
const showItemModal = ref(false)
const showDeleteConfirm = ref(false)
const showDeleteGroupConfirm = ref(false)
const showLeaveConfirm = ref(false)

// Формы
const groupForm = reactive({
  name: '',
  inviteIds: []
})

const itemForm = reactive({
  id: null,
  title: '',
  content_type: 'movie',
  status: 'planned',
  current_episode: 1,
  max_episodes: 1,
  poster_url: '',
  genresString: '',
  notes: ''
})
const isEditingItem = ref(false)
const itemToDelete = ref(null)

// Ссылки
const showLinkInputId = ref(null)
const linkForm = reactive({
  label: '',
  url: ''
})

// Приглашения в группе
const inviteIdentifier = ref('')
const inviteError = ref('')
const inviteSuccess = ref('')

const activeGroup = computed(() => groupsStore.activeGroup)

// ── Табы и типы ──
const tabs = [
  { key: 'watching',  icon: '📺', label: 'Смотрю', countBg: 'rgba(6,182,212,0.3)' },
  { key: 'completed', icon: '✅', label: 'Просмотрено', countBg: 'rgba(34,197,94,0.3)' },
  { key: 'planned',   icon: '📋', label: 'Запланировано', countBg: 'rgba(99,102,241,0.3)' },
  { key: 'dropped',   icon: '⛔', label: 'Брошено', countBg: 'rgba(239,68,68,0.3)' },
]

const contentTypes = [
  { value: 'movie',  icon: '🎬', label: 'Фильм'  },
  { value: 'anime',  icon: '✨', label: 'Аниме'  },
  { value: 'series', icon: '📺', label: 'Сериал' },
]

// ── Инициализация ──
onMounted(async () => {
  await groupsStore.fetchGroups()
  await groupsStore.fetchInvites()
  
  try {
    const data = await friendsApi.getFriends(authStore.accessToken)
    friends.value = data || []
  } catch (err) {
    console.error('Failed to load friends:', err)
  }
})

onUnmounted(() => {
  groupsStore.disconnectWS()
})

// Реактивный выбор группы
watch(activeGroupId, async (newId) => {
  if (newId) {
    inviteIdentifier.value = ''
    inviteError.value = ''
    inviteSuccess.value = ''
    showLinkInputId.value = null
    
    await groupsStore.fetchGroupDetail(newId)
    groupsStore.connectWS(newId)
  } else {
    groupsStore.disconnectWS()
  }
})

function selectGroup(id) {
  activeGroupId.value = id
}

// ── Вычисление фильтров ──

function getTabCount(status) {
  if (!activeGroup.value || !activeGroup.value.items) return 0
  return activeGroup.value.items.filter(it => it.status === status).length
}

const filteredItems = computed(() => {
  if (!activeGroup.value || !activeGroup.value.items) return []
  
  return activeGroup.value.items.filter(item => {
    // Вкладка статуса
    if (item.status !== activeTab.value) return false
    
    // Тип контента
    if (selectedContentType.value !== 'all' && item.content_type !== selectedContentType.value) return false
    
    // Поисковый запрос
    if (searchQuery.value) {
      const q = searchQuery.value.toLowerCase()
      const titleMatch = item.title.toLowerCase().includes(q)
      const notesMatch = item.notes?.toLowerCase().includes(q)
      const authorMatch = item.added_by_username?.toLowerCase().includes(q)
      if (!titleMatch && !notesMatch && !authorMatch) return false
    }
    
    return true
  })
})

function resetFilters() {
  searchQuery.value = ''
  selectedContentType.value = 'all'
}

// ── Операции над группой ──

function openCreateGroup() {
  groupForm.name = ''
  groupForm.inviteIds = []
  showCreateGroupModal.value = true
}

async function createGroup() {
  if (!groupForm.name) return
  try {
    const g = await groupsStore.createGroup(groupForm.name, groupForm.inviteIds)
    showCreateGroupModal.value = false
    selectGroup(g.id)
  } catch (err) {
    console.error('Failed to create group:', err)
  }
}

function deleteGroupConfirm() {
  showDeleteGroupConfirm.value = true
}

async function confirmDeleteGroup() {
  if (!activeGroup.value) return
  try {
    await groupsStore.deleteGroup(activeGroup.value.id)
    activeGroupId.value = null
    showDeleteGroupConfirm.value = false
  } catch (err) {
    console.error(err)
  }
}

function leaveGroupConfirm() {
  showLeaveConfirm.value = true
}

async function confirmLeaveGroup() {
  if (!activeGroup.value) return
  try {
    await groupsStore.leaveGroup(activeGroup.value.id)
    activeGroupId.value = null
    showLeaveConfirm.value = false
  } catch (err) {
    console.error(err)
  }
}

async function sendGroupInvite() {
  if (!inviteIdentifier.value || !activeGroup.value) return
  inviteError.value = ''
  inviteSuccess.value = ''
  try {
    await groupsStore.inviteUser(activeGroup.value.id, inviteIdentifier.value)
    inviteSuccess.value = true
    inviteIdentifier.value = ''
  } catch (err) {
    inviteError.value = err.message || 'Ошибка отправки приглашения'
  }
}

// ── Приглашения ──

async function acceptInvite(inviteId) {
  try {
    const joinedGroupId = await groupsStore.acceptInvite(inviteId)
    showInvitesModal.value = false
    selectGroup(joinedGroupId)
  } catch (err) {
    console.error(err)
  }
}

async function declineInvite(inviteId) {
  try {
    await groupsStore.declineInvite(inviteId)
  } catch (err) {
    console.error(err)
  }
}

// ── Записи ──

function openCreateItem() {
  isEditingItem.value = false
  itemForm.id = null
  itemForm.title = ''
  itemForm.content_type = 'movie'
  itemForm.status = activeTab.value
  itemForm.current_episode = 1
  itemForm.max_episodes = 1
  itemForm.poster_url = ''
  itemForm.genresString = ''
  itemForm.notes = ''
  showItemModal.value = true
}

function openEditItem(item) {
  isEditingItem.value = true
  itemForm.id = item.id
  itemForm.title = item.title
  itemForm.content_type = item.content_type
  itemForm.status = item.status
  itemForm.current_episode = item.current_episode || 1
  itemForm.max_episodes = item.max_episodes || 1
  itemForm.poster_url = item.poster_url
  itemForm.genresString = item.genres ? item.genres.join(', ') : ''
  itemForm.notes = item.notes
  showItemModal.value = true
}

function closeItemModal() {
  showItemModal.value = false
}

// Следим за типом контента: для фильмов всегда 1 из 1
watch(() => itemForm.content_type, (newType) => {
  if (newType === 'movie') {
    itemForm.current_episode = 1
    itemForm.max_episodes = 1
  }
})

async function saveItem() {
  if (!itemForm.title || !activeGroup.value) return
  
  const genres = itemForm.genresString
    ? itemForm.genresString.split(',').map(s => s.trim()).filter(Boolean)
    : []

  const payload = {
    title: itemForm.title,
    content_type: itemForm.content_type,
    status: itemForm.status,
    current_episode: itemForm.current_episode || 1,
    max_episodes: itemForm.max_episodes || 1,
    poster_url: itemForm.poster_url,
    genres,
    notes: itemForm.notes
  }

  try {
    if (isEditingItem.value) {
      await groupsStore.updateGroupItem(activeGroup.value.id, itemForm.id, payload)
    } else {
      await groupsStore.addGroupItem(activeGroup.value.id, payload)
    }
    closeItemModal()
  } catch (err) {
    console.error(err)
  }
}

function deleteItemConfirm(item) {
  itemToDelete.value = item
  showDeleteConfirm.value = true
}

async function confirmDeleteItem() {
  if (!itemToDelete.value || !activeGroup.value) return
  try {
    await groupsStore.deleteGroupItem(activeGroup.value.id, itemToDelete.value.id)
    showDeleteConfirm.value = false
    itemToDelete.value = null
  } catch (err) {
    console.error(err)
  }
}

// ── Оценки ──

function getMyRating(item) {
  const vote = item.ratings?.find(r => r.user_id === authStore.user?.id)
  return vote && vote.rating !== undefined ? vote.rating : null
}

async function rateItem(itemId, rating) {
  if (!activeGroup.value) return
  try {
    await groupsStore.rateGroupItem(activeGroup.value.id, itemId, rating)
  } catch (err) {
    console.error(err)
  }
}

// ── Ссылки ──

function toggleAddLinkInput(itemId) {
  if (showLinkInputId.value === itemId) {
    showLinkInputId.value = null
  } else {
    showLinkInputId.value = itemId
    linkForm.label = ''
    linkForm.url = ''
  }
}

async function addLink(itemId) {
  if (!linkForm.label || !linkForm.url || !activeGroup.value) return
  try {
    await groupsStore.addGroupItemLink(activeGroup.value.id, itemId, linkForm.label, linkForm.url)
    showLinkInputId.value = null
  } catch (err) {
    console.error(err)
  }
}

async function deleteLink(itemId, linkId) {
  if (!activeGroup.value) return
  try {
    await groupsStore.deleteGroupItemLink(activeGroup.value.id, itemId, linkId)
  } catch (err) {
    console.error(err)
  }
}

// ── Вспомогательные ──

function posterStyle(item) {
  if (item.poster_url) {
    return {
      backgroundImage: `url(${item.poster_url})`,
      backgroundSize: 'cover',
      backgroundPosition: 'center',
    }
  }
  return {
    background: 'linear-gradient(135deg, rgba(255,255,255,0.05) 0%, rgba(255,255,255,0.02) 100%)',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
  }
}

function typeLabel(type) {
  switch (type) {
    case 'movie':  return 'Фильм'
    case 'anime':  return 'Аниме'
    case 'series': return 'Сериал'
    default:       return 'Контент'
  }
}

function typeIcon(type) {
  switch (type) {
    case 'movie':  return '🎬'
    case 'anime':  return '✨'
    case 'series': return '📺'
    default:       return '📦'
  }
}

function typeColor(type) {
  switch (type) {
    case 'movie':  return 'linear-gradient(135deg, #f43f5e 0%, #be123c 100%)' // Розовый/Красный
    case 'anime':  return 'linear-gradient(135deg, #a855f7 0%, #6b21a8 100%)' // Фиолетовый
    case 'series': return 'linear-gradient(135deg, #06b6d4 0%, #0891b2 100%)' // Синий
    default:       return 'linear-gradient(135deg, #6b7280 0%, #374151 100%)'
  }
}
</script>

<style scoped>
.groups-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--bg-base);
}

/* Навбар */
.rv-nav { border-radius: 0; border: none; border-bottom: 1px solid var(--border); }
.rv-nav-inner { padding: 11px 24px; display: flex; align-items: center; justify-content: space-between; }
.nav-sep { width: 1px; height: 20px; background: var(--border); }
.add-btn {
  display: flex; align-items: center; gap: 6px;
  padding: 8px 16px; border-radius: var(--radius-md);
  background: var(--primary); border: none; color: white;
  font-size: 13px; font-weight: 600; cursor: pointer;
  transition: opacity 0.2s, transform 0.15s;
}
.add-btn:hover { opacity: 0.85; transform: scale(1.02); }

.nav-link-toggle {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-muted);
  text-decoration: none;
  padding: 6px 12px;
  border-radius: var(--radius-md);
  transition: all 0.2s;
}
.nav-link-toggle:hover {
  color: var(--text-primary);
  background: rgba(255,255,255,0.05);
}
.nav-link-toggle.active {
  color: var(--text-primary);
  background: rgba(255,255,255,0.08);
  font-weight: 600;
}

.badge-invites-count {
  background: #ef4444;
  color: white;
  font-size: 10px;
  font-weight: 700;
  border-radius: 50%;
  width: 16px;
  height: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
}

/* Лайаут */
.main-layout {
  flex: 1;
  display: flex;
  overflow: hidden;
}

/* Боковая панель */
.sidebar {
  width: 280px;
  flex-shrink: 0;
  border-right: 1px solid var(--border);
  background: rgba(255, 255, 255, 0.015);
  display: flex;
  flex-direction: column;
  padding: 16px;
}
.sidebar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}
.sidebar-header h3 {
  font-size: 14px;
  font-weight: 700;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.icon-btn {
  background: rgba(255,255,255,0.06);
  border: 1px solid var(--border);
  color: var(--text-primary);
  width: 28px;
  height: 28px;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 15px;
  transition: background 0.15s;
}
.icon-btn:hover {
  background: rgba(255,255,255,0.12);
}

.groups-list {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
  overflow-y: auto;
}
.sidebar-empty, .sidebar-loading {
  text-align: center;
  color: var(--text-muted);
  font-size: 13px;
  padding: 20px 0;
}
.group-item {
  padding: 12px 14px;
  border-radius: var(--radius-md);
  border: 1px solid transparent;
  background: rgba(255, 255, 255, 0.02);
  cursor: pointer;
  transition: all 0.2s;
}
.group-item:hover {
  background: rgba(255,255,255,0.06);
}
.group-item.active {
  background: rgba(99, 102, 241, 0.08);
  border-color: rgba(99, 102, 241, 0.3);
}
.group-name {
  font-size: 13.5px;
  font-weight: 600;
  color: var(--text-primary);
}
.group-meta {
  font-size: 11px;
  color: var(--text-muted);
  margin-top: 3px;
}

.active-group-sidebar-info {
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
}
.info-divider {
  height: 1px;
  background: var(--border);
  margin: 16px 0;
}
.sidebar-subheader h4 {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
}
.leave-btn {
  font-size: 11px;
  color: var(--text-muted);
  background: transparent;
  border: none;
  cursor: pointer;
  transition: color 0.15s;
}
.leave-btn:hover {
  color: #f87171;
}
.members-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-top: 10px;
  max-h: 120px;
  overflow-y: auto;
}
.member-row {
  display: flex;
  align-items: center;
  gap: 8px;
}
.member-dot {
  width: 6px;
  height: 6px;
  background: #10b981;
  border-radius: 50%;
}
.member-name {
  font-size: 12.5px;
  color: var(--text-secondary);
  flex-1: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
.owner-badge {
  font-size: 9px;
  background: rgba(251, 191, 36, 0.15);
  color: #fbbf24;
  padding: 1px 5px;
  border-radius: 4px;
  font-weight: 700;
}

.invite-section h4 {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-secondary);
}
.sidebar-input {
  padding: 6px 10px;
  font-size: 12px;
}
.error-msg {
  font-size: 11px;
  color: #f87171;
}
.success-msg {
  font-size: 11px;
  color: #4ade80;
}

/* Главный контент */
.content-area {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow-y: auto;
}

.no-active-group {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px;
}
.empty-bubble {
  max-width: 480px;
  padding: 32px;
  border-radius: var(--radius-xl);
  text-align: center;
  border: 1px solid var(--border);
  background: rgba(255,255,255,0.01);
}
.bubble-icon {
  font-size: 48px;
  margin-bottom: 16px;
}
.empty-bubble h3 {
  font-size: 18px;
  font-weight: 700;
  margin-bottom: 8px;
}
.empty-bubble p {
  font-size: 13.5px;
  color: var(--text-muted);
  line-height: 1.5;
}

.watchlist-area {
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.watchlist-header h2 {
  font-size: 20px;
  font-weight: 800;
}
.ping-dot {
  width: 8px;
  height: 8px;
  background: #10b981;
  border-radius: 50%;
  box-shadow: 0 0 8px #10b981;
}

/* Табы статусов */
.status-tabs {
  display: flex;
  gap: 4px;
  background: rgba(255,255,255,0.01);
  overflow-x: auto;
  scrollbar-width: none;
}
.status-tabs::-webkit-scrollbar { display: none; }
.status-tab {
  display: flex; align-items: center; gap: 7px;
  padding: 8px 16px;
  border-radius: var(--radius-md);
  background: transparent; border: 1px solid transparent;
  color: var(--text-secondary); cursor: pointer;
  font-size: 13px; font-weight: 500;
  white-space: nowrap;
  transition: all 0.2s;
}
.status-tab:hover { background: rgba(255,255,255,0.05); color: var(--text-primary); }
.status-tab.active {
  background: rgba(255,255,255,0.06);
  border-color: var(--border);
  color: var(--text-primary);
  font-weight: 600;
}
.tab-icon { font-size: 15px; }
.tab-count {
  font-size: 11px; font-weight: 700;
  padding: 2px 8px; border-radius: 20px;
  color: var(--text-primary);
  min-width: 22px; text-align: center;
}

/* Фильтры */
.filter-panel {
  padding: 12px 16px;
  border-radius: var(--radius-lg);
  border: 1px solid var(--border);
}
.radio-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: color 0.15s;
}
.radio-label:hover {
  color: var(--text-primary);
}
.radio-label input {
  accent-color: var(--primary);
}

/* Тело */
.rv-body {
  margin-top: 10px;
}
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  min-height: 240px;
  text-align: center;
  color: var(--text-muted);
}

/* Карточки */
.card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
  gap: 20px;
}
.rv-card {
  background: rgba(255,255,255,0.03);
  border: 1px solid var(--border);
  border-radius: var(--radius-xl);
  overflow: hidden;
  transition: transform 0.2s, border-color 0.2s;
}
.rv-card:hover {
  transform: translateY(-3px);
  border-color: rgba(99,102,241,0.4);
}
.card-poster {
  position: relative;
  height: 200px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  padding: 12px;
}
.card-poster-overlay {
  position: absolute; inset: 0;
  background: linear-gradient(to top, rgba(0,0,0,0.85) 0%, transparent 60%);
  pointer-events: none;
}
.card-type-badge {
  position: relative; z-index: 1;
  display: inline-flex; align-items: center; gap: 4px;
  padding: 4px 10px; border-radius: 20px;
  font-size: 10.5px; font-weight: 700; color: white;
  align-self: flex-start;
}
.card-rating {
  position: relative; z-index: 1;
  display: flex; align-items: baseline; gap: 2px;
  background: rgba(0,0,0,0.65); backdrop-filter: blur(4px);
  padding: 4px 8px; border-radius: 8px;
  cursor: help;
}
.rating-star { font-size: 13px; color: #fbbf24; }
.rating-num  { font-size: 15px; font-weight: 700; color: white; }

/* Тултипы с оценками */
.tooltip-ratings {
  position: relative;
}
.tooltip-ratings .tooltip-text {
  visibility: hidden;
  position: absolute;
  bottom: 125%;
  right: 0;
  width: 160px;
  background: rgba(0, 0, 0, 0.85);
  border: 1px solid var(--border);
  color: var(--text-primary);
  text-align: left;
  padding: 8px 12px;
  border-radius: 8px;
  z-index: 100;
  opacity: 0;
  transition: opacity 0.2s;
  box-shadow: 0 4px 16px rgba(0,0,0,0.5);
  pointer-events: none;
}
.tooltip-ratings:hover .tooltip-text {
  visibility: visible;
  opacity: 1;
}

/* Инфо */
.card-info { padding: 14px; }
.card-title-row {
  margin-bottom: 2px;
}
.card-title {
  font-size: 15px; font-weight: 700; color: var(--text-primary);
  line-height: 1.4;
}
.edit-actions {
  display: flex;
  gap: 6px;
}
.small-icon-btn {
  background: transparent;
  border: none;
  font-size: 12px;
  cursor: pointer;
  opacity: 0.6;
  transition: opacity 0.15s;
}
.small-icon-btn:hover {
  opacity: 1;
}

.author-label {
  font-size: 11px;
  color: var(--text-muted);
}
.card-genres {
  margin-top: 6px;
}
.card-genre-pill {
  font-size: 10px;
  background: rgba(255,255,255,0.05);
  color: var(--text-secondary);
  padding: 2px 7px;
  border-radius: 4px;
  border: 1px solid var(--border);
}
.card-notes {
  font-size: 12px;
  color: var(--text-secondary);
  background: rgba(255,255,255,0.02);
  border-radius: 8px;
  padding: 8px 10px;
  line-height: 1.4;
  border-left: 2px solid var(--primary);
}

/* Оценка (Personal Voting) */
.item-personal-vote {
  background: rgba(255,255,255,0.015);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  padding: 8px 10px;
}
.btn-clear-rating {
  font-size: 10px;
  color: #f87171;
  background: transparent;
  border: none;
  cursor: pointer;
}
.rating-stars-row {
  display: flex;
  justify-content: space-between;
}
.star-pill-btn {
  background: rgba(255,255,255,0.04);
  border: 1px solid var(--border);
  color: var(--text-secondary);
  width: 20px;
  height: 20px;
  font-size: 9px;
  font-weight: 700;
  border-radius: 4px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
}
.star-pill-btn:hover {
  background: rgba(251, 191, 36, 0.15);
  border-color: #fbbf24;
  color: #fbbf24;
}
.star-pill-btn.active {
  background: #fbbf24;
  border-color: #fbbf24;
  color: black;
}

/* Ссылки */
.item-links-section {
  border-top: 1px dashed var(--border);
  padding-top: 10px;
}
.add-link-small-btn {
  font-size: 10.5px;
  color: var(--primary);
  background: transparent;
  border: none;
  cursor: pointer;
}
.add-link-inline {
  display: flex;
  flex-direction: column;
  gap: 6px;
  background: rgba(255,255,255,0.02);
  padding: 8px;
  border-radius: var(--radius-md);
  border: 1px solid var(--border);
}
.group-link-row {
  display: flex;
  gap: 6px;
}
.link-pill {
  font-size: 11px;
  padding: 3px 9px;
  border-radius: 6px;
  background: rgba(255,255,255,0.04);
  border: 1px solid var(--border);
  color: var(--text-secondary);
  text-decoration: none;
}
.link-pill:hover {
  background: rgba(99,102,241,0.15);
  color: #a5b4fc;
}
.link-del-small {
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 10px;
  cursor: pointer;
}
.link-del-small:hover {
  color: #f87171;
}

/* Модалки */
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
}
.modal-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 18px 22px 14px;
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}
.modal-title { font-size: 16px; font-weight: 700; }
.modal-close {
  width: 30px; height: 30px; border-radius: 50%;
  background: rgba(255,255,255,0.06); border: none;
  color: var(--text-secondary); font-size: 14px;
  cursor: pointer; display: flex; align-items: center; justify-content: center;
}
.modal-close:hover { background: rgba(255,255,255,0.12); color: var(--text-primary); }

.modal-body {
  flex: 1; overflow-y: auto; padding: 22px;
}
.form-group { display: flex; flex-direction: column; gap: 6px; }
.form-label { font-size: 13px; font-weight: 600; color: var(--text-secondary); }
.type-selector { display: flex; gap: 6px; }
.type-btn {
  flex: 1; padding: 8px; border-radius: var(--radius-md);
  background: rgba(255,255,255,0.03); border: 1px solid var(--border);
  color: var(--text-secondary); cursor: pointer; font-size: 13px;
  transition: all 0.2s;
}
.type-btn:hover { background: rgba(255,255,255,0.08); color: var(--text-primary); }
.type-btn.active {
  background: rgba(99,102,241,0.15); border-color: var(--primary);
  color: white; font-weight: 600;
}

/* Приглашения */
.invite-card {
  border: 1px solid var(--border);
  background: rgba(255,255,255,0.02);
}

/* Диалог подтверждения */
.confirm-box {
  width: 100%; max-width: 380px; padding: 24px;
  border-radius: var(--radius-xl); text-align: center;
  border: 1px solid var(--border);
}

/* Адаптив */
@media (max-width: 768px) {
  .main-layout {
    flex-direction: column;
  }
  .sidebar {
    width: 100%;
    border-right: none;
    border-bottom: 1px solid var(--border);
  }
  .card-grid {
    grid-template-columns: 1fr;
  }
}
</style>
