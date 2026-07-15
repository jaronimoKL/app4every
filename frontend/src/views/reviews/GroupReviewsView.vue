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
      
      <!-- Оверлей для мобильного сайдбара групп -->
      <div v-if="showMobileGroups" class="groups-backdrop xl:hidden" @click="showMobileGroups = false"></div>

      <!-- ══ БОКОВАЯ ПАНЕЛЬ (СПИСОК ГРУПП) ══ -->
      <aside class="sidebar glass" :class="{ 'is-mobile-open': showMobileGroups }">
        <div class="sidebar-header">
          <h3>Мои Группы</h3>
          <button class="drawer-close-btn xl:hidden" @click="showMobileGroups = false" style="background:none; border:none; color:var(--text-secondary); cursor:pointer; font-size:16px;">✕</button>
          <button v-if="!showMobileGroups" class="icon-btn" @click="openCreateGroup" title="Создать группу">＋</button>
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
          <!-- Активные комнаты -->
          <div v-if="activeRooms.length > 0" class="active-rooms-sidebar mb-4">
            <div class="info-divider"></div>
            <h4>📺 Сейчас смотрят</h4>
            <div class="rooms-list mt-2 flex flex-col gap-2">
              <div 
                v-for="room in activeRooms" 
                :key="room.room_id" 
                class="sidebar-room-card glass p-2 rounded cursor-pointer hover:bg-white/5 transition-colors" 
                @click="$router.push(`/watch/room/${room.room_id}`)"
              >
                <div class="text-xs font-semibold flex items-center gap-1.5 text-green-400">
                  <span class="online-dot-small"></span> Комната {{ room.room_id.substring(0, 4) }}
                </div>
                <div class="text-[10px] text-gray-400 mt-1">
                  Участников: {{ room.participants?.length || 1 }}
                </div>
              </div>
            </div>
          </div>

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
            <div class="flex items-center gap-2">
              <button class="mobile-toggle-btn xl:hidden" @click="showMobileGroups = true">
                👥 Группы
              </button>
              <button class="mobile-toggle-btn xl:hidden" @click="showMobileFilters = true">
                🔍 Фильтры
              </button>
              <span class="online-indicator flex items-center gap-1.5 text-xs text-green-400">
                <span class="ping-dot"></span> в сети (Live)
              </span>
            </div>
          </div>

          <!-- ══ ТАБЫ СТАТУСОВ ══ -->
          <div class="status-tabs">
            <div class="status-tabs-inner">
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
          </div>

          <!-- ══ ОСНОВНОЙ КОНТЕНТ (SPLIT PANE CATALOG) ══ -->
          <div class="catalog-layout">
            
            <!-- Оверлей мобильных фильтров -->
            <div v-if="showMobileFilters" class="filters-backdrop xl:hidden" @click="showMobileFilters = false"></div>

            <!-- Левый Сайдбар Фильтров -->
            <aside class="sidebar-filters glass" :class="{ 'is-mobile-open': showMobileFilters }">
              <!-- Drawer close header (only visible on mobile/tablet) -->
              <div class="drawer-header xl:hidden flex justify-between items-center mb-4">
                <span class="drawer-title">Фильтры</span>
                <button class="drawer-close-btn" @click="showMobileFilters = false">✕</button>
              </div>

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

              <!-- Кнопка сброса -->
              <button
                v-if="searchQuery || selectedContentType !== 'all'"
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
                  <span class="feed-count-badge">{{ filteredItems.length }}</span>
                  <span class="feed-count-label">найдено в совместном списке</span>
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

              <!-- Пустой список -->
              <div v-if="filteredItems.length === 0" class="empty-state">
                <div class="empty-state-icon">🔍</div>
                <h3 style="font-weight:700;font-size:16px;margin-bottom:6px;">Ничего не найдено</h3>
                <p style="font-size:13px;color:var(--text-muted);margin-bottom:16px;">Попробуйте изменить параметры поиска или добавить новые фильмы</p>
              </div>

              <!-- Сетка карточек (Grid View) -->
              <div v-else-if="viewMode === 'grid'" class="card-grid">
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
                    <div class="card-episode-progress mt-1.5 flex items-center gap-1.5 text-xs font-semibold" style="color: var(--primary);">
                      <span>🎬</span> Серия: {{ item.current_episode }} из {{ item.max_episodes }}
                      <button 
                        v-if="item.added_by === authStore.user?.id && item.current_episode < item.max_episodes" 
                        class="btn-increment-ep"
                        @click="incrementEpisode(item)"
                        title="Просмотрена еще 1 серия"
                      >
                        ＋1
                      </button>
                    </div>

                    <!-- Жанры -->
                    <div class="card-genres flex flex-wrap gap-1 mb-2 mt-1" v-if="item.genres && item.genres.length > 0">
                      <span v-for="g in item.genres.slice(0, 2)" :key="g" class="card-genre-pill">
                        {{ g }}
                      </span>
                      <span v-if="item.genres.length > 2" class="card-genre-pill badge-more">
                        +{{ item.genres.length - 2 }}
                      </span>
                    </div>

                    <!-- Быстрая смена статуса (только для создателя) -->
                    <div v-if="item.added_by === authStore.user?.id" class="quick-status-change flex items-center gap-1 mt-2.5">
                      <span style="font-size:11px;color:var(--text-secondary);margin-right:4px;">Статус:</span>
                      <button 
                        v-for="opt in statusOptions" 
                        :key="opt.value"
                        class="quick-status-btn"
                        :class="{ active: item.status === opt.value }"
                        @click="changeItemStatus(item, opt.value)"
                        :title="opt.label"
                      >
                        {{ opt.icon }}
                      </button>
                    </div>

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
                      
                      <template v-if="getActiveRoomForItem(item)">
                        <button
                          class="watch-together-btn mt-2 flex items-center justify-center gap-1.5 bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 hover:bg-emerald-500/20 transition-all rounded-md px-3 py-1.5 text-xs font-semibold"
                          @click.stop="$router.push(`/watch/room/${getActiveRoomForItem(item).room_id}`)"
                        >
                          <span class="online-dot-small" style="display:inline-block; vertical-align:middle; margin-bottom:1px;"></span>
                          Присоединиться к просмотру
                        </button>
                      </template>
                      <template v-else-if="item.shikimori_id || item.aniliberty_alias || (item.links && item.links.length > 0)">
                        <button
                          class="watch-together-btn mt-2 flex items-center justify-center gap-1.5 bg-indigo-500/10 text-indigo-400 border border-indigo-500/30 hover:bg-indigo-500/20 transition-all rounded-md px-3 py-1.5 text-xs font-semibold"
                          @click.stop="handleWatchTogether(item)"
                        >
                          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" class="w-3.5 h-3.5">
                            <polygon points="5 3 19 12 5 21 5 3" fill="currentColor"/>
                          </svg>
                          Смотреть вместе
                        </button>
                      </template>
                    </div>
                  </div>
                </div>
              </div>

              <!-- Список строк (List View) -->
              <div v-else-if="viewMode === 'list'" class="card-list">
                <div 
                  v-for="item in filteredItems" 
                  :key="item.id"
                  class="list-row glass glass-hover"
                >
                  <!-- Миниатюрный постер -->
                  <div class="list-poster-thumbnail" :style="posterStyle(item)">
                    <span class="text-sm">{{ typeIcon(item.content_type) }}</span>
                  </div>

                  <!-- Детали и метаданные -->
                  <div class="list-details">
                    <div class="flex items-center gap-2 flex-wrap">
                      <h4 class="list-title">{{ item.title }}</h4>
                      <span class="list-type-badge" :style="{ color: typeColor(item.content_type), borderColor: typeColor(item.content_type) + '30', background: typeColor(item.content_type) + '10' }">
                        {{ typeIcon(item.content_type) }} {{ typeLabel(item.content_type) }}
                      </span>
                      <span class="text-[11px] text-gray-400">Добавил: {{ item.added_by_username }}</span>
                    </div>
                    <p class="list-notes-text" v-if="item.notes">{{ item.notes }}</p>
                    <div class="list-genres flex flex-wrap gap-1 mt-1" v-if="item.genres && item.genres.length > 0">
                      <span v-for="g in item.genres" :key="g" class="list-genre-badge">{{ g }}</span>
                    </div>
                  </div>

                  <!-- Средние оценки и эпизоды -->
                  <div class="list-stats-info">
                    <div class="list-rating tooltip-ratings" v-if="item.average_rating > 0">
                      <span style="color: var(--primary);">★</span>
                      <span style="font-weight: 700;">{{ item.average_rating.toFixed(1) }}</span>
                      <!-- Тултип с голосами -->
                      <span class="tooltip-text glass">
                        <div style="font-weight:700; margin-bottom:4px; font-size:11px;">Оценки участников:</div>
                        <div v-for="r in item.ratings" :key="r.id" style="font-size:11px;">
                          {{ r.username }}: <span style="color:#fbbf24; font-weight:bold;">{{ r.rating }}★</span>
                        </div>
                      </span>
                    </div>
                    
                    <div class="list-episodes" v-if="item.max_episodes && item.content_type !== 'movie'">
                      <div class="text-xs font-semibold mb-1" style="color: var(--text-secondary);">
                        Серии: {{ item.current_episode || 0 }} / {{ item.max_episodes }}
                      </div>
                      <div class="list-progress-bar">
                        <div class="list-progress-fill" :style="{ width: ((item.current_episode || 0) / item.max_episodes * 100) + '%' }"></div>
                      </div>
                    </div>
                  </div>

                  <!-- Действия -->
                  <div class="list-row-actions" @click.stop>
                    <button 
                      class="list-inc-btn"
                      v-if="item.added_by === authStore.user?.id && item.current_episode < item.max_episodes" 
                      @click="incrementEpisode(item)"
                      title="Плюс одна серия"
                    >
                      ＋1 серию
                    </button>
                    
                    <template v-if="getActiveRoomForItem(item)">
                      <button
                        class="watch-together-btn btn-sm bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 hover:bg-emerald-500/20"
                        @click.stop="$router.push(`/watch/room/${getActiveRoomForItem(item).room_id}`)"
                      >
                        Присоединиться
                      </button>
                    </template>
                    <template v-else-if="item.shikimori_id || item.aniliberty_alias || (item.links && item.links.length > 0)">
                      <button
                        class="watch-together-btn btn-sm"
                        @click="handleWatchTogether(item)"
                      >
                        Смотреть вместе
                      </button>
                    </template>

                    <div class="edit-actions" v-if="item.added_by === authStore.user?.id">
                      <button class="small-icon-btn" @click="openEditItem(item)">✏️</button>
                      <button class="small-icon-btn text-red-400" @click="deleteItemConfirm(item)">🗑</button>
                    </div>
                  </div>
                </div>
              </div>

            </div>

          </div>
        </div>
      </main>
    </div>

    <!-- ══ МОДАЛЬНОЕ ОКНО: Входящие приглашения ══ -->
    <Transition name="modal">
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
                  Группа <span style="font-weight:700; color:var(--text-primary);">«{{ inv.group_name }}»</span>
                </div>
                <div style="font-size:12px; color: var(--text-muted);">
                  Пригласил: <span style="color:var(--text-secondary);">{{ inv.inviter_username }}</span>
                </div>
                <div class="flex gap-2 justify-end mt-1">
                  <button class="btn btn-ghost btn-xs" style="color: var(--dusty-rose); border-color: rgba(201, 112, 100, 0.2);" @click="declineInvite(inv.id)">Отклонить</button>
                  <button class="btn btn-primary btn-xs" @click="acceptInvite(inv.id)">Принять</button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </Transition>

    <!-- ══ МОДАЛЬНОЕ ОКНО: Создание группы ══ -->
    <Transition name="modal">
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
    </Transition>

    <!-- ══ МОДАЛЬНОЕ ОКНО: Добавить / Редактировать запись в группе ══ -->
    <Transition name="modal">
      <div v-if="showItemModal" class="modal-overlay" @click.self="closeItemModal">
        <div class="modal-box glass" @click.stop>
          <div class="modal-header">
            <h3 class="modal-title">{{ isEditingItem ? 'Редактировать запись' : 'Добавить запись' }}</h3>
            <button class="modal-close" @click="closeItemModal">✕</button>
          </div>
          <div class="modal-body" style="padding:20px;">
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
                  :class="{ active: itemForm.content_type === t.value }"
                  @click="itemForm.content_type = t.value"
                >{{ t.icon }} {{ t.label }}</button>
              </div>
            </div>

            <!-- Кнопка поиска для аниме -->
            <div v-if="itemForm.content_type === 'anime' && !isEditingItem" class="mt-4 mb-2">
              <button class="btn btn-primary w-full text-sm py-2" @click="showAnimeSearch = true">
                🔍 Найти на Shikimori (Автозаполнение)
              </button>
            </div>
            <div v-if="(itemForm.content_type === 'movie' || itemForm.content_type === 'series') && !isEditingItem" class="mt-4 mb-2">
              <button class="btn btn-primary w-full text-sm py-2" @click="showMovieSearch = true">
                🔍 Найти на TMDB (Автозаполнение)
              </button>
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
              <button class="btn btn-primary" :disabled="!itemForm.title || isSaving" @click="saveItem">{{ isSaving ? 'Сохранение...' : 'Сохранить' }}</button>
            </div>
            </template>
          </div>
        </div>
      </div>
    </Transition>

    <!-- Диалоги подтверждения -->
    <Transition name="modal">
      <div v-if="showDeleteConfirm" class="modal-overlay" @click.self="showDeleteConfirm = false">
        <div class="confirm-box glass" @click.stop>
          <div style="font-size:32px;margin-bottom:12px;">🗑</div>
          <h3 style="font-weight:700;margin-bottom:8px;">Удалить запись?</h3>
          <p style="font-size:13px;color:var(--text-secondary);margin-bottom:20px;">
            Запись «{{ itemToDelete?.title }}» будет удалена из группового списка.
          </p>
          <div class="flex items-center gap-3" style="justify-content:flex-end;">
            <button class="btn btn-ghost" @click="showDeleteConfirm = false">Отмена</button>
            <button class="btn btn-primary" style="background:var(--dusty-rose);" @click="confirmDeleteItem">Удалить</button>
          </div>
        </div>
      </div>
    </Transition>

    <Transition name="modal">
      <div v-if="showDeleteGroupConfirm" class="modal-overlay" @click.self="showDeleteGroupConfirm = false">
        <div class="confirm-box glass" @click.stop>
          <div style="font-size:32px;margin-bottom:12px;">⚠️</div>
          <h3 style="font-weight:700;margin-bottom:8px;">Удалить группу?</h3>
          <p style="font-size:13px;color:var(--text-secondary);margin-bottom:20px;">
            Группа «{{ activeGroup?.name }}» и весь совместный список будут безвозвратно удалены.
          </p>
          <div class="flex items-center gap-3" style="justify-content:flex-end;">
            <button class="btn btn-ghost" @click="showDeleteGroupConfirm = false">Отмена</button>
            <button class="btn btn-primary" style="background:var(--dusty-rose);" @click="confirmDeleteGroup">Удалить</button>
          </div>
        </div>
      </div>
    </Transition>

    <Transition name="modal">
      <div v-if="showLeaveConfirm" class="modal-overlay" @click.self="showLeaveConfirm = false">
        <div class="confirm-box glass" @click.stop>
          <div style="font-size:32px;margin-bottom:12px;">🚪</div>
          <h3 style="font-weight:700;margin-bottom:8px;">Выйти из группы?</h3>
          <p style="font-size:13px;color:var(--text-secondary);margin-bottom:20px;">
            Вы больше не сможете просматривать совместный список группы «{{ activeGroup?.name }}».
          </p>
          <div class="flex items-center gap-3" style="justify-content:flex-end;">
            <button class="btn btn-ghost" @click="showLeaveConfirm = false">Отмена</button>
            <button class="btn btn-primary" style="background:var(--dusty-rose);" @click="confirmLeaveGroup">Выйти</button>
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
import { ref, computed, onMounted, onUnmounted, reactive, watch } from 'vue'
import { useGroupsStore } from '@/stores/groups'
import { useAuthStore } from '@/stores/auth'
import { friendsApi, watchpartyApi } from '@/services/api'
import AnimeSearchStep from '@/components/reviews/AnimeSearchStep.vue'
import MovieSearchStep from '@/components/reviews/MovieSearchStep.vue'
import EpisodePickerModal from '@/components/reviews/EpisodePickerModal.vue'
import { useRouter } from 'vue-router'

