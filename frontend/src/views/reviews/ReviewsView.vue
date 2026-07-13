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
          {{ (store.byStatus[tab.key] || []).length }}
        </span>
      </button>
    </div>

    <!-- ══ ОСНОВНОЙ КОНТЕНТ (SPLIT PANE CATALOG) ══ -->
    <div class="catalog-layout">
      
      <!-- Левый Сайдбар Фильтров -->
      <aside class="sidebar-filters glass">
        <!-- Поиск -->
        <div class="sidebar-group">
          <label class="sidebar-label">Поиск</label>
          <div class="search-input-wrapper">
            <input
              v-model="searchQuery"
              type="text"
              class="form-input search-input"
              placeholder="Название тайтла..."
            />
            <span class="search-indicator-icon">🔍</span>
          </div>
        </div>

        <!-- Тип медиа -->
        <div class="sidebar-group">
          <label class="sidebar-label">Тип контента</label>
          <div class="stack-filters">
            <button
              class="stack-filter-btn"
              :class="{ active: selectedContentType === 'all' }"
              @click="selectedContentType = 'all'"
            >
              <span>📺 Все медиа</span>
            </button>
            <button
              v-for="t in contentTypes"
              :key="t.value"
              class="stack-filter-btn"
              :class="{ active: selectedContentType === t.value }"
              @click="selectedContentType = t.value"
            >
              <span>{{ t.icon }} {{ t.label }}</span>
            </button>
          </div>
        </div>

        <!-- Источник контента -->
        <div class="sidebar-group">
          <label class="sidebar-label">Источник</label>
          <div class="source-selector-grid">
            <button
              class="source-btn"
              :class="{ active: selectedShikimoriFilter === 'all' }"
              @click="selectedShikimoriFilter = 'all'"
            >
              Все
            </button>
            <button
              class="source-btn"
              :class="{ active: selectedShikimoriFilter === 'shikimori_only' }"
              @click="selectedShikimoriFilter = 'shikimori_only'"
            >
              Shiki
            </button>
            <button
              class="source-btn"
              :class="{ active: selectedShikimoriFilter === 'app_only' }"
              @click="selectedShikimoriFilter = 'app_only'"
            >
              Ручные
            </button>
          </div>
        </div>

        <!-- Мультиселект жанров -->
        <div class="sidebar-group" v-if="availableGenres.length > 0">
          <label class="sidebar-label">Фильтр по жанрам</label>
          <div class="genres-multiselect">
            <button
              v-for="genreName in availableGenres"
              :key="genreName"
              class="genre-select-badge"
              :class="{ active: selectedGenres.has(genreName) }"
              @click="toggleFilterGenre(genreName)"
            >
              {{ genreName }}
            </button>
          </div>
        </div>

        <!-- Кнопка сброса -->
        <button
          v-if="isAnyFilterActive"
          @click="resetFilters"
          class="btn btn-ghost w-full text-xs mt-4"
          style="padding: 8px;"
        >
          Сбросить фильтры
        </button>
      </aside>

      <!-- Правый Фид Медиа Каталога -->
      <div class="catalog-feed">
        
        <!-- Заголовок фида & Переключатель видов -->
        <div class="feed-header glass">
          <div class="feed-title-meta">
            <span class="feed-count-badge">{{ currentReviews.length }}</span>
            <span class="feed-count-label">найдено в списке «{{ currentTab?.label }}»</span>
          </div>

          <div class="view-mode-toggle">
            <button
              class="mode-btn"
              :class="{ active: viewMode === 'grid' }"
              @click="toggleViewMode('grid')"
              title="Режим плитки"
            >
              ⬚ Сетка
            </button>
            <button
              class="mode-btn"
              :class="{ active: viewMode === 'list' }"
              @click="toggleViewMode('list')"
              title="Режим списка"
            >
              ☰ Список
            </button>
          </div>
        </div>

        <!-- Состояние: загрузка / пусто -->
        <div v-if="store.loading" class="empty-state">
          <div class="spinner" style="width:32px;height:32px;"></div>
        </div>

        <div v-else-if="currentReviews.length === 0" class="empty-state glass">
          <template v-if="isAnyFilterActive">
            <div style="font-size:44px;margin-bottom:12px;">🔍</div>
            <h3 style="font-weight:700;font-size:16px;margin-bottom:6px;">Ничего не найдено</h3>
            <p style="font-size:13px;color:var(--text-muted);margin-bottom:16px;">Попробуйте изменить параметры поиска или сбросить фильтры</p>
            <button class="btn btn-primary text-xs py-2 px-4" @click="resetFilters">Сбросить фильтры</button>
          </template>
          <template v-else>
            <div style="font-size:44px;margin-bottom:12px;">{{ currentTab.icon }}</div>
            <h3 style="font-weight:700;font-size:16px;margin-bottom:6px;">{{ currentTab.emptyTitle }}</h3>
            <p style="font-size:13px;color:var(--text-muted);margin-bottom:16px;">{{ currentTab.emptyDesc }}</p>
            <button class="btn btn-primary text-xs py-2 px-4" @click="openCreate">Добавить</button>
          </template>
        </div>

        <!-- Сетка карточек (Grid View) -->
        <div v-else-if="viewMode === 'grid'" class="card-grid">
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
              <div class="card-episode-progress mt-1.5 flex items-center gap-1.5 text-xs font-semibold" v-if="rev.episodes_total && rev.content_type !== 'movie'" style="color: var(--primary)">
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

        <!-- Список строк (List View) -->
        <div v-else-if="viewMode === 'list'" class="card-list">
          <div
            v-for="rev in currentReviews.slice(0, visibleLimit)"
            :key="rev.id"
            class="list-row glass glass-hover"
            @click="openEdit(rev)"
          >
            <!-- Миниатюрный постер -->
            <div class="list-poster-thumbnail" :style="rev.poster_url ? { backgroundImage: `url(${rev.poster_url})` } : { background: posterGradients[rev.content_type] || posterGradients.movie }">
              <span v-if="!rev.poster_url" class="text-sm">{{ typeIcon(rev.content_type) }}</span>
            </div>

            <!-- Контент и мета -->
            <div class="list-details">
              <div class="flex items-center gap-2 flex-wrap">
                <h4 class="list-title">{{ rev.title }}</h4>
                <span class="list-type-badge" :style="{ color: typeColor(rev.content_type), borderColor: typeColor(rev.content_type) + '30', background: typeColor(rev.content_type) + '10' }">
                  {{ typeIcon(rev.content_type) }} {{ typeLabel(rev.content_type) }}
                </span>
                <span v-if="rev.shikimori_id" class="list-shiki-badge" title="Синхронизировано с Shikimori">⛩️ Shiki</span>
              </div>
              <p class="list-notes-text" v-if="rev.notes">{{ rev.notes }}</p>
              <div class="list-genres flex flex-wrap gap-1 mt-1" v-if="rev.genres && rev.genres.length > 0">
                <span v-for="g in rev.genres" :key="g.id" class="list-genre-badge">{{ g.name }}</span>
              </div>
            </div>

            <!-- Оценки и прогресс серий -->
            <div class="list-stats-info">
              <div class="list-rating" v-if="rev.rating">
                <span style="color: var(--primary);">★</span>
                <span style="font-weight: 700;">{{ rev.rating }}</span>
                <span style="color: var(--text-muted); font-size: 11px;">/10</span>
              </div>
              
              <div class="list-episodes" v-if="rev.episodes_total && rev.content_type !== 'movie'">
                <div class="text-xs font-semibold mb-1" style="color: var(--text-secondary);">
                  Серии: {{ rev.current_episode || 0 }} / {{ rev.episodes_total }}
                </div>
                <div class="list-progress-bar">
                  <div class="list-progress-fill" :style="{ width: ((rev.current_episode || 0) / rev.episodes_total * 100) + '%' }"></div>
                </div>
              </div>
            </div>

            <!-- Действия -->
            <div class="list-row-actions" @click.stop>
              <button 
                class="list-inc-btn"
                v-if="rev.episodes_total && rev.content_type !== 'movie' && (rev.current_episode || 0) < rev.episodes_total" 
                @click="incrementEpisode(rev)"
                title="Плюс одна серия"
              >
                ＋1 серию
              </button>
              
              <button
                class="watch-together-btn btn-sm"
                v-if="rev.shikimori_id || rev.aniliberty_alias || (rev.links && rev.links.length > 0)"
                @click="handleWatchTogether(rev)"
              >
                Смотреть вместе
              </button>
            </div>
          </div>
        </div>

        <!-- Триггер подгрузки -->
        <div ref="loadMoreTrigger" class="flex justify-center py-6" style="width: 100%;">
          <div v-if="isLoadingMore" class="spinner" style="width:32px;height:32px;"></div>
        </div>
      </div>

    </div>

    <!-- ══ МОДАЛЬНОЕ ОКНО: Добавить / Редактировать ══ -->
    <Transition name="modal">
      <div v-if="showModal" class="modal-overlay" @click.self="closeModal">
        <div class="modal-box glass" @click.stop>

          <!-- Заголовок модала -->
          <div class="modal-header">
            <h2 class="modal-title">{{ isEditing ? 'Редактировать рецензию' : 'Добавить рецензию' }}</h2>
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

            <!-- Двухколоночная сетка формы -->
            <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
              
              <!-- Левая колонка: Основные метаданные -->
              <div class="flex flex-col gap-4">
                
                <!-- Тип контента -->
                <div class="form-group">
                  <label class="form-label">Тип медиа</label>
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

                <!-- Кнопки автозаполнения -->
                <div v-if="form.content_type === 'anime' && !isEditing" class="mb-1">
                  <button class="btn btn-primary w-full text-xs py-2" @click="showAnimeSearch = true">
                    🔍 Найти на Shikimori (Автозаполнение)
                  </button>
                </div>
                <div v-if="(form.content_type === 'movie' || form.content_type === 'series') && !isEditing" class="mb-1">
                  <button class="btn btn-primary w-full text-xs py-2" @click="showMovieSearch = true">
                    🔍 Найти на TMDB (Автозаполнение)
                  </button>
                </div>

                <!-- Название -->
                <div class="form-group">
                  <label class="form-label">Название тайтла <span style="color:var(--dusty-rose)">*</span></label>
                  <input v-model="form.title" type="text" class="form-input" placeholder="Введите название..." />
                </div>

                <!-- Статус -->
                <div class="form-group">
                  <label class="form-label">Статус просмотра</label>
                  <div class="status-selector">
                    <button
                      v-for="s in statusOptions"
                      :key="s.value"
                      class="status-opt"
                      :class="{ active: form.status === s.value }"
                      :style="form.status === s.value ? { borderColor: s.color, background: s.color + '1a' } : {}"
                      @click="form.status = s.value"
                    >{{ s.icon }} {{ s.label }}</button>
                  </div>
                </div>
              </div>

              <!-- Правая колонка: Прогресс, Оценка, Постер -->
              <div class="flex flex-col gap-4">
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

                <!-- Прогресс просмотра (серии) -->
                <div class="form-group" v-if="form.content_type !== 'movie'">
                  <label class="form-label">Прогресс серий</label>
                  <div class="flex items-center gap-2">
                    <input v-model.number="form.current_episode" type="number" min="0" class="form-input text-center" style="width: 80px;" placeholder="Тек." />
                    <span class="text-gray-400">из</span>
                    <input v-model.number="form.episodes_total" type="number" min="0" class="form-input text-center" style="width: 80px;" placeholder="Всего" />
                  </div>
                </div>

                <!-- Постер URL -->
                <div class="form-group">
                  <label class="form-label">URL постера</label>
                  <input v-model="form.poster_url" type="url" class="form-input" placeholder="https://ссылка-на-картинку..." />
                </div>
              </div>
            </div>

            <div class="grid grid-cols-1 gap-4 mt-4">
              <!-- Постер превью -->
              <div v-if="form.poster_url" class="poster-preview">
                <img :src="form.poster_url" alt="Постер" @error="posterLoadError = true" @load="posterLoadError = false" />
                <span v-if="posterLoadError" class="poster-error">❌ Не удалось загрузить изображение постера</span>
              </div>

              <!-- Заметка -->
              <div class="form-group">
                <label class="form-label">Заметка</label>
                <textarea v-model="form.notes" class="form-input form-textarea" rows="2" placeholder="Ваши личные впечатления..."></textarea>
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

              <!-- Ссылки -->
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
            </div>
            </template>
          </div>

          <!-- Кнопки -->
          <div class="modal-footer">
            <button
              v-if="isEditing"
              class="btn btn-ghost"
              style="color:var(--dusty-rose); border-color: rgba(201, 112, 100, 0.2);"
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
    </Transition>

    <!-- Диалог подтверждения удаления -->
    <Transition name="modal">
      <div v-if="showDeleteConfirm" class="modal-overlay" @click.self="showDeleteConfirm = false">
        <div class="confirm-box glass" @click.stop>
          <div style="font-size:32px;margin-bottom:12px;">🗑</div>
          <h3 style="font-weight:700;margin-bottom:8px;">Удалить рецензию?</h3>
          <p style="font-size:13px;color:var(--text-secondary);margin-bottom:20px;">
            «{{ editingReview?.title }}» будет удалена без возможности восстановления.
          </p>
          <div class="flex items-center gap-3" style="justify-content:flex-end;">
            <button class="btn btn-ghost" @click="showDeleteConfirm = false">Отмена</button>
            <button class="btn btn-primary" style="background:var(--dusty-rose);" @click="confirmDelete">Удалить</button>
          </div>
        </div>
      </div>
    </Transition>
    
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
const viewMode = ref(localStorage.getItem('reviews-view-mode') || 'grid')
function toggleViewMode(mode) {
  viewMode.value = mode
  localStorage.setItem('reviews-view-mode', mode)
}

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
    countBg: 'rgba(192, 133, 82, 0.25)', // Caramel
    emptyTitle: 'Сейчас ничего не смотришь',
    emptyDesc: 'Добавь то, что смотришь прямо сейчас',
  },
  {
    key: 'on_hold',
    icon: '⏸',
    label: 'Отложено',
    countBg: 'rgba(107, 74, 54, 0.25)', // Mocha
    emptyTitle: 'Нет отложенных',
    emptyDesc: 'Здесь будут тайтлы, отложенные на потом',
  },
  {
    key: 'completed',
    icon: '✅',
    label: 'Просмотрено',
    countBg: 'rgba(124, 154, 110, 0.25)', // Matcha
    emptyTitle: 'Список просмотренного пуст',
    emptyDesc: 'Добавляй фильмы и аниме, которые уже посмотрел',
  },
  {
    key: 'planned',
    icon: '📋',
    label: 'Запланировано',
    countBg: 'rgba(192, 133, 82, 0.15)', // Light Caramel
    emptyTitle: 'Список планов пуст',
    emptyDesc: 'Добавляй что планируешь посмотреть',
  },
  {
    key: 'dropped',
    icon: '⛔',
    label: 'Брошено',
    countBg: 'rgba(201, 112, 100, 0.25)', // Dusty Rose
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
  { value: 'watching',  icon: '📺', label: 'Смотрю',       color: 'var(--caramel)' },
  { value: 'completed', icon: '✅', label: 'Просмотрено',  color: 'var(--matcha)' },
  { value: 'planned',   icon: '📋', label: 'Запланировано', color: 'var(--caramel)' },
  { value: 'on_hold',   icon: '⏸', label: 'Отложено',     color: 'var(--mocha)' },
  { value: 'dropped',   icon: '⛔', label: 'Брошено',      color: 'var(--dusty-rose)' },
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
    form.genres = item.details.genres.map(g => ({ name: g.name }))
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
  movie:  'linear-gradient(135deg, var(--espresso) 0%, #291c15 100%)',
  anime:  'linear-gradient(135deg, var(--mocha) 0%, var(--cinnamon) 100%)',
  series: 'linear-gradient(135deg, var(--espresso) 0%, var(--mocha) 100%)',
}

