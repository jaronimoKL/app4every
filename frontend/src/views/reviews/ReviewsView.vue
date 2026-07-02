<template>
  <div class="reviews-page">

    <!-- ══ НАВБАР ══ -->
    <nav class="rv-nav glass">
      <div class="rv-nav-inner">
        <div class="flex items-center gap-3">
          <RouterLink to="/dashboard" class="btn btn-ghost" style="padding:7px 12px;font-size:13px;">← Назад</RouterLink>
          <div class="nav-sep"></div>
          <span style="font-size:18px;">⭐</span>
          <span style="font-weight:700;font-size:15px;margin-right:8px;">Рецензии</span>
          <div class="nav-sep"></div>
          <RouterLink to="/reviews" class="nav-link-toggle active">Личные</RouterLink>
          <RouterLink to="/reviews/groups" class="nav-link-toggle flex items-center gap-1.5">
            Групповые
            <span v-if="groupsStore.invites.length > 0" class="badge-invites-count">
              {{ groupsStore.invites.length }}
            </span>
          </RouterLink>
        </div>
        <div class="flex items-center gap-3">
          <button v-if="auth.user?.shikimori_user_id" class="btn btn-outline btn-sync" :class="{ 'is-syncing': store.syncing }" style="padding:7px 12px;font-size:13px;" @click="syncWithShikimori">
            <span v-if="store.syncing" class="spinner" style="width:14px;height:14px;margin-right:6px;"></span>
            <span v-else style="margin-right:4px;">🔄</span>
            Синхронизировать
            <div v-if="store.syncing || store.syncProgress > 0" class="sync-progress" :style="{ width: store.syncProgress + '%' }"></div>
          </button>
          <button class="add-btn" @click="openCreate">
            <span>＋</span> Добавить
          </button>
        </div>
      </div>
    </nav>

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
          {{ store.byStatus[tab.key].length }}
        </span>
      </button>
    </div>

    <!-- ══ ПАНЕЛЬ ПОИСКА И ФИЛЬТРОВ ══ -->
    <div class="filter-panel glass">
      <div class="filter-row flex flex-wrap gap-4 items-center">
        <!-- Поиск -->
        <div class="filter-search flex-1 min-w-[200px]">
          <input
            v-model="searchQuery"
            type="text"
            class="form-input"
            placeholder="Поиск по названию..."
            style="padding: 8px 12px; font-size: 13px;"
          />
        </div>
        
        <!-- Тип контента (radio) -->
        <div class="filter-types modern-pills">
          <button class="pill-btn" :class="{active: selectedContentType === 'all'}" @click="selectedContentType = 'all'">Все</button>
          <button v-for="t in contentTypes" :key="t.value" class="pill-btn" :class="{active: selectedContentType === t.value}" @click="selectedContentType = t.value">
            {{ t.icon }} {{ t.label }}
          </button>
        </div>

        <!-- Shikimori фильтр -->
        <div class="shikimori-filter flex items-center bg-[rgba(255,255,255,0.05)] rounded-md p-1">
          <button 
            class="shikimori-btn" 
            :class="{ active: selectedShikimoriFilter === 'all' }" 
            @click="selectedShikimoriFilter = 'all'"
            title="Все рецензии"
          >
            <div class="shiki-icon shiki-half"></div>
          </button>
          <button 
            class="shikimori-btn" 
            :class="{ active: selectedShikimoriFilter === 'shikimori_only' }" 
            @click="selectedShikimoriFilter = 'shikimori_only'"
            title="Только с Shikimori"
          >
            <div class="shiki-icon shiki-full"></div>
          </button>
          <button 
            class="shikimori-btn" 
            :class="{ active: selectedShikimoriFilter === 'app_only' }" 
            @click="selectedShikimoriFilter = 'app_only'"
            title="Созданные в приложении"
          >
            <div class="shiki-icon shiki-empty"></div>
          </button>
        </div>

        <!-- Кнопка сброса -->
        <button
          v-if="isAnyFilterActive"
          @click="resetFilters"
          class="btn btn-ghost reset-filters-btn"
          style="padding: 7px 12px; font-size: 12px;"
        >
          Сбросить фильтры
        </button>
      </div>

      <!-- Фильтр по жанрам (мультиселект-теги) -->
      <div class="filter-genres-row mt-3" v-if="availableGenres.length > 0">
        <span style="font-size:12px; color:var(--text-secondary); margin-right:8px;">Жанры:</span>
        <div class="filter-genres flex flex-wrap gap-1 inline-flex">
          <button
            v-for="genreName in availableGenres"
            :key="genreName"
            class="filter-genre-btn"
            :class="{ active: selectedGenres.has(genreName) }"
            @click="toggleFilterGenre(genreName)"
          >
            {{ genreName }}
          </button>
        </div>
      </div>
    </div>

    <!-- ══ ОСНОВНОЙ КОНТЕНТ ══ -->
    <div class="rv-body">

      <!-- Загрузка -->
      <div v-if="store.loading" class="empty-state">
        <div class="spinner" style="width:32px;height:32px;"></div>
      </div>

      <!-- Пустой список -->
      <div v-else-if="currentReviews.length === 0" class="empty-state">
        <template v-if="isAnyFilterActive">
          <div style="font-size:52px;margin-bottom:16px;">🔍</div>
          <h3 style="font-weight:700;font-size:17px;margin-bottom:8px;">Ничего не найдено</h3>
          <p style="font-size:14px;color:var(--text-muted);margin-bottom:20px;">Попробуйте изменить параметры поиска или сбросить фильтры</p>
          <button class="btn btn-primary" style="padding:10px 24px;" @click="resetFilters">Сбросить фильтры</button>
        </template>
        <template v-else>
          <div style="font-size:52px;margin-bottom:16px;">{{ currentTab.icon }}</div>
          <h3 style="font-weight:700;font-size:17px;margin-bottom:8px;">{{ currentTab.emptyTitle }}</h3>
          <p style="font-size:14px;color:var(--text-muted);margin-bottom:20px;">{{ currentTab.emptyDesc }}</p>
          <button class="btn btn-primary" style="padding:10px 24px;" @click="openCreate">Добавить</button>
        </template>
      </div>

      <!-- Сетка карточек -->
      <div v-else class="card-grid">
        <div
          v-for="rev in currentReviews.slice(0, visibleLimit)"
          :key="rev.id"
          class="rv-card"
          @click="openEdit(rev)"
        >
          <!-- Постер или заглушка -->
          <div class="card-poster" :style="!rev.poster_url ? { background: posterGradients[rev.content_type] || posterGradients.movie } : {}">
            <div v-if="rev.poster_url && !loadedImages[rev.id]" class="poster-skeleton">
              <div class="spinner" style="width:24px;height:24px;border-width:2px;"></div>
            </div>
            <img 
              v-if="rev.poster_url" 
              :src="rev.poster_url" 
              class="poster-img"
              :class="{ 'img-loaded': loadedImages[rev.id] }"
              @load="loadedImages[rev.id] = true"
              @error="loadedImages[rev.id] = true"
              loading="lazy"
            />
            <div class="card-poster-overlay"></div>
            <!-- Бейдж типа -->
            <div class="card-type-badge" :style="{ background: typeColor(rev.content_type) }">
              {{ typeIcon(rev.content_type) }} {{ typeLabel(rev.content_type) }}
            </div>
            <!-- Бейдж Shikimori -->
            <div v-if="rev.shikimori_id" class="shikimori-badge">
              <span title="Синхронизировано с Shikimori">⛩️ Shikimori</span>
            </div>
            <!-- Оценка и Прогресс -->
            <div class="card-badges-top-right" v-if="rev.rating || rev.episodes_total">
              <div class="card-rating" v-if="rev.rating">
                <span class="rating-star">★</span>
                <span class="rating-num">{{ rev.rating }}</span>
                <span class="rating-max">/10</span>
              </div>
            </div>
          </div>

          <!-- Информация под постером -->
          <div class="card-info">
            <div class="card-title">{{ rev.title }}</div>
            <div class="card-episode-progress mt-1.5 flex items-center gap-1.5 text-xs text-indigo-300 font-semibold" v-if="rev.episodes_total && rev.content_type !== 'movie'">
              <span>🎬</span> Серия: {{ rev.current_episode || 0 }} из {{ rev.episodes_total }}
              <button 
                class="btn-increment-ep ml-auto"
                v-if="(rev.current_episode || 0) < rev.episodes_total" 
                @click.stop="incrementEpisode(rev)"
                title="Отметить следующую серию просмотренной"
              >
                ＋1
              </button>
            </div>
            
            <!-- Теги жанров на карточке -->
            <div class="card-genres flex flex-wrap gap-1 mb-2" v-if="rev.genres && rev.genres.length > 0">
              <span v-for="g in rev.genres" :key="g.id" class="card-genre-pill">
                {{ g.name }}
              </span>
            </div>

            <div class="card-notes" v-if="rev.notes">{{ rev.notes }}</div>
            <!-- Ссылки -->
            <div class="card-links" v-if="rev.links && rev.links.length > 0" @click.stop>
              <a
                v-for="link in rev.links"
                :key="link.id"
                :href="link.url"
                target="_blank"
                rel="noopener"
                class="link-pill"
              >
                🔗 {{ link.label || 'Ссылка' }}
              </a>
              <button
                v-if="rev.shikimori_id || rev.aniliberty_alias || (rev.links && rev.links.length > 0)"
                class="watch-together-btn"
                @click.stop="handleWatchTogether(rev)"
              >
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" style="width:13px; height:13px; margin-right:4px; display:inline-block; vertical-align:middle;">
                  <polygon points="5 3 19 12 5 21 5 3" fill="currentColor"/>
                </svg>
                Смотреть вместе
              </button>
            </div>
            <div class="card-links" v-else-if="rev.shikimori_id || rev.aniliberty_alias" @click.stop>
              <button
                class="watch-together-btn"
                @click.stop="handleWatchTogether(rev)"
              >
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" style="width:13px; height:13px; margin-right:4px; display:inline-block; vertical-align:middle;">
                  <polygon points="5 3 19 12 5 21 5 3" fill="currentColor"/>
                </svg>
                Смотреть вместе
              </button>
            </div>
          </div>
        </div>
      </div>
      <!-- Триггер для подгрузки элементов при скролле -->
      <div ref="loadMoreTrigger" class="flex justify-center py-6" style="width: 100%;">
        <div v-if="isLoadingMore" class="spinner" style="width:32px;height:32px;"></div>
      </div>
    </div>

    <!-- ══ МОДАЛЬНОЕ ОКНО: Добавить / Редактировать ══ -->
    <div v-if="showModal" class="modal-overlay" @click.self="closeModal">
      <div class="modal-box glass" @click.stop>

        <!-- Заголовок модала -->
        <div class="modal-header">
          <h2 class="modal-title">{{ isEditing ? 'Редактировать' : 'Добавить' }}</h2>
          <button class="modal-close" @click="closeModal">✕</button>
        </div>

        <div class="modal-body">
          <AnimeSearchStep 
            v-if="showAnimeSearch"
            @select="handleAnimeSelect"
            @skip="showAnimeSearch = false"
          />
          <MovieSearchStep
            v-else-if="showMovieSearch"
            @select="handleMovieSelect"
            @skip="showMovieSearch = false"
          />
          <template v-else>

          <!-- Тип контента -->
          <div class="form-group">
            <label class="form-label">Тип</label>
            <div class="type-selector">
              <button
                v-for="t in contentTypes"
                :key="t.value"
                class="type-btn"
                :class="{ active: form.content_type === t.value }"
                @click="form.content_type = t.value"
              >{{ t.icon }} {{ t.label }}</button>
            </div>
          </div>

          <!-- Кнопка поиска для контента (только при создании) -->
          <div v-if="form.content_type === 'anime' && !isEditing" class="mb-4">
            <button class="btn btn-primary w-full text-sm py-2" @click="showAnimeSearch = true">
              🔍 Найти на Shikimori (Автозаполнение)
            </button>
          </div>
          <div v-if="(form.content_type === 'movie' || form.content_type === 'series') && !isEditing" class="mb-4">
            <button class="btn btn-primary w-full text-sm py-2" @click="showMovieSearch = true">
              🔍 Найти на TMDB (Автозаполнение)
            </button>
          </div>

          <!-- Название -->
          <div class="form-group">
            <label class="form-label">Название <span style="color:#f87171">*</span></label>
            <input v-model="form.title" type="text" class="form-input" placeholder="Введите название..." />
          </div>

          <!-- Статус -->
          <div class="form-group">
            <label class="form-label">Статус</label>
            <div class="status-selector">
              <button
                v-for="s in statusOptions"
                :key="s.value"
                class="status-opt"
                :class="{ active: form.status === s.value }"
                :style="form.status === s.value ? { borderColor: s.color, background: s.color + '22' } : {}"
                @click="form.status = s.value"
              >{{ s.icon }} {{ s.label }}</button>
            </div>
          </div>

          <!-- Оценка -->
          <div class="form-group">
            <label class="form-label">
              Оценка
              <span style="color:var(--text-muted);font-size:12px;margin-left:6px;">({{ form.rating ? form.rating + '/10' : 'не задана' }})</span>
            </label>
            <div class="rating-row">
              <button
                v-for="n in 10"
                :key="n"
                class="rating-btn"
                :class="{ active: form.rating >= n, high: n >= 8, mid: n >= 5 && n < 8 }"
                @click="form.rating = form.rating === n ? null : n"
              >{{ n }}</button>
              <button class="rating-btn clear-btn" v-if="form.rating" @click="form.rating = null" title="Сбросить">✕</button>
            </div>
          </div>

          <!-- Постер URL -->
          <div class="form-group">
            <label class="form-label">URL постера <span style="color:var(--text-muted);font-size:12px;">(необязательно)</span></label>
            <input v-model="form.poster_url" type="url" class="form-input" placeholder="https://..." />
            <div v-if="form.poster_url" class="poster-preview">
              <img :src="form.poster_url" alt="Постер" @error="posterLoadError = true" @load="posterLoadError = false" />
              <span v-if="posterLoadError" class="poster-error">❌ Не удалось загрузить изображение</span>
            </div>
          </div>

          <!-- Прогресс просмотра (серии) -->
          <div class="form-group" v-if="form.content_type !== 'movie'">
            <label class="form-label">Прогресс серий</label>
            <div class="flex items-center gap-2">
              <input v-model.number="form.current_episode" type="number" min="0" class="form-input text-center" style="width: 80px;" placeholder="Текущая" />
              <span class="text-gray-400">из</span>
              <input v-model.number="form.episodes_total" type="number" min="0" class="form-input text-center" style="width: 80px;" placeholder="Всего" />
            </div>
          </div>

          <!-- Заметка -->
          <div class="form-group">
            <label class="form-label">Заметка <span style="color:var(--text-muted);font-size:12px;">(необязательно)</span></label>
            <textarea v-model="form.notes" class="form-textarea" rows="2" placeholder="Впечатления, мысли..."></textarea>
          </div>

          <!-- Жанры -->
          <div class="form-group">
            <label class="form-label">Жанры</label>

            <!-- Текущие жанры -->
            <div class="genres-list flex flex-wrap gap-2 mb-1" v-if="form.genres && form.genres.length > 0">
              <span v-for="g in form.genres" :key="g.name" class="genre-pill">
                {{ g.name }}
                <button type="button" class="genre-del-btn" @click="removeGenre(g)">×</button>
              </span>
            </div>

            <!-- Быстрые жанры -->
            <div class="quick-genres flex flex-wrap gap-1 mb-2">
              <button
                v-for="gName in availableQuickGenres"
                :key="gName"
                type="button"
                class="quick-genre-btn"
                :class="{ active: form.genres.some(g => g.name.toLowerCase() === gName.toLowerCase()) }"
                @click="toggleQuickGenre(gName)"
              >
                {{ gName }}
              </button>
            </div>

            <!-- Свой жанр -->
            <div class="flex gap-2">
              <input
                v-model="newGenreName"
                type="text"
                class="form-input"
                placeholder="Свой жанр..."
                @keydown.enter.prevent="addGenre(newGenreName)"
              />
              <button type="button" class="btn btn-ghost" style="padding: 9px 16px;" @click="addGenre(newGenreName)">＋</button>
            </div>
          </div>

          <!-- ── Ссылки ── -->
          <div class="form-group">
            <label class="form-label">Ссылки</label>

            <!-- Список существующих ссылок (при редактировании) -->
            <div v-if="isEditing && editingReview?.links?.length > 0" class="existing-links">
              <div v-for="link in editingReview.links" :key="link.id" class="existing-link">
                <div class="link-info">
                  <span class="link-label-text">{{ link.label || '—' }}</span>
                  <a :href="link.url" target="_blank" class="link-url-text" rel="noopener">{{ truncUrl(link.url) }}</a>
                </div>
                <button class="link-del-btn" @click="handleDeleteLink(link.id)" title="Удалить">✕</button>
              </div>
            </div>

            <!-- Добавить новую ссылку -->
            <div v-for="(nl, i) in newLinks" :key="i" class="new-link-row">
              <input v-model="nl.label" type="text" class="form-input link-label-input" placeholder="Kinopoisk / IMDB..." />
              <input v-model="nl.url" type="url" class="form-input link-url-input" placeholder="https://..." />
              <button class="link-del-btn" @click="newLinks.splice(i, 1)">✕</button>
            </div>

            <button class="add-link-btn" @click="newLinks.push({ label: '', url: '' })">
              ＋ Добавить ссылку
            </button>
          </div>
          </template>
        </div>

        <!-- Кнопки -->
        <div class="modal-footer">
          <button
            v-if="isEditing"
            class="btn btn-ghost"
            style="color:#f87171;"
            @click="handleDelete"
          >🗑 Удалить</button>
          <div class="flex items-center gap-3" style="margin-left:auto;">
            <button class="btn btn-ghost" @click="closeModal">Отмена</button>
            <button class="btn btn-primary" :disabled="!form.title.trim() || store.saving" @click="handleSave">
              {{ store.saving ? 'Сохраняется...' : (isEditing ? 'Сохранить' : 'Добавить') }}
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Диалог подтверждения удаления -->
    <div v-if="showDeleteConfirm" class="modal-overlay" @click.self="showDeleteConfirm = false">
      <div class="confirm-box glass">
        <div style="font-size:32px;margin-bottom:12px;">🗑</div>
        <h3 style="font-weight:700;margin-bottom:8px;">Удалить рецензию?</h3>
        <p style="font-size:13px;color:var(--text-secondary);margin-bottom:20px;">
          «{{ editingReview?.title }}» будет удалена без возможности восстановления.
        </p>
        <div class="flex items-center gap-3" style="justify-content:flex-end;">
          <button class="btn btn-ghost" @click="showDeleteConfirm = false">Отмена</button>
          <button class="btn btn-primary" style="background:#ef4444;" @click="confirmDelete">Удалить</button>
        </div>
      </div>
    </div>
    
    <!-- Диалог выбора эпизода -->
    <EpisodePickerModal 
      v-if="showEpisodePicker"
      :alias="activeAlias"
      :shikimori-id="activeShikimoriId"
      @close="showEpisodePicker = false"
      @select="onEpisodeSelect"
    />

  </div>
