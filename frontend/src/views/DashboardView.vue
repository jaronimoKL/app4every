<template>
  <div class="dashboard-page">
    <!-- ── Основной контент ── -->
    <main class="dashboard-main">

      <!-- Приветствие -->
      <section class="welcome-section">
        <div class="orb" style="width:500px;height:500px;background:var(--caramel);top:-150px;right:-100px;opacity:0.06;"></div>
        <div class="orb" style="width:300px;height:300px;background:var(--cinnamon);bottom:0;left:-50px;opacity:0.05;"></div>
        <div class="dot-grid"></div>

        <div class="welcome-content">
          <div class="badge-chip badge-status-ok" style="margin-bottom:16px;">
            <span class="badge-dot"></span>
            Всё работает
          </div>
          <h1 class="welcome-title">
            Привет,
            <span class="gradient-text">{{ auth.user?.username || auth.user?.email?.split('@')[0] }}</span>
            👋
          </h1>
          <p style="color:var(--text-secondary);font-size:16px;max-width:480px;line-height:1.6;">
            Твоя приватная экосистема готова. Модули в разработке —
            следи за обновлениями.
          </p>
        </div>
      </section>

      <!-- Карточки модулей -->
      <section class="modules-section">
        <h2 class="section-title">Модули</h2>
        <div class="modules-grid">
          <component
            v-for="mod in modules"
            :key="mod.name"
            :is="mod.link ? 'RouterLink' : 'div'"
            :to="mod.link || undefined"
            class="module-card glass glass-hover"
            style="text-decoration:none;"
          >
            <div class="module-icon" :style="{ background: mod.iconBg }">
              {{ mod.icon }}
            </div>
            <div class="module-info">
              <div class="module-name">{{ mod.name }}</div>
              <div class="module-desc">{{ mod.desc }}</div>
            </div>
            <div class="module-footer">
              <span v-if="mod.link" class="btn-open-cta">Открыть →</span>
              <span v-else class="coming-soon">Скоро</span>
              <div class="module-tech">
                <span v-for="tag in mod.tech" :key="tag" class="tech-tag">{{ tag }}</span>
              </div>
            </div>
          </component>
        </div>
      </section>

      <!-- Статус стека -->
      <section class="stack-section">
        <h2 class="section-title">Стек</h2>
        <div class="stack-grid">
          <div v-for="item in stack" :key="item.name" class="stack-item glass">
            <span class="stack-icon">{{ item.icon }}</span>
            <div>
              <div style="font-weight:600;font-size:14px;">{{ item.name }}</div>
              <div style="font-size:12px;color:var(--text-muted);">{{ item.role }}</div>
            </div>
            <div class="status-dot" :class="item.status"></div>
          </div>
        </div>
      </section>

    </main>
  </div>
</template>