const groupsStore = useGroupsStore()
const authStore = useAuthStore()
const router = useRouter()

// ── Состояние ──
const activeGroupId = ref(null)
const activeTab = ref('watching')
const searchQuery = ref('')
const selectedContentType = ref('all')
const viewMode = ref(localStorage.getItem('group-reviews-view-mode') || 'grid')
const showMobileGroups = ref(false)
const showMobileFilters = ref(false)
function toggleViewMode(mode) {
  viewMode.value = mode
  localStorage.setItem('group-reviews-view-mode', mode)
}

const friends = ref([])
const activeRooms = ref([])
let roomsPollingInterval = null

async function fetchGroupActiveRooms() {
  if (!activeGroup.value || !activeGroup.value.members) return
  try {
    const userIds = activeGroup.value.members.map(m => m.user_id)
    if (userIds.length > 0) {
      const rooms = await watchpartyApi.getActiveRooms(userIds)
      activeRooms.value = rooms || []
    }
  } catch (err) {
    console.error('Failed to fetch active rooms for group:', err)
  }
}

function startRoomsPolling() {
  stopRoomsPolling()
  roomsPollingInterval = setInterval(fetchGroupActiveRooms, 10000)
}

function stopRoomsPolling() {
  if (roomsPollingInterval) {
    clearInterval(roomsPollingInterval)
    roomsPollingInterval = null
  }
}