</template>

<script setup>
import { ref, computed, onMounted, reactive, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useReviewsStore } from '@/stores/reviews'
import { useGroupsStore } from '@/stores/groups'
import { useAuthStore } from '@/stores/auth'
import AnimeSearchStep from '@/components/reviews/AnimeSearchStep.vue'
import MovieSearchStep from '@/components/reviews/MovieSearchStep.vue'
import EpisodePickerModal from '@/components/reviews/EpisodePickerModal.vue'

const router = useRouter()
const store = useReviewsStore()
const groupsStore = useGroupsStore()
const auth = useAuthStore()

async function syncWithShikimori() {
  try {
    await store.syncWithShikimori()
  } catch (err) {
    alert("Ошибка при синхронизации: " + (err.message || err.error || JSON.stringify(err)))
  }
}

const loadedImages = ref({})
const isLoadingMore = ref(false)

// Observer и пагинация будут настроены ниже, после инициализации currentReviews

onMounted(async () => {
  await store.fetchReviews()
  await groupsStore.fetchInvites()
  
  observer = new IntersectionObserver((entries) => {
    if (entries[0].isIntersecting && !isLoadingMore.value && visibleLimit.value < currentReviews.value.length) {
      isLoadingMore.value = true
      setTimeout(() => {
        visibleLimit.value += 24
        isLoadingMore.value = false
      }, 500) // небольшая задержка для плавности пагинации
    }
  }, { rootMargin: '150px' })
  
  if (loadMoreTrigger.value) {
    observer.observe(loadMoreTrigger.value)
  }
})