function posterStyle(rev) {
  return { background: posterGradients[rev.content_type] || posterGradients.movie }
}

const typeColors = { movie: 'var(--caramel)', anime: 'var(--cinnamon)', series: 'var(--mocha)' }
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

/* ══ Навбар ══ */
.rv-nav { border-radius: 0; border: none; border-bottom: 1px solid var(--border); transition: background-color 0.3s; }
.rv-nav-inner { padding: 11px 24px; display: flex; align-items: center; justify-content: space-between; }
.nav-sep { width: 1px; height: 20px; background: var(--border); }
.add-btn {
  display: flex; align-items: center; gap: 6px;
  padding: 8px 16px; border-radius: var(--radius-md);
  background: var(--primary); border: none; color: var(--latte-foam);
  font-size: 13px; font-weight: 600; cursor: pointer;
  transition: background-color 0.2s, transform 0.15s, box-shadow 0.2s;
  box-shadow: 0 2px 8px var(--primary-glow);
}
.add-btn:hover { background: var(--primary-hover); transform: translateY(-1px); box-shadow: 0 4px 12px var(--primary-glow); }

.badge-invites-count {
  background: var(--dusty-rose);
  color: var(--latte-foam);
  font-size: 10px;
  font-weight: 700;
  border-radius: 50%;
  width: 16px;
  height: 16px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
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
  background: var(--btn-ghost-hover-bg);
}
.nav-link-toggle.active {
  color: var(--text-primary);
  background: var(--btn-ghost-bg);
  border: 1px solid var(--border);
  font-weight: 600;
}