function getActiveRoomForItem(item) {
  if (!item || !activeRooms.value.length) return null
  return activeRooms.value.find(r => 
    (item.shikimori_id && r.shikimori_id === String(item.shikimori_id)) || 
    (item.aniliberty_alias && r.aniliberty_alias === item.aniliberty_alias)
  )
}

// Модалки
const showInvitesModal = ref(false)
const showCreateGroupModal = ref(false)
const showItemModal = ref(false)
const isSaving = ref(false)
const showDeleteConfirm = ref(false)
const showDeleteGroupConfirm = ref(false)
const showLeaveConfirm = ref(false)
const showAnimeSearch = ref(false)
const showMovieSearch = ref(false)

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
  notes: '',
  shikimori_id: null,
  description: '',
  episodes_total: null,
  aniliberty_alias: '',
  shikimori_score: null,
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
  { key: 'watching',  icon: '📺', label: 'Смотрю', countBg: 'rgba(192, 133, 82, 0.25)' },
  { key: 'completed', icon: '✅', label: 'Просмотрено', countBg: 'rgba(124, 154, 110, 0.25)' },
  { key: 'planned',   icon: '📋', label: 'Запланировано', countBg: 'rgba(192, 133, 82, 0.15)' },
  { key: 'dropped',   icon: '⛔', label: 'Брошено', countBg: 'rgba(201, 112, 100, 0.25)' },
]