<script setup>
import { computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()

// Первая буква username или email для аватара
const userInitial = computed(() =>
  (auth.user?.username || auth.user?.email)?.charAt(0).toUpperCase() ?? '?'
)

const modules = [
  {
    icon: '📝',
    name: 'Заметки',
    desc: 'Полностью адаптивные заметки с автосохранением и поиском. Real-time синхронизация через WebSockets — в планах.',
    iconBg: 'rgba(169, 103, 59, 0.15)', // Cinnamon/Clay
    tech: ['Go', 'PostgreSQL'],
    link: '/notes',  // уже работает!
  },
  {
    icon: '⭐',
    name: 'Рецензии',
    desc: 'Список просмотренного: фильмы, аниме, сериалы. 4 статуса + оценки + ссылки на Kinopoisk, IMDB, Shikimori.',
    iconBg: 'rgba(192, 133, 82, 0.15)', // Caramel/Ochre
    tech: ['Go', 'PostgreSQL'],
    link: '/reviews',
  },
  {
    icon: '📹',
    name: 'Видеочат',
    desc: 'P2P видеозвонки и демонстрация экрана на базе WebRTC. Go выступает только сигнальным сервером.',
    iconBg: 'rgba(124, 154, 110, 0.18)', // Matcha/Sage
    tech: ['Go', 'WebRTC', 'WebSocket'],
    link: '/screenshare',
  },
  {
    icon: '📺',
    name: 'Watch Party',
    desc: 'Смотри YouTube, Rutube и видеофайлы вместе с друзьями. Синхронный play/pause/seek.',
    iconBg: 'rgba(201, 112, 100, 0.15)', // Dusty-rose
    tech: ['Go', 'WebSocket', 'YouTube API'],
    link: '/watch',
  },
]

// Статус сервисов (в реальности можно подтягивать с /api/health)
const stack = [
  { icon: '🐹', name: 'Go Backend',   role: 'REST API + сигналинг', status: 'online' },
  { icon: '🐘', name: 'PostgreSQL',   role: 'Основная БД',          status: 'online' },
  { icon: '🔴', name: 'Redis',        role: 'Сессии + кэш',         status: 'online' },
  { icon: '⚡', name: 'Caddy',        role: 'Reverse Proxy',        status: 'online' },
  { icon: '🟢', name: 'Vue 3 + Vite', role: 'Фронтенд',             status: 'online' },
]

// При монтировании — проверяем что пользователь загружен
onMounted(async () => {
  if (!auth.user && auth.accessToken) {
    // Если токен есть, но пользователь не загружен (редкий случай)
    await auth.tryRestoreSession()
  }
})
</script>

<style scoped>
.dashboard-page {
  min-height: 100vh;
  background: var(--bg-base);
  position: relative;
}

/* ── Основной контент ── */
.dashboard-main {
  max-width: 1200px;
  margin: 0 auto;
  padding: 0 24px 80px;
}

/* ── Приветствие ── */
.welcome-section {
  position: relative;
  overflow: hidden;
  padding: 60px 0 48px;
}

.welcome-content {
  position: relative;
  z-index: 1;
}

.welcome-title {
  font-family: var(--font-display);
  font-size: clamp(32px, 5vw, 52px);
  font-weight: 800;
  letter-spacing: -0.03em;
  margin-bottom: 12px;
  color: var(--text-primary);
}

.badge-status-ok {
  background: rgba(124, 154, 110, 0.12) !important;
  border: 1px solid rgba(124, 154, 110, 0.25) !important;
  color: var(--matcha) !important;
}

/* ── Секции ── */
.section-title {
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.08em;
  color: var(--text-muted);
  text-transform: uppercase;
  margin-bottom: 16px;
}

/* ── Карточки модулей ── */
.modules-section {
  margin-bottom: 40px;
}

.modules-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 20px;
}

.module-card {
  border-radius: var(--radius-lg);
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 14px;
  min-height: 200px;
  transition: all 0.25s cubic-bezier(0.16, 1, 0.3, 1);
  box-shadow: var(--shadow-sm);
}

.module-card:hover {
  transform: translateY(-4px);
  box-shadow: var(--shadow-lg);
}

.module-info {
  flex: 1;
}

.module-name {
  font-family: var(--font-display);
  font-size: 20px;
  font-weight: 700;
  margin-bottom: 6px;
  color: var(--text-primary);
}

.module-desc {
  font-size: 13.5px;
  color: var(--text-secondary);
  line-height: 1.6;
}

.module-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
  margin-top: 10px;
}

.btn-open-cta {
  display: inline-flex;
  align-items: center;
  font-size: 12px;
  font-weight: 600;
  padding: 6px 14px;
  border-radius: 99px;
  background: var(--primary);
  color: var(--latte-foam);
  transition: all 0.25s cubic-bezier(0.16, 1, 0.3, 1);
  box-shadow: 0 2px 8px var(--primary-glow);
}

.module-card:hover .btn-open-cta {
  background: var(--primary-hover);
  transform: translateY(-1px);
  box-shadow: 0 4px 12px var(--primary-glow);
}

.module-tech {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.tech-tag {
  font-family: var(--font-mono);
  font-size: 10px;
  padding: 2px 7px;
  border-radius: 6px;
  background: rgba(107, 74, 54, 0.08);
  color: var(--text-muted);
  border: 1px solid var(--border);
  font-weight: 500;
}

/* ── Стек ── */
.stack-section {
  margin-bottom: 40px;
}

.stack-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 12px;
}

.stack-item {
  border-radius: var(--radius-md);
  padding: 14px 16px;
  display: flex;
  align-items: center;
  gap: 12px;
  box-shadow: var(--shadow-sm);
  transition: all 0.2s ease;
}

.stack-item:hover {
  transform: translateY(-1px);
  box-shadow: var(--shadow-md);
}

.stack-icon {
  font-size: 20px;
  flex-shrink: 0;
}

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-left: auto;
  flex-shrink: 0;
}

.status-dot.online {
  background: var(--matcha);
  box-shadow: 0 0 8px rgba(124, 154, 110, 0.5);
}

.status-dot.offline {
  background: var(--text-muted);
}
</style>