/* ══ Табы статусов ══ */
.status-tabs {
  display: flex;
  gap: 6px;
  padding: 8px 24px;
  border-bottom: 1px solid var(--border);
  background: var(--bg-surface);
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
  font-size: 13px; font-weight: 600;
  white-space: nowrap;
  transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
}
.status-tab:hover {
  background: var(--btn-ghost-hover-bg);
  color: var(--text-primary);
}
.status-tab.active {
  background: var(--bg-base);
  border-color: var(--primary);
  color: var(--primary);
  font-weight: 700;
  box-shadow: var(--shadow-sm);
}
.tab-icon { font-size: 15px; }
.tab-count {
  font-size: 11px; font-weight: 700;
  padding: 2px 8px; border-radius: 20px;
  color: var(--text-primary);
  min-width: 22px; text-align: center;
}

/* ══ Сплит-панели ══ */
.catalog-layout {
  display: grid;
  grid-template-columns: 280px 1fr;
  gap: 24px;
  padding: 24px;
  max-width: 1200px;
  margin: 0 auto;
  width: 100%;
  align-items: start;
}

/* Сайдбар */
.sidebar-filters {
  padding: 20px;
  border-radius: var(--radius-lg);
  border: 1px solid var(--border);
  background: var(--bg-surface);
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.sidebar-group {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.sidebar-label {
  font-size: 10px;
  font-weight: 700;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.search-input-wrapper {
  position: relative;
}

.search-input {
  padding: 9px 12px 9px 32px !important;
  font-size: 13px;
}

.search-indicator-icon {
  position: absolute;
  left: 10px;
  top: 50%;
  transform: translateY(-50%);
  font-size: 13px;
  color: var(--text-muted);
}

/* Фильтры в сайдбаре */
.stack-filters {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.stack-filter-btn {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  font-size: 13px;
  font-weight: 600;
  border-radius: var(--radius-md);
  background: var(--btn-ghost-bg);
  border: 1px solid var(--border);
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
  text-align: left;
}

.stack-filter-btn:hover {
  background: var(--btn-ghost-hover-bg);
  color: var(--text-primary);
  border-color: var(--border-hover);
}

.stack-filter-btn.active {
  background: rgba(192, 133, 82, 0.12);
  border-color: var(--primary);
  color: var(--primary);
}

.source-selector-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 4px;
}

.source-btn {
  padding: 6px 4px;
  font-size: 11px;
  font-weight: 600;
  border-radius: var(--radius-sm);
  background: var(--btn-ghost-bg);
  border: 1px solid var(--border);
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
  text-align: center;
}

.source-btn:hover {
  background: var(--btn-ghost-hover-bg);
  border-color: var(--border-hover);
  color: var(--text-primary);
}

.source-btn.active {
  background: var(--primary);
  border-color: var(--primary);
  color: var(--latte-foam);
}

.genres-multiselect {
  display: flex;
  flex-wrap: wrap;
  gap: 5px;
  max-h: 220px;
  overflow-y: auto;
  padding-right: 4px;
}

.genre-select-badge {
  padding: 3px 8px;
  font-size: 11px;
  font-weight: 500;
  border-radius: 6px;
  background: var(--btn-ghost-bg);
  border: 1px solid var(--border);
  color: var(--text-muted);
  cursor: pointer;
  transition: all 0.15s;
}

.genre-select-badge:hover {
  background: var(--btn-ghost-hover-bg);
  color: var(--text-primary);
}

.genre-select-badge.active {
  background: rgba(192, 133, 82, 0.15);
  border-color: var(--primary);
  color: var(--primary);
  font-weight: 600;
}

/* Правый фид каталога */
.catalog-feed {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.feed-header {
  padding: 12px 20px;
  border-radius: var(--radius-lg);
  border: 1px solid var(--border);
  background: var(--bg-surface);
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.feed-count-badge {
  font-family: var(--font-mono);
  font-size: 12px;
  font-weight: 700;
  color: var(--latte-foam);
  background: var(--primary);
  padding: 2px 8px;
  border-radius: 20px;
  margin-right: 8px;
}

.feed-count-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary);
}

.view-mode-toggle {
  display: flex;
  background: var(--btn-ghost-bg);
  padding: 3px;
  border-radius: 8px;
  border: 1px solid var(--border);
}

.mode-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 11px;
  font-weight: 700;
  padding: 4px 10px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
}

.mode-btn.active {
  background: var(--bg-surface);
  color: var(--primary);
  box-shadow: var(--shadow-sm);
}

/* ══ Сетка карточек (Grid Mode) ══ */
.card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 20px;
}

.rv-card {
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-lg);
  overflow: hidden;
  cursor: pointer;
  transition: transform 0.25s cubic-bezier(0.16, 1, 0.3, 1), border-color 0.25s, box-shadow 0.25s;
  box-shadow: var(--shadow-sm);
}
.rv-card:hover {
  transform: translateY(-4px) scale(1.01);
  border-color: var(--border-hover);
  box-shadow: var(--shadow-lg);
}