const contentTypes = [
  { value: 'movie',  icon: '🎬', label: 'Фильм'  },
  { value: 'anime',  icon: '✨', label: 'Аниме'  },
  { value: 'series', icon: '📺', label: 'Сериал' },
]

const statusOptions = [
  { value: 'watching',  icon: '📺', label: 'Смотрю' },
  { value: 'completed', icon: '✅', label: 'Просмотрено' },
  { value: 'planned',   icon: '📋', label: 'Запланировано' },
  { value: 'dropped',   icon: '⛔', label: 'Брошено' },
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
  stopRoomsPolling()
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
    fetchGroupActiveRooms()
    startRoomsPolling()
  } else {
    groupsStore.disconnectWS()
    stopRoomsPolling()
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
  itemForm.shikimori_id = null
  itemForm.description = ''
  itemForm.episodes_total = null
  itemForm.aniliberty_alias = ''
  itemForm.shikimori_score = null
  showAnimeSearch.value = false
  showMovieSearch.value = false
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
  itemForm.shikimori_id = item.shikimori_id || null
  itemForm.description = item.description || ''
  itemForm.episodes_total = item.episodes_total || null
  itemForm.aniliberty_alias = item.aniliberty_alias || ''
  itemForm.shikimori_score = item.shikimori_score || null
  showAnimeSearch.value = false
  showMovieSearch.value = false
  showItemModal.value = true
}