// ── Конфигурация ──

const tabs = [
  {
    key: 'watching',
    icon: '📺',
    label: 'Смотрю',
    countBg: 'rgba(6,182,212,0.3)',
    emptyTitle: 'Сейчас ничего не смотришь',
    emptyDesc: 'Добавь то, что смотришь прямо сейчас',
  },
  {
    key: 'on_hold',
    icon: '⏸',
    label: 'Отложено',
    countBg: 'rgba(234,179,8,0.3)', // yellow-ish
    emptyTitle: 'Нет отложенных',
    emptyDesc: 'Здесь будут тайтлы, отложенные на потом',
  },
  {
    key: 'completed',
    icon: '✅',
    label: 'Просмотрено',
    countBg: 'rgba(34,197,94,0.3)',
    emptyTitle: 'Список просмотренного пуст',
    emptyDesc: 'Добавляй фильмы и аниме, которые уже посмотрел',
  },
  {
    key: 'planned',
    icon: '📋',
    label: 'Запланировано',
    countBg: 'rgba(99,102,241,0.3)',
    emptyTitle: 'Список планов пуст',
    emptyDesc: 'Добавляй что планируешь посмотреть',
  },
  {
    key: 'dropped',
    icon: '⛔',
    label: 'Брошено',
    countBg: 'rgba(239,68,68,0.3)',
    emptyTitle: 'Брошенного нет',
    emptyDesc: 'Здесь будет то, что не смог досмотреть',
  },
]