.card-poster {
  width: 100%;
  padding-bottom: 140%;
  position: relative;
  overflow: hidden;
  background: var(--bg-base);
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
  background: var(--bg-elevated);
  display: flex;
  align-items: center;
  justify-content: center;
}

.card-poster-overlay {
  position: absolute; inset: 0;
  background: linear-gradient(to top, rgba(20, 16, 13, 0.95) 0%, transparent 60%);
  pointer-events: none;
}

.card-type-badge {
  position: absolute;
  bottom: 8px;
  left: 8px;
  z-index: 1;
  padding: 3px 8px;
  border-radius: 20px;
  font-size: 10px;
  font-weight: 700;
  color: var(--latte-foam);
  box-shadow: var(--shadow-sm);
}

.shikimori-badge {
  position: absolute;
  bottom: 8px;
  right: 8px;
  background: rgba(31, 24, 19, 0.85);
  backdrop-filter: blur(4px);
  padding: 3px 6px;
  border-radius: 4px;
  color: var(--latte-foam);
  font-size: 10px;
  font-weight: 700;
  border: 1px solid var(--border);
}

.card-rating {
  position: absolute;
  top: 8px;
  right: 8px;
  background: rgba(31, 24, 19, 0.85);
  backdrop-filter: blur(8px);
  padding: 3px 6px;
  border-radius: 6px;
  color: var(--latte-foam);
  display: flex;
  align-items: center;
  gap: 2px;
  font-size: 11px;
  font-weight: 700;
  border: 1px solid var(--border);
}
.rating-star { color: var(--primary); }
.rating-max { color: var(--text-muted); font-size: 9px; }