function handleAnimeSelect(anime) {
  itemForm.title = anime.title
  itemForm.poster_url = anime.posterFull || anime.poster
  itemForm.shikimori_id = anime.id
  itemForm.shikimori_score = anime.score
  
  if (anime.details) {
    itemForm.description = anime.details.description || ''
    // Если сериал онгоинг, то episodes может быть 0
    const totalEps = anime.details.episodes || anime.details.episodes_total || anime.details.episodes_aired || 1
    if (totalEps) {
      itemForm.max_episodes = totalEps
      itemForm.episodes_total = totalEps
    }
    if (anime.details.genres) {
      itemForm.genresString = anime.details.genres.map(g => g.russian).join(', ')
    }
  }

  if (anime.anilibertyRelease) {
    itemForm.aniliberty_alias = anime.anilibertyRelease.alias
  }

  showAnimeSearch.value = false
}

function handleMovieSelect(item) {
  itemForm.title = item.title
  itemForm.description = item.overview || ''
  
  if (item.poster_url) {
    itemForm.poster_url = item.poster_url
  }
  
  itemForm.tmdb_id = item.id
  itemForm.rating = item.vote_average ? Math.round(item.vote_average) : null

  if (item.media_type === 'movie') {
    itemForm.content_type = 'movie'
    itemForm.episodes_total = 1
    itemForm.max_episodes = 1
  } else if (item.media_type === 'tv') {
    itemForm.content_type = 'series'
    if (item.details && item.details.number_of_episodes) {
      itemForm.episodes_total = item.details.number_of_episodes
      itemForm.max_episodes = item.details.number_of_episodes
    }
  }

  if (item.details && item.details.genres && item.details.genres.length > 0) {
    itemForm.genres = item.details.genres.map(g => ({ name: g.name }))
  }

  showMovieSearch.value = false
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
  if (!itemForm.title || !activeGroup.value || isSaving.value) return
  isSaving.value = true
  
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
    notes: itemForm.notes,
    shikimori_id: itemForm.shikimori_id,
    description: itemForm.description,
    episodes_total: itemForm.episodes_total,
    aniliberty_alias: itemForm.aniliberty_alias,
    shikimori_score: itemForm.shikimori_score
  }

  try {
    if (isEditingItem.value) {
      await groupsStore.updateGroupItem(activeGroup.value.id, itemForm.id, payload)
    } else {
      await groupsStore.addGroupItem(activeGroup.value.id, payload)
    }
    // Автоматически переключаем таб, чтобы показать новую запись
    activeTab.value = itemForm.status
    closeItemModal()
  } catch (err) {
    console.error(err)
  } finally {
    isSaving.value = false
  }
}