const contentTypes = [
  { value: 'movie',  icon: '🎬', label: 'Фильм'  },
  { value: 'anime',  icon: '✨', label: 'Аниме'  },
  { value: 'series', icon: '📺', label: 'Сериал' },
]

const statusOptions = [
  { value: 'watching',  icon: '📺', label: 'Смотрю',       color: '#06b6d4' },
  { value: 'completed', icon: '✅', label: 'Просмотрено',  color: '#22c55e' },
  { value: 'planned',   icon: '📋', label: 'Запланировано', color: '#6366f1' },
  { value: 'on_hold',   icon: '⏸', label: 'Отложено',     color: '#eab308' },
  { value: 'dropped',   icon: '⛔', label: 'Брошено',      color: '#ef4444' },
]

// ── Поиск и фильтры ──

const searchQuery        = ref('')
const selectedContentType= ref('all')
const selectedGenres     = ref(new Set())
const selectedShikimoriFilter = ref('all') // 'all', 'app_only', 'shikimori_only'

const isAnyFilterActive = computed(() => {
  return searchQuery.value.trim() !== '' 
      || selectedContentType.value !== 'all' 
      || selectedGenres.value.size > 0
      || selectedShikimoriFilter.value !== 'all'
})

function resetFilters() {
  searchQuery.value = ''
  selectedContentType.value = 'all'
  selectedGenres.value.clear()
  selectedShikimoriFilter.value = 'all'
}

function toggleFilterGenre(genreName) {
  if (selectedGenres.value.has(genreName)) {
    selectedGenres.value.delete(genreName)
  } else {
    selectedGenres.value.add(genreName)
  }
}

const availableGenres = computed(() => {
  const list = store.byStatus[activeTab.value] || []
  const genresSet = new Set()
  for (const r of list) {
    if (r.genres) {
      for (const g of r.genres) {
        genresSet.add(g.name)
      }
    }
  }
  return Array.from(genresSet).sort()
})

// ── Табы ──

const activeTab    = ref('watching')
const currentTab   = computed(() => tabs.find(t => t.key === activeTab.value))

const currentReviews = computed(() => {
  let list = store.byStatus[activeTab.value] || []

  if (searchQuery.value.trim()) {
    const q = searchQuery.value.trim().toLowerCase()
    list = list.filter(r => r.title.toLowerCase().includes(q))
  }

  if (selectedContentType.value && selectedContentType.value !== 'all') {
    list = list.filter(r => r.content_type === selectedContentType.value)
  }

  if (selectedShikimoriFilter.value === 'app_only') {
    list = list.filter(r => !r.shikimori_id)
  } else if (selectedShikimoriFilter.value === 'shikimori_only') {
    list = list.filter(r => !!r.shikimori_id)
  }

  if (selectedGenres.value.size > 0) {
    list = list.filter(r => {
      const reviewGenreNames = new Set((r.genres || []).map(g => g.name.toLowerCase()))
      return Array.from(selectedGenres.value).every(gName => reviewGenreNames.has(gName.toLowerCase()))
    })
  }

  return list
})
watch(activeTab, () => {
  const av = new Set(availableGenres.value.map(g => g.toLowerCase()))
  for (const g of Array.from(selectedGenres.value)) {
    if (!av.has(g.toLowerCase())) {
      selectedGenres.value.delete(g)
    }
  }
})

// ── Модал ──

const showModal       = ref(false)
const showDeleteConfirm = ref(false)
const isEditing       = ref(false)
const editingReview   = ref(null)
const posterLoadError = ref(false)
const newLinks        = ref([])  // новые ссылки для добавления
const newGenreName    = ref('')  // имя нового жанра в инпуте
const showAnimeSearch = ref(false)
const showMovieSearch = ref(false)
const isSaving = ref(false)

const defaultGenres = {
  movie: ['Боевик', 'Комедия', 'Драма', 'Триллер', 'Ужасы', 'Фантастика', 'Фэнтези', 'Мелодрама', 'Криминал', 'Приключения', 'Аниме-фильм', 'Документальный'],
  series: ['Боевик', 'Комедия', 'Драма', 'Триллер', 'Ужасы', 'Фантастика', 'Фэнтези', 'Мелодрама', 'Криминал', 'Приключения', 'Аниме-фильм', 'Документальный'],
  anime: ['Сёнэн', 'Сёдзё', 'Сэйнэн', 'Меха', 'Исекай', 'Романтика', 'Этти', 'Слайс-оф-лайф', 'Спокон', 'Психологическое', 'Ужасы', 'Фэнтези']
}

const availableQuickGenres = computed(() => {
  return defaultGenres[form.content_type] || defaultGenres.movie
})