.card-info {
  padding: 12px;
}

.card-title {
  font-size: 14px;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1.4;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.card-genres {
  margin: 6px 0;
}

.card-genre-pill {
  font-size: 9px;
  font-weight: 600;
  background: var(--btn-ghost-bg);
  border: 1px solid var(--border);
  color: var(--text-muted);
  padding: 1px 5px;
  border-radius: 4px;
}

.card-notes {
  font-size: 11.5px;
  color: var(--text-secondary);
  background: var(--form-bg);
  border-radius: 6px;
  padding: 6px 8px;
  line-height: 1.4;
  border-left: 2px solid var(--primary);
  margin-top: 6px;
}

.card-links {
  margin-top: 10px;
  border-top: 1px dashed var(--border);
  padding-top: 8px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.link-pill {
  font-size: 10px;
  color: var(--text-secondary);
  text-decoration: none;
  background: var(--btn-ghost-bg);
  border: 1px solid var(--border);
  padding: 3px 6px;
  border-radius: 4px;
  transition: all 0.2s;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.link-pill:hover {
  background: var(--btn-ghost-hover-bg);
  color: var(--primary);
  border-color: var(--primary);
}

.watch-together-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  padding: 4px 10px;
  border-radius: 6px;
  background: linear-gradient(135deg, rgba(192, 133, 82, 0.12) 0%, rgba(169, 103, 59, 0.12) 100%);
  color: var(--primary);
  border: 1px solid rgba(192, 133, 82, 0.35);
  font-size: 10px;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s;
  width: 100%;
}
.watch-together-btn:hover {
  background: linear-gradient(135deg, rgba(192, 133, 82, 0.25) 0%, rgba(169, 103, 59, 0.25) 100%);
  border-color: var(--primary);
}

.btn-increment-ep {
  background: rgba(192, 133, 82, 0.12);
  border: 1px solid rgba(192, 133, 82, 0.3);
  color: var(--primary);
  font-size: 9px;
  font-weight: 700;
  padding: 1px 5px;
  border-radius: 3px;
  cursor: pointer;
  transition: all 0.15s;
}
.btn-increment-ep:hover {
  background: var(--primary);
  color: var(--latte-foam);
}

/* ══ Список строк (List Mode) ══ */
.card-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.list-row {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px 16px;
  border-radius: var(--radius-md);
  border: 1px solid var(--border);
  background: var(--bg-surface);
  cursor: pointer;
  transition: all 0.25s ease;
}

.list-row:hover {
  transform: translateX(4px);
  border-color: var(--border-hover);
}

.list-poster-thumbnail {
  width: 44px;
  height: 60px;
  border-radius: 6px;
  background-size: cover;
  background-position: center;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.list-details {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.list-title {
  font-size: 15px;
  font-weight: 700;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.list-type-badge {
  padding: 2px 8px;
  border-radius: 20px;
  font-size: 9px;
  font-weight: 700;
  border: 1px solid transparent;
}

.list-shiki-badge {
  font-size: 10px;
  background: rgba(31, 24, 19, 0.08);
  color: var(--text-muted);
  border: 1px solid var(--border);
  padding: 1px 5px;
  border-radius: 4px;
}

.list-notes-text {
  font-size: 12px;
  color: var(--text-secondary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.list-genre-badge {
  font-size: 9px;
  background: var(--btn-ghost-bg);
  border: 1px solid var(--border);
  color: var(--text-muted);
  padding: 1px 4px;
  border-radius: 3px;
}

.list-stats-info {
  display: flex;
  align-items: center;
  gap: 20px;
  flex-shrink: 0;
}

.list-rating {
  font-size: 13px;
}

.list-episodes {
  width: 110px;
}

.list-progress-bar {
  width: 100%;
  height: 4px;
  background: var(--border);
  border-radius: 99px;
  overflow: hidden;
}

.list-progress-fill {
  height: 100%;
  background: var(--primary);
}

.list-row-actions {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.list-inc-btn {
  background: rgba(192, 133, 82, 0.12);
  border: 1px solid rgba(192, 133, 82, 0.3);
  color: var(--primary);
  font-size: 11px;
  font-weight: 700;
  padding: 6px 12px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all 0.2s;
}

.list-inc-btn:hover {
  background: var(--primary);
  color: var(--latte-foam);
}

/* ══ Модалки ══ */
.modal-overlay {
  position: fixed; inset: 0;
  background: rgba(20, 16, 13, 0.65); backdrop-filter: blur(12px);
  display: flex; align-items: center; justify-content: center;
  z-index: 1000; padding: 20px;
}

.modal-box {
  background: var(--bg-surface);
  border: 1px solid var(--border);
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-xl);
  width: 100%; max-width: 600px;
  max-height: 90vh;
  display: flex; flex-direction: column;
  overflow: hidden;
}

.modal-header {
  display: flex; align-items: center; justify-content: space-between;
  padding: 20px 24px;
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}

.modal-title { font-family: var(--font-display); font-size: 20px; font-weight: 700; color: var(--text-primary); }

.modal-close {
  width: 32px; height: 32px; border-radius: 50%;
  background: var(--btn-ghost-bg); border: 1px solid var(--border);
  color: var(--text-secondary); font-size: 14px;
  cursor: pointer; display: flex; align-items: center; justify-content: center;
  transition: all 0.15s;
}
.modal-close:hover { background: var(--btn-ghost-hover-bg); color: var(--text-primary); border-color: var(--border-hover); }

.modal-body {
  flex: 1; overflow-y: auto; padding: 24px;
}

.type-selector { display: flex; gap: 6px; }
.type-btn {
  flex: 1; padding: 8px; border-radius: var(--radius-md);
  background: var(--btn-ghost-bg); border: 1px solid var(--border);
  color: var(--text-secondary); cursor: pointer; font-size: 13px;
  transition: all 0.2s;
}
.type-btn:hover { background: var(--btn-ghost-hover-bg); color: var(--text-primary); }
.type-btn.active {
  background: rgba(192, 133, 82, 0.15); border-color: var(--primary);
  color: var(--primary); font-weight: 600;
}

.status-selector {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(90px, 1fr));
  gap: 4px;
}

.status-opt {
  padding: 6px 4px;
  font-size: 10px;
  font-weight: 600;
  border-radius: var(--radius-md);
  background: var(--btn-ghost-bg);
  border: 1px solid var(--border);
  color: var(--text-secondary);
  cursor: pointer;
  transition: all 0.2s;
  text-align: center;
}
.status-opt:hover {
  background: var(--btn-ghost-hover-bg);
}

.rating-row {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.rating-btn {
  width: 26px;
  height: 26px;
  border-radius: var(--radius-sm);
  background: var(--btn-ghost-bg);
  border: 1px solid var(--border);
  color: var(--text-secondary);
  font-size: 11px;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.15s;
  display: flex;
  align-items: center;
  justify-content: center;
}

.rating-btn:hover {
  border-color: var(--primary);
  color: var(--primary);
}

.rating-btn.active {
  background: var(--primary);
  border-color: var(--primary);
  color: var(--latte-foam);
}

.rating-btn.clear-btn {
  background: transparent;
  border: 1px solid var(--border);
  color: var(--dusty-rose);
}
.rating-btn.clear-btn:hover {
  background: rgba(201, 112, 100, 0.1);
}

.poster-preview {
  margin-top: 10px;
  border-radius: var(--radius-md);
  overflow: hidden;
  max-height: 200px;
  display: flex;
  justify-content: center;
  border: 1px solid var(--border);
  background: var(--bg-base);
}

.poster-preview img {
  height: 100%;
  object-fit: contain;
}

.poster-error {
  padding: 20px;
  font-size: 12px;
  color: var(--dusty-rose);
}

.form-textarea {
  resize: vertical;
}

.genre-pill {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 4px 10px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
  background: rgba(192, 133, 82, 0.1);
  border: 1px solid rgba(192, 133, 82, 0.2);
  color: var(--primary);
}

.genre-del-btn {
  background: transparent;
  border: none;
  color: var(--dusty-rose);
  font-size: 13px;
  cursor: pointer;
  padding: 0;
}

.quick-genre-btn {
  background: var(--btn-ghost-bg);
  border: 1px solid var(--border);
  color: var(--text-muted);
  font-size: 11px;
  font-weight: 500;
  padding: 4px 10px;
  border-radius: var(--radius-sm);
  cursor: pointer;
  transition: all 0.15s;
}

.quick-genre-btn:hover {
  background: var(--btn-ghost-hover-bg);
}

.quick-genre-btn.active {
  background: var(--primary);
  color: var(--latte-foam);
  border-color: var(--primary);
}

.existing-links {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-bottom: 10px;
}

.existing-link {
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: var(--form-bg);
  padding: 6px 12px;
  border-radius: var(--radius-md);
  border: 1px solid var(--border);
}

.link-info {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
}

.link-label-text {
  font-size: 11px;
  font-weight: 700;
  color: var(--primary);
}

.link-url-text {
  font-size: 12px;
  color: var(--text-secondary);
  text-decoration: none;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.link-del-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  font-size: 12px;
}
.link-del-btn:hover {
  color: var(--dusty-rose);
}

.new-link-row {
  display: flex;
  gap: 6px;
  margin-bottom: 6px;
}

.add-link-btn {
  background: var(--btn-ghost-bg);
  border: 1px dashed var(--border);
  color: var(--primary);
  font-size: 12px;
  font-weight: 600;
  padding: 8px;
  width: 100%;
  border-radius: var(--radius-md);
  cursor: pointer;
  transition: all 0.2s;
}

.add-link-btn:hover {
  background: var(--btn-ghost-hover-bg);
  border-color: var(--primary);
}

.modal-footer {
  display: flex;
  align-items: center;
  padding: 16px 24px;
  border-top: 1px solid var(--border);
  flex-shrink: 0;
}

.confirm-box {
  width: 100%; max-width: 380px; padding: 24px;
  border-radius: var(--radius-lg); text-align: center;
  border: 1px solid var(--border);
  background: var(--bg-surface);
  box-shadow: var(--shadow-xl);
}

/* ══ Анимации модалок ══ */
.modal-enter-active,
.modal-leave-active {
  transition: opacity 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}
.modal-enter-active .modal-box,
.modal-enter-active .confirm-box,
.modal-leave-active .modal-box,
.modal-leave-active .confirm-box {
  transition: transform 0.25s cubic-bezier(0.16, 1, 0.3, 1), opacity 0.25s ease;
}

.modal-enter-from {
  opacity: 0;
}
.modal-enter-from .modal-box,
.modal-enter-from .confirm-box {
  opacity: 0;
  transform: scale(0.95);
}

.modal-leave-to {
  opacity: 0;
}
.modal-leave-to .modal-box,
.modal-leave-to .confirm-box {
  opacity: 0;
  transform: scale(0.95);
}

/* Адаптив */
@media (max-width: 992px) {
  .catalog-layout {
    grid-template-columns: 1fr;
  }
}
</style>