function deleteItemConfirm(item) {
  itemToDelete.value = item
  showDeleteConfirm.value = true
}

const activeAlias = ref(null)
const activeShikimoriId = ref(null)
const showEpisodePicker = ref(false)

function handleWatchTogether(item) {
  activeShikimoriId.value = item.shikimori_id || null
  activeAlias.value = item.aniliberty_alias || null
  if (item.shikimori_id || item.aniliberty_alias) {
    showEpisodePicker.value = true
  } else if (item.links && item.links.length > 0) {
    openWatchParty(item.links[0].url)
  }
}

function onEpisodeSelect(url) {
  showEpisodePicker.value = false
  openWatchParty(url, activeShikimoriId.value, activeAlias.value)
}

function generateUUID() {
  return Math.random().toString(36).substring(2, 10)
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

async function incrementEpisode(item) {
  if (item.current_episode >= item.max_episodes) return
  const newEpisode = item.current_episode + 1
  
  const payload = {
    title:            item.title,
    content_type:     item.content_type,
    status:           newEpisode === item.max_episodes ? 'completed' : item.status,
    current_episode:  newEpisode,
    max_episodes:     item.max_episodes,
    poster_url:       item.poster_url,
    genres:           item.genres || [],
    notes:            item.notes,
    shikimori_id:     item.shikimori_id || null,
    description:      item.description || '',
    episodes_total:   item.episodes_total || null,
    aniliberty_alias: item.aniliberty_alias || '',
    shikimori_score:  item.shikimori_score || null
  }
  
  try {
    await groupsStore.updateGroupItem(activeGroup.value.id, item.id, payload)
  } catch (err) {
    console.error(err)
  }
}

async function changeItemStatus(item, newStatus) {
  if (item.status === newStatus) return
  
  const payload = {
    title:            item.title,
    content_type:     item.content_type,
    status:           newStatus,
    current_episode:  newStatus === 'completed' ? item.max_episodes : item.current_episode,
    max_episodes:     item.max_episodes,
    poster_url:       item.poster_url,
    genres:           item.genres || [],
    notes:            item.notes,
    shikimori_id:     item.shikimori_id || null,
    description:      item.description || '',
    episodes_total:   item.episodes_total || null,
    aniliberty_alias: item.aniliberty_alias || '',
    shikimori_score:  item.shikimori_score || null
  }
  
  try {
    await groupsStore.updateGroupItem(activeGroup.value.id, item.id, payload)
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
    background: 'linear-gradient(135deg, var(--espresso) 0%, #291c15 100%)',
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

/* ══ Навбар ══ */
.rv-nav { border-radius: 0; border: none; border-bottom: 1px solid var(--border); transition: background-color 0.3s; }
.rv-nav-inner { padding: 11px 24px; display: flex; align-items: center; justify-content: space-between; max-width: 1200px; margin: 0 auto; width: 100%; }
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

/* ══ Лайаут ══ */
.main-layout {
  flex: 1;
  display: flex;
  overflow: hidden;
}

/* ══ Боковая панель (Группы) ══ */
.sidebar {
  width: 280px;
  flex-shrink: 0;
  border-right: 1px solid var(--border);
  background: var(--bg-surface);
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
  font-size: 11px;
  font-weight: 700;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}
.icon-btn {
  background: var(--btn-ghost-bg);
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
  transition: all 0.15s;
}
.icon-btn:hover {
  background: var(--btn-ghost-hover-bg);
  border-color: var(--border-hover);
}

.groups-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  overflow-y: auto;
  flex: 1;
}

.group-item {
  padding: 12px;
  border-radius: var(--radius-md);
  background: var(--btn-ghost-bg);
  border: 1px solid var(--border);
  cursor: pointer;
  transition: all 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}
.group-item:hover {
  background: var(--btn-ghost-hover-bg);
  border-color: var(--border-hover);
}
.group-item.active {
  background: rgba(192, 133, 82, 0.1);
  border-color: var(--primary);
}
.group-name {
  font-size: 14px;
  font-weight: 700;
  color: var(--text-primary);
}
.group-meta {
  font-size: 11px;
  color: var(--text-muted);
  margin-top: 4px;
}

.active-group-sidebar-info {
  margin-top: 16px;
  border-top: 1px dashed var(--border);
  padding-top: 16px;
}

.info-divider {
  height: 1px;
  background: var(--border);
  margin: 12px 0;
}

.sidebar-room-card {
  background: rgba(124, 154, 110, 0.08);
  border: 1px solid rgba(124, 154, 110, 0.25);
  border-radius: var(--radius-md);
  padding: 8px;
}
.sidebar-room-card:hover {
  background: rgba(124, 154, 110, 0.15);
}

.sidebar-subheader {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}
.sidebar-subheader h4 {
  font-size: 11px;
  font-weight: 700;
  color: var(--text-muted);
  text-transform: uppercase;
}

.leave-btn {
  background: transparent;
  border: 1px solid rgba(201, 112, 100, 0.3);
  color: var(--dusty-rose);
  font-size: 10px;
  font-weight: 700;
  padding: 2px 6px;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
}
.leave-btn:hover {
  background: var(--dusty-rose);
  color: var(--latte-foam);
  border-color: var(--dusty-rose);
}

.members-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  margin-top: 8px;
}
.member-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--text-secondary);
}
.member-dot {
  width: 6px;
  height: 6px;
  background: var(--matcha);
  border-radius: 50%;
}
.owner-badge {
  font-size: 9px;
  font-weight: 700;
  background: rgba(192, 133, 82, 0.15);
  color: var(--primary);
  padding: 1px 4px;
  border-radius: 4px;
  margin-left: auto;
}