const form = reactive({
  title:        '',
  content_type: 'movie',
  status:       'planned',
  rating:       null,
  notes:        '',
  poster_url:   '',
  genres:       [],
  shikimori_id: null,
  description: '',
  episodes_total: null,
  aniliberty_alias: '',
  shikimori_score: null,
  current_episode: 0,
})

function resetForm() {
  form.title        = ''
  form.content_type = 'movie'
  form.status       = activeTab.value in { watching:1, completed:1, planned:1, dropped:1 }
    ? activeTab.value : 'planned'
  form.rating       = null
  form.notes        = ''
  form.poster_url   = ''
  form.genres       = []
  form.shikimori_id = null
  form.description = ''
  form.episodes_total = null
  form.aniliberty_alias = ''
  form.shikimori_score = null
  form.current_episode = 0
  newLinks.value    = []
  newGenreName.value = ''
  posterLoadError.value = false
  showAnimeSearch.value = false
  showMovieSearch.value = false
}


async function incrementEpisode(rev) {
  if (rev.current_episode >= rev.episodes_total) return
  const newEpisode = (rev.current_episode || 0) + 1
  
  const payload = {
    title: rev.title,
    content_type: rev.content_type,
    status: newEpisode === rev.episodes_total ? 'completed' : rev.status,
    rating: rev.rating,
    notes: rev.notes,
    poster_url: rev.poster_url,
    shikimori_id: rev.shikimori_id,
    description: rev.description,
    episodes_total: rev.episodes_total,
    current_episode: newEpisode,
    aniliberty_alias: rev.aniliberty_alias,
    shikimori_score: rev.shikimori_score
  }
  
  await store.updateReview(rev.id, payload)
}

function openCreate() {
  showAnimeSearch.value = false
  showMovieSearch.value = false

  isEditing.value    = false
  editingReview.value = null
  resetForm()
  form.status = activeTab.value
  showModal.value = true
}

function openEdit(rev) {
  isEditing.value     = true
  editingReview.value = rev
  form.title          = rev.title
  form.content_type   = rev.content_type
  form.status         = rev.status
  form.rating         = rev.rating ?? null
  form.notes          = rev.notes
  form.poster_url     = rev.poster_url
  form.genres         = rev.genres ? [...rev.genres] : []
  form.shikimori_id   = rev.shikimori_id || null
  form.description    = rev.description || ''
  form.episodes_total = rev.episodes_total || null
  form.aniliberty_alias = rev.aniliberty_alias || ''
  form.shikimori_score = rev.shikimori_score
  form.current_episode = rev.current_episode || 0 || null
  newLinks.value      = []
  newGenreName.value = ''
  posterLoadError.value = false
  showAnimeSearch.value = false
  showMovieSearch.value = false
  showModal.value = true
}

function handleAnimeSelect(anime) {
  form.title = anime.title
  form.poster_url = anime.posterFull || anime.poster
  form.shikimori_id = anime.id
  form.shikimori_score = anime.score
  form.current_episode = anime.episodes || 0
  
  if (anime.details) {
    form.description = anime.details.description || ''
    if (anime.details.episodes) form.episodes_total = anime.details.episodes
    if (anime.details.genres) {
      anime.details.genres.forEach(g => {
        if (!form.genres.some(fg => fg.name.toLowerCase() === g.russian.toLowerCase())) {
          form.genres.push({ name: g.russian })
        }
      })
    }
  }

  if (anime.anilibertyRelease) {
    form.aniliberty_alias = anime.anilibertyRelease.alias
  }

  showAnimeSearch.value = false
}

function handleMovieSelect(item) {
  form.title = item.title
  form.description = item.overview || ''
  
  if (item.poster_url) {
    form.poster_url = item.poster_url
  }
  
  form.tmdb_id = item.id
  form.rating = item.vote_average ? Math.round(item.vote_average) : null

  if (item.media_type === 'movie') {
    form.content_type = 'movie'
    form.episodes_total = 1
  } else if (item.media_type === 'tv') {
    form.content_type = 'series'
    if (item.details && item.details.number_of_episodes) {
      form.episodes_total = item.details.number_of_episodes
    }
  }

  if (item.details && item.details.genres && item.details.genres.length > 0) {
    form.genres = item.details.genres.map(g => g.name)
  }

  showMovieSearch.value = false
}

function closeModal() {
  showModal.value = false
  isEditing.value = false
  editingReview.value = null
}

// ── Жанры в Модале ──

async function addGenre(name) {
  name = name.trim()
  if (!name) return
  if (form.genres.some(g => g.name.toLowerCase() === name.toLowerCase())) {
    newGenreName.value = ''
    return
  }
  if (isEditing.value) {
    const genre = await store.addGenre(editingReview.value.id, name)
    form.genres.push(genre)
  } else {
    form.genres.push({ name })
  }
  newGenreName.value = ''
}

async function removeGenre(genre) {
  if (isEditing.value && genre.id) {
    await store.deleteGenre(editingReview.value.id, genre.id)
  }
  form.genres = form.genres.filter(g => g.name.toLowerCase() !== genre.name.toLowerCase())
}

async function toggleQuickGenre(gName) {
  const existing = form.genres.find(g => g.name.toLowerCase() === gName.toLowerCase())
  if (existing) {
    await removeGenre(existing)
  } else {
    await addGenre(gName)
  }
}

// ── Сохранение ──

async function handleSave() {
  if (!form.title.trim()) return

  const payload = {
    title:            form.title.trim(),
    content_type:     form.content_type,
    status:           form.status,
    rating:           form.rating,
    notes:            form.notes.trim(),
    poster_url:       form.poster_url.trim(),
    shikimori_id:     form.shikimori_id || null,
    description:      form.description || '',
    episodes_total:   form.episodes_total || null,
    current_episode:  form.current_episode || 0,
    aniliberty_alias: form.aniliberty_alias || '',
    shikimori_score:  form.shikimori_score || null
  }

  let savedReview
  try {
    isSaving.value = true
    if (isEditing.value && !editingReview.value.isShikimoriOnly) {
      savedReview = await store.updateReview(editingReview.value.id, payload)
    } else {
      // Если это только с Shikimori, мы сохраняем его локально впервые
      savedReview = await store.createReview(payload)
      if (isEditing.value && editingReview.value.isShikimoriOnly) {
         // Удаляем виртуальную запись из массива
         store.reviews = store.reviews.filter(r => r.id !== editingReview.value.id)
      }
    }

    // Синхронизация с Shikimori (если привязан и это аниме)
    if (auth.user?.shikimori_user_id && payload.shikimori_id) {
       await syncToShikimori(payload.shikimori_id, payload.status, payload.rating)
    }

    // Добавляем новые ссылки если есть
    for (const nl of newLinks.value) {
      if (nl.url.trim()) {
        await store.addLink(savedReview.id, nl.label.trim(), nl.url.trim())
      }
    }

    // Добавляем новые жанры если есть (для создания)
    if (!isEditing.value) {
      for (const g of form.genres) {
        await store.addGenre(savedReview.id, g.name)
      }
    }

    // Автоматически переключаем таб, чтобы показать новую запись
    activeTab.value = form.status
    closeModal()
  } catch (err) {
    console.error('Failed to save review', err)
  } finally {
    isSaving.value = false
  }
}