.invite-section h4 {
  font-size: 11px;
  font-weight: 700;
  color: var(--text-muted);
  text-transform: uppercase;
}

/* ══ Контентная область ══ */
.content-area {
  flex: 1;
  overflow-y: auto;
  background: var(--bg-base);
  display: flex;
  flex-direction: column;
}

.no-active-group {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 40px;
}
.empty-bubble {
  max-width: 460px;
  text-align: center;
  padding: 32px;
  border-radius: var(--radius-xl);
  border: 1px solid var(--border);
  background: var(--bg-surface);
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
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.5;
}

.watchlist-area {
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.watchlist-header h2 {
  font-family: var(--font-display);
  font-size: 24px;
  font-weight: 800;
  color: var(--text-primary);
}

/* ══ Табы статусов ══ */
.status-tabs {
  border-bottom: 1px solid var(--border);
  overflow-x: auto;
  scrollbar-width: none;
}
.status-tabs-inner {
  display: flex;
  gap: 6px;
  padding: 8px 0;
  max-width: 1200px;
  margin: 0 auto;
  width: 100%;
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
  background: var(--bg-surface);
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

/* ══ Сплит-панель внутри watchlist ══ */
.catalog-layout {
  display: grid;
  grid-template-columns: 240px 1fr;
  gap: 24px;
  width: 100%;
  align-items: start;
  margin-top: 8px;
}

/* Сайдбар фильтров */
.sidebar-filters {
  padding: 16px;
  border-radius: var(--radius-lg);
  border: 1px solid var(--border);
  background: var(--bg-surface);
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.sidebar-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
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

.stack-filters {
  display: flex;
  flex-direction: column;
  gap: 5px;
}

.stack-filter-btn {
  display: flex;
  align-items: center;
  padding: 8px 12px;
  font-size: 12.5px;
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

/* Фид медиа */
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

/* ══ Сетка карточек ══ */
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
  display: flex;
  flex-direction: column;
  height: 100%;
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
  aspect-ratio: 2 / 3;
  position: relative;
  overflow: hidden;
  background: var(--bg-base);
  flex-shrink: 0;
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

.card-info {
  padding: 12px;
  flex: 1;
  display: flex;
  flex-direction: column;
}

.card-title {
  font-size: 14px;
  font-weight: 700;
  color: var(--text-primary);
  line-height: 1.3;
  height: 2.6em;
  margin-bottom: 4px;
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
}

.edit-actions {
  display: flex;
  gap: 4px;
}

.small-icon-btn {
  background: transparent;
  border: none;
  font-size: 11px;
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
  margin-bottom: 4px;
}

.card-episode-progress {
  margin: 6px 0;
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
  margin-left: auto;
}
.btn-increment-ep:hover {
  background: var(--primary);
  color: var(--latte-foam);
}

.card-genres {
  margin-bottom: 6px;
  height: 18px;
  overflow: hidden;
  display: flex;
  flex-wrap: wrap;
  gap: 4px;
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

.card-genre-pill.badge-more {
  background: rgba(192, 133, 82, 0.1);
  color: var(--primary);
  border-color: rgba(192, 133, 82, 0.2);
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

.quick-status-change {
  border-top: 1px dashed var(--border);
  padding-top: 8px;
  margin-top: auto;
}
.quick-status-btn {
  background: var(--btn-ghost-bg);
  border: 1px solid var(--border);
  border-radius: 6px;
  width: 26px;
  height: 26px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  font-size: 12px;
  transition: all 0.15s;
}
.quick-status-btn:hover {
  background: var(--btn-ghost-hover-bg);
}
.quick-status-btn.active {
  background: rgba(192, 133, 82, 0.15);
  border-color: var(--primary);
}

.item-personal-vote {
  background: var(--form-bg);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  padding: 6px 8px;
  margin-top: 8px;
}

.btn-clear-rating {
  font-size: 10px;
  color: var(--dusty-rose);
  background: transparent;
  border: none;
  cursor: pointer;
}

.star-pill-btn {
  background: var(--btn-ghost-bg);
  border: 1px solid var(--border);
  color: var(--text-secondary);
  width: 18px;
  height: 18px;
  font-size: 8px;
  font-weight: 700;
  border-radius: 3px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
  flex: 1;
}
.star-pill-btn:hover {
  background: rgba(192, 133, 82, 0.15);
  border-color: var(--primary);
  color: var(--primary);
}
.star-pill-btn.active {
  background: var(--primary);
  border-color: var(--primary);
  color: var(--latte-foam);
}

.item-links-section {
  border-top: 1px dashed var(--border);
  padding-top: 8px;
  margin-top: 8px;
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
  gap: 4px;
  background: var(--form-bg);
  padding: 6px;
  border-radius: var(--radius-md);
  border: 1px solid var(--border);
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

.link-del-small {
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 10px;
  cursor: pointer;
  padding: 0 4px;
}
.link-del-small:hover {
  color: var(--dusty-rose);
}

.watch-together-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
  padding: 4px 10px;
  border-radius: 6px;
  color: var(--primary);
  border: 1px solid rgba(192, 133, 82, 0.35);
  font-size: 10px;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s;
  width: 100%;
}

.tooltip-ratings {
  position: relative;
}
.tooltip-ratings .tooltip-text {
  visibility: hidden;
  position: absolute;
  bottom: 125%;
  right: 0;
  width: 160px;
  background: var(--bg-elevated);
  border: 1px solid var(--border);
  color: var(--text-primary);
  text-align: left;
  padding: 8px 12px;
  border-radius: 8px;
  z-index: 100;
  opacity: 0;
  transition: opacity 0.2s;
  box-shadow: var(--shadow-md);
  pointer-events: none;
}
.tooltip-ratings:hover .tooltip-text {
  visibility: visible;
  opacity: 1;
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

/* ══ Модальные окна ══ */
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
  width: 100%; max-width: 520px;
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

.modal-title { font-family: var(--font-display); font-size: 18px; font-weight: 700; color: var(--text-primary); }

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
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.form-group { display: flex; flex-direction: column; gap: 6px; }
.form-label { font-size: 13px; font-weight: 600; color: var(--text-secondary); }

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

.invite-card {
  border: 1px solid var(--border);
  background: var(--bg-base);
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
.groups-backdrop, .filters-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(20, 16, 13, 0.6);
  backdrop-filter: blur(8px);
  z-index: 290;
}

.mobile-toggle-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  background: var(--btn-ghost-bg);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  color: var(--text-primary);
  font-size: 11px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
}
.mobile-toggle-btn:hover {
  background: var(--btn-ghost-hover-bg);
  border-color: var(--primary);
}

.drawer-header {
  border-bottom: 1px solid var(--border);
  padding-bottom: 12px;
}
.drawer-title {
  font-weight: 700;
  font-size: 15px;
  color: var(--primary);
}
.drawer-close-btn {
  background: transparent;
  border: none;
  font-size: 16px;
  color: var(--text-secondary);
  cursor: pointer;
}
.drawer-close-btn:hover {
  color: var(--primary);
}

.card-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: 20px;
}

@media (max-width: 1280px) {
  .main-layout {
    grid-template-columns: 1fr;
    gap: 0;
  }
  .sidebar {
    position: fixed;
    top: 0;
    left: -320px;
    width: 300px;
    height: 100vh;
    border-radius: 0;
    border: none;
    border-right: 1px solid var(--border);
    z-index: 300;
    transition: left 0.3s cubic-bezier(0.16, 1, 0.3, 1);
    display: flex;
    flex-direction: column;
    overflow-y: auto;
    padding: 24px;
    background: var(--bg-surface);
  }
  .sidebar.is-mobile-open {
    left: 0;
  }
  
  .catalog-layout {
    grid-template-columns: 1fr;
    padding: 16px;
  }
  .sidebar-filters {
    position: fixed;
    top: 0;
    left: -320px;
    width: 300px;
    height: 100vh;
    border-radius: 0;
    border: none;
    border-right: 1px solid var(--border);
    z-index: 300;
    transition: left 0.3s cubic-bezier(0.16, 1, 0.3, 1);
    display: flex;
    flex-direction: column;
    overflow-y: auto;
    padding: 24px;
    background: var(--bg-surface);
  }
  .sidebar-filters.is-mobile-open {
    left: 0;
  }
  
  .card-grid {
    grid-template-columns: repeat(auto-fill, minmax(190px, 1fr));
  }
}

@media (max-width: 640px) {
  .card-grid {
    grid-template-columns: 1fr;
  }
}
</style>