// ── Удаление ──

function handleDelete() {
  showDeleteConfirm.value = true
}

async function confirmDelete() {

  if (!editingReview.value) return
  isDeleting.value = true
  try {
    if (editingReview.value.isShikimoriOnly) {
      // Удаляем виртуальную запись
      store.reviews = store.reviews.filter(r => r.id !== editingReview.value.id)
    } else {
      await store.deleteReview(editingReview.value.id)
    }
    showDeleteConfirm.value = false
    closeModal()
  } catch (err) {
    console.error('Failed to delete review', err)
  } finally {
    isDeleting.value = false
  }
}

async function syncToShikimori(animeId, status, score) {
  try {
    const payload = {
      user_rate: {
        target_id: animeId,
        target_type: 'Anime',
        status: status,
        score: score || 0
      }
    }
    await fetch('/api/v1/auth/shikimori/rates', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${auth.accessToken}`
      },
      body: JSON.stringify(payload)
    })
  } catch (err) {
    console.warn('Failed to sync to Shikimori', err)
  }
}

async function handleDeleteLink(linkId) {
  await store.deleteLink(editingReview.value.id, linkId)
  // editingReview обновится реактивно через store
}

// ── Вспомогательные функции ──

const posterGradients = {
  movie:  'linear-gradient(135deg, #312e81 0%, #1e1b4b 100%)',
  anime:  'linear-gradient(135deg, #831843 0%, #500724 100%)',
  series: 'linear-gradient(135deg, #134e4a 0%, #042f2e 100%)',
}

function posterStyle(rev) {
  return { background: posterGradients[rev.content_type] || posterGradients.movie }
}

const typeColors = { movie: '#6366f1', anime: '#ec4899', series: '#06b6d4' }
const typeIcons  = { movie: '🎬', anime: '✨', series: '📺' }
const typeLabels = { movie: 'Фильм', anime: 'Аниме', series: 'Сериал' }

function typeColor(ct)  { return typeColors[ct] || typeColors.movie }
function typeIcon(ct)   { return typeIcons[ct]  || '🎬' }
function typeLabel(ct)  { return typeLabels[ct] || ct }

function truncUrl(url) {
  try {
    const u = new URL(url)
    return u.hostname + (u.pathname.length > 20 ? u.pathname.slice(0, 20) + '…' : u.pathname)
  } catch { return url.slice(0, 40) }
}

function generateUUID() {
  return Math.random().toString(36).substring(2, 10)
}

const activeAlias = ref(null)
const activeShikimoriId = ref(null)
const showEpisodePicker = ref(false)

function handleWatchTogether(rev) {
  activeShikimoriId.value = rev.shikimori_id || null
  activeAlias.value = rev.aniliberty_alias || null
  if (rev.shikimori_id || rev.aniliberty_alias) {
    showEpisodePicker.value = true
  } else if (rev.links && rev.links.length > 0) {
    openWatchParty(rev.links[0].url)
  }
}

function onEpisodeSelect(url) {
  showEpisodePicker.value = false
  openWatchParty(url, activeShikimoriId.value, activeAlias.value)
}

function openWatchParty(videoUrl, shikimoriId, alias) {
  const roomId = generateUUID()
  sessionStorage.setItem(`wp_url_${roomId}`, videoUrl)
  if (shikimoriId) {
    sessionStorage.setItem(`wp_shikimori_${roomId}`, shikimoriId)
  }
  if (alias) {
    sessionStorage.setItem(`wp_alias_${roomId}`, alias)
  }
  router.push(`/watch/room/${roomId}`).catch(err => {
    console.error('Router push error:', err)
    window.location.href = `/watch/room/${roomId}`
  })
}
</script>

<style scoped>
/* ══ Страница ══ */
.reviews-page {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--bg-base);
}

.rv-body {
  flex: 1;
  padding: 24px;
  background: var(--bg-body);
}

.btn-increment-ep {
  background: rgba(99, 102, 241, 0.15);
  border: 1px solid rgba(99, 102, 241, 0.35);
  color: #a5b4fc;
  font-size: 10.5px;
  font-weight: 700;
  padding: 1px 6px;
  border-radius: 4px;
  cursor: pointer;
  margin-left: 4px;
  transition: all 0.15s;
}
.btn-increment-ep:hover {
  background: var(--primary);
  color: white;
  border-color: var(--primary);
}

/* ══ Навбар ══ */
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

/* ══ Табы ══ */
.status-tabs {
  display: flex;
  gap: 4px;
  padding: 12px 24px;
  border-bottom: 1px solid var(--border);
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

/* ══ Тело ══ */
.rv-body { flex: 1; padding: 24px; }

.empty-state {
  display: flex; flex-direction: column;
  align-items: center; justify-content: center;
  min-height: 320px; text-align: center;
  color: var(--text-muted);
}

/* ══ Карточки ══ */
.card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 16px;
}

.rv-card {
  background: rgba(255,255,255,0.03);
  border: 1px solid var(--border);
  border-radius: var(--radius-xl);
  overflow: hidden;
  cursor: pointer;
  transition: transform 0.2s, border-color 0.2s, box-shadow 0.2s;
}
.rv-card:hover {
  transform: translateY(-3px);
  border-color: rgba(99,102,241,0.4);
  box-shadow: 0 8px 32px rgba(0,0,0,0.3);
}

/* Постер */
.card-poster {
  width: 100%;
  padding-bottom: 145%; /* соотношение ~2:3 */
  position: relative;
  overflow: hidden;
  border-radius: 12px;
  background: var(--bg-surface);
  box-shadow: 0 4px 20px rgba(0,0,0,0.3);
}

.poster-img {
  position: absolute;
  top: 0; left: 0; width: 100%; height: 100%;
  object-fit: cover;
  opacity: 0;
  transition: opacity 0.4s ease;
}
.poster-img.img-loaded {
  opacity: 1;
}
.poster-skeleton {
  position: absolute;
  top: 0; left: 0; width: 100%; height: 100%;
  background: rgba(255,255,255,0.02);
  display: flex;
  align-items: center;
  justify-content: center;
}

.card-poster-overlay {
  position: absolute; inset: 0;
  background: linear-gradient(to top, rgba(0,0,0,0.85) 0%, transparent 50%);
  pointer-events: none;
}
.card-type-badge {
  position: relative; z-index: 1;
  display: inline-flex; align-items: center; gap: 4px;
  padding: 4px 10px; border-radius: 20px;
  font-size: 11px; font-weight: 700; color: white;
  align-self: flex-start;
}
.card-rating {
  position: absolute;
  top: 8px;
  right: 8px;
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(8px);
  padding: 4px 8px;
  border-radius: 8px;
  color: white;
  display: flex;
  align-items: center;
  gap: 2px;
  font-size: 13px;
  font-weight: 700;
  border: 1px solid rgba(255, 255, 255, 0.1);
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
}

.shikimori-badge {
  position: absolute;
  bottom: 8px;
  left: 8px;
  background: rgba(33, 33, 33, 0.85);
  backdrop-filter: blur(4px);
  padding: 3px 6px;
  border-radius: 4px;
  color: #fff;
  font-size: 10px;
  font-weight: 600;
  border: 1px solid rgba(255, 255, 255, 0.1);
}

/* Информация */
.card-info { padding: 10px 12px 12px; }
.card-title {
  font-size: 14px; font-weight: 600; color: var(--text-primary);
  display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical;
  overflow: hidden; line-height: 1.4; margin-bottom: 4px;
}
.card-notes {
  font-size: 12px; color: var(--text-muted);
  display: -webkit-box; -webkit-line-clamp: 1; -webkit-box-orient: vertical;
  overflow: hidden; margin-bottom: 8px;
}
.card-links { display: flex; flex-wrap: wrap; gap: 5px; }
.link-pill {
  font-size: 11px; font-weight: 500;
  padding: 3px 9px; border-radius: 20px;
  background: rgba(255,255,255,0.06);
  border: 1px solid var(--border);
  color: var(--text-secondary); text-decoration: none;
  transition: background 0.15s, color 0.15s;
}
.link-pill:hover { background: rgba(99,102,241,0.25); color: #fff; }

.watch-together-btn {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 5px 12px;
  border-radius: var(--radius-md, 8px);
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.15) 0%, rgba(139, 92, 246, 0.15) 100%);
  color: #c7d2fe;
  border: 1px solid rgba(99, 102, 241, 0.35);
  font-size: 11px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}
.watch-together-btn:hover {
  background: linear-gradient(135deg, rgba(99, 102, 241, 0.3) 0%, rgba(139, 92, 246, 0.3) 100%);
  color: #ffffff;
  border-color: rgba(99, 102, 241, 0.6);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(99, 102, 241, 0.2);
}
.watch-together-btn:active {
  transform: translateY(0);
}

/* ══ Модал ══ */
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
.modal-title { font-size: 17px; font-weight: 700; }
.modal-close {
  width: 30px; height: 30px; border-radius: 50%;
  background: rgba(255,255,255,0.06); border: none;
  color: var(--text-secondary); cursor: pointer; font-size: 14px;
  display: flex; align-items: center; justify-content: center;
  transition: background 0.15s;
}
.modal-close:hover { background: rgba(255,255,255,0.12); }

.modal-body {
  overflow-y: auto; padding: 18px 22px;
  display: flex; flex-direction: column; gap: 16px;
}
.modal-body::-webkit-scrollbar { width: 4px; }
.modal-body::-webkit-scrollbar-thumb { background: var(--border); border-radius: 2px; }

.modal-footer {
  display: flex; align-items: center;
  padding: 14px 22px;
  border-top: 1px solid var(--border);
  flex-shrink: 0;
}

/* Форма */
.form-group { display: flex; flex-direction: column; gap: 7px; }
.form-label { font-size: 13px; font-weight: 600; color: var(--text-secondary); }
.form-input, .form-textarea {
  background: rgba(255,255,255,0.04); border: 1px solid var(--border);
  border-radius: var(--radius-md); color: var(--text-primary);
  font-size: 14px; padding: 9px 12px; outline: none;
  transition: border-color 0.2s; font-family: inherit; width: 100%;
}
.form-input:focus, .form-textarea:focus { border-color: var(--primary); }
.form-input::placeholder, .form-textarea::placeholder { color: var(--text-muted); }
.form-textarea { resize: vertical; min-height: 60px; }

/* Тип контента */
.type-selector { display: flex; gap: 8px; }
.type-btn {
  flex: 1; padding: 9px 6px; border-radius: var(--radius-md);
  border: 1px solid var(--border); background: rgba(255,255,255,0.03);
  color: var(--text-secondary); cursor: pointer; font-size: 13px;
  transition: all 0.15s; text-align: center;
}
.type-btn:hover { background: rgba(255,255,255,0.07); color: var(--text-primary); }
.type-btn.active { border-color: var(--primary); background: rgba(99,102,241,0.15); color: #a5b4fc; font-weight: 600; }

/* Статус */
.status-selector { display: grid; grid-template-columns: 1fr 1fr; gap: 6px; }
.status-opt {
  padding: 8px 10px; border-radius: var(--radius-md);
  border: 1px solid var(--border); background: rgba(255,255,255,0.03);
  color: var(--text-secondary); cursor: pointer; font-size: 12px;
  transition: all 0.15s;
}
.status-opt:hover { background: rgba(255,255,255,0.06); color: var(--text-primary); }
.status-opt.active { font-weight: 600; color: var(--text-primary); }

/* Рейтинг */
.rating-row { display: flex; align-items: center; gap: 5px; flex-wrap: wrap; }
.rating-btn {
  width: 32px; height: 32px; border-radius: var(--radius-md);
  border: 1px solid var(--border); background: rgba(255,255,255,0.04);
  color: var(--text-secondary); cursor: pointer; font-size: 13px;
  font-weight: 600; transition: all 0.15s;
}
.rating-btn:hover { background: rgba(255,255,255,0.1); color: var(--text-primary); }
.rating-btn.active { background: rgba(251,191,36,0.2); border-color: rgba(251,191,36,0.5); color: #fbbf24; }
.rating-btn.active.high { background: rgba(34,197,94,0.2); border-color: rgba(34,197,94,0.5); color: #4ade80; }
.rating-btn.active.mid  { background: rgba(251,191,36,0.2); border-color: rgba(251,191,36,0.5); color: #fbbf24; }
.clear-btn { font-size: 11px; color: #f87171; border-color: rgba(248,113,113,0.3); }

/* Постер превью */
.poster-preview {
  margin-top: 6px;
  border-radius: var(--radius-md);
  overflow: hidden;
  max-height: 180px;
  display: flex; align-items: center; justify-content: center;
  background: rgba(255,255,255,0.04); border: 1px solid var(--border);
}
.poster-preview img { max-height: 180px; object-fit: contain; width: 100%; }
.poster-error { font-size: 12px; color: #f87171; padding: 12px; }

/* Существующие ссылки */
.existing-links { display: flex; flex-direction: column; gap: 5px; margin-bottom: 8px; }
.existing-link {
  display: flex; align-items: center; justify-content: space-between;
  background: rgba(255,255,255,0.04); border: 1px solid var(--border);
  border-radius: var(--radius-md); padding: 7px 10px; gap: 8px;
}
.link-info { display: flex; flex-direction: column; gap: 2px; min-width: 0; }
.link-label-text { font-size: 12px; font-weight: 600; color: var(--text-primary); }
.link-url-text {
  font-size: 11px; color: #a5b4fc; text-decoration: none;
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
}
.link-url-text:hover { text-decoration: underline; }

/* Новые ссылки */
.new-link-row {
  display: grid; grid-template-columns: 1fr 2fr 28px; gap: 6px;
  align-items: center; margin-bottom: 6px;
}
.link-label-input, .link-url-input { font-size: 13px; padding: 7px 10px; }
.link-del-btn {
  width: 28px; height: 28px; border-radius: var(--radius-md);
  background: rgba(248,113,113,0.1); border: 1px solid rgba(248,113,113,0.3);
  color: #f87171; cursor: pointer; font-size: 12px;
  display: flex; align-items: center; justify-content: center;
  transition: background 0.15s; flex-shrink: 0;
}
.link-del-btn:hover { background: rgba(248,113,113,0.2); }

.add-link-btn {
  display: flex; align-items: center; gap: 6px;
  padding: 7px 12px; border-radius: var(--radius-md);
  border: 1px dashed var(--border); background: transparent;
  color: var(--text-muted); cursor: pointer; font-size: 13px;
  transition: all 0.15s; width: 100%; justify-content: center;
}
.add-link-btn:hover { border-color: var(--primary); color: #a5b4fc; background: rgba(99,102,241,0.05); }

/* Фильтры и поиск */
.filter-panel {
  margin: 12px 24px 0 24px;
  padding: 14px 20px;
  border-radius: var(--radius-md);
  border: 1px solid var(--border);
}
.radio-label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  font-size: 13px;
  color: var(--text-secondary);
  cursor: pointer;
}
.radio-label input {
  accent-color: var(--primary);
}
.filter-genre-btn {
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid var(--border);
  color: var(--text-secondary);
  padding: 3px 8px;
  border-radius: var(--radius-sm);
  font-size: 11px;
  cursor: pointer;
  transition: all 0.15s;
}
.filter-genre-btn:hover {
  background: rgba(255, 255, 255, 0.06);
  color: var(--text-primary);
}
.filter-genre-btn.active {
  background: rgba(99, 102, 241, 0.25);
  border-color: var(--primary);
  color: white;
}
.reset-filters-btn {
  color: #f87171;
  border-color: rgba(248, 113, 113, 0.2);
}
.reset-filters-btn:hover {
  background: rgba(248, 113, 113, 0.08);
  color: #f87171;
}

/* Жанры на карточке */
.card-genre-pill {
  font-size: 10px;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid var(--border);
  color: var(--text-secondary);
  padding: 2px 6px;
  border-radius: var(--radius-sm);
  display: inline-block;
}

/* Жанры в модале */
.genre-pill {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  background: rgba(99, 102, 241, 0.15);
  border: 1px solid rgba(99, 102, 241, 0.3);
  color: #a5b4fc;
  padding: 4px 8px;
  border-radius: var(--radius-sm);
  font-size: 12px;
  font-weight: 500;
}
.genre-del-btn {
  background: transparent;
  border: none;
  color: #fca5a5;
  cursor: pointer;
  font-size: 12px;
  font-weight: bold;
}
.genre-del-btn:hover {
  color: #ef4444;
}
.quick-genre-btn {
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid var(--border);
  color: var(--text-secondary);
  padding: 4px 8px;
  border-radius: var(--radius-sm);
  font-size: 11px;
  cursor: pointer;
  transition: all 0.15s;
}
.quick-genre-btn:hover {
  background: rgba(255, 255, 255, 0.08);
  color: var(--text-primary);
}
.quick-genre-btn.active {
  background: rgba(99, 102, 241, 0.2);
  border-color: var(--primary);
  color: white;
}

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

/* ══ Адаптив ══ */
@media (max-width: 600px) {
  .card-grid { grid-template-columns: 1fr 1fr; }
  .status-tabs { padding: 10px 12px; }
  .rv-body { padding: 16px 12px; }
  .modal-box { max-width: 100%; max-height: 95vh; }
  .type-selector { flex-direction: column; }
  .new-link-row { grid-template-columns: 1fr; }
}

.modern-pills {
  display: inline-flex;
  background: rgba(0, 0, 0, 0.2);
  border-radius: 9999px;
  padding: 4px;
}
.pill-btn {
  padding: 6px 14px;
  border-radius: 9999px;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-muted);
  transition: all 0.2s ease;
  background: transparent;
  border: none;
  cursor: pointer;
}
.pill-btn:hover {
  color: var(--text-main);
}
.pill-btn.active {
  background: var(--primary-color, #6366f1);
  color: #fff;
  box-shadow: 0 2px 8px rgba(99, 102, 241, 0.4);
}

</style>

<style scoped>
/* Shikimori Filter */
.shikimori-filter {
  display: flex;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 8px;
  padding: 4px;
  gap: 4px;
}
.shikimori-btn {
  background: transparent;
  border: none;
  border-radius: 6px;
  padding: 6px 12px;
  cursor: pointer;
  transition: all 0.2s;
  display: flex;
  align-items: center;
  justify-content: center;
  opacity: 0.6;
}
.shikimori-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  opacity: 0.9;
}
.shikimori-btn.active {
  background: rgba(255, 255, 255, 0.15);
  opacity: 1;
}
.shiki-icon {
  width: 14px;
  height: 14px;
  border: 2px solid white;
  border-radius: 3px;
}
.shiki-full {
  background: white;
}
.shiki-empty {
  background: transparent;
}
.shiki-half {
  background: linear-gradient(135deg, white 50%, transparent 50%);
}

.modern-pills {
  display: inline-flex;
  background: rgba(0, 0, 0, 0.2);
  border-radius: 9999px;
  padding: 4px;
}
.pill-btn {
  padding: 6px 14px;
  border-radius: 9999px;
  font-size: 13px;
  font-weight: 500;
  color: var(--text-muted);
  transition: all 0.2s ease;
  background: transparent;
  border: none;
  cursor: pointer;
}
.pill-btn:hover {
  color: var(--text-main);
}
.pill-btn.active {
  background: var(--primary-color, #6366f1);
  color: #fff;
  box-shadow: 0 2px 8px rgba(99, 102, 241, 0.4);
}

</style>

<style scoped>
.btn-sync {
  position: relative;
  overflow: hidden;
  transition: all 0.3s;
}
.btn-sync.is-syncing {
  border-color: rgba(59, 130, 246, 0.5);
  background: rgba(59, 130, 246, 0.05);
}
.sync-progress {
  position: absolute;
  bottom: 0;
  left: 0;
  height: 3px;
  background: #3b82f6; /* primary color */
  transition: width 0.4s ease-out;
  border-radius: 0 2px 0 0;
}
</style>
