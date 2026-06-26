<template>
  <div class="dashboard-page">
    <!-- ── Основной контент ── -->
    <main class="dashboard-main">

      <!-- Приветствие -->
      <section class="welcome-section">
        <div class="orb" style="width:500px;height:500px;background:#6366f1;top:-150px;right:-100px;opacity:0.08;"></div>
        <div class="orb" style="width:300px;height:300px;background:#8b5cf6;bottom:0;left:-50px;opacity:0.07;"></div>

        <div class="welcome-content">
          <div class="badge-chip" style="margin-bottom:16px;">
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
              <span v-if="mod.link" class="badge-chip" style="font-size:11px;padding:3px 10px;">Открыть →</span>
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
    iconBg: 'rgba(99,102,241,0.15)',
    tech: ['Go', 'PostgreSQL'],
    link: '/notes',  // уже работает!
  },
  {
    icon: '⭐',
    name: 'Рецензии',
    desc: 'Список просмотренного: фильмы, аниме, сериалы. 4 статуса + оценки + ссылки на Kinopoisk, IMDB, Shikimori.',
    iconBg: 'rgba(245,158,11,0.15)',
    tech: ['Go', 'PostgreSQL'],
    link: '/reviews',
  },
  {
    icon: '📹',
    name: 'Видеочат',
    desc: 'P2P видеозвонки и демонстрация экрана на базе WebRTC. Go выступает только сигнальным сервером.',
    iconBg: 'rgba(6,182,212,0.15)',
    tech: ['Go', 'WebRTC', 'WebSocket'],
    link: '/screenshare',
  },
  {
    icon: '📺',
    name: 'Watch Party',
    desc: 'Смотри YouTube, Rutube и видеофайлы вместе с друзьями. Синхронный play/pause/seek.',
    iconBg: 'rgba(239,68,68,0.15)',
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

/* ── Навбар ── */
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
  max-width: 1200px;
  margin: 0 auto;
  padding: 14px 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.user-chip {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px 6px 6px;
  border-radius: 99px;
}

.user-avatar {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--primary), var(--violet));
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 700;
  flex-shrink: 0;
}

.user-email {
  font-size: 13px;
  color: var(--text-secondary);
  max-width: 200px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
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

.welcome-content { position: relative; z-index: 1; }

.welcome-title {
  font-size: clamp(28px, 4vw, 48px);
  font-weight: 800;
  letter-spacing: -0.03em;
  margin-bottom: 12px;
}

/* ── Секции ── */
.section-title {
  font-size: 18px;
  font-weight: 700;
  letter-spacing: -0.01em;
  color: var(--text-secondary);
  text-transform: uppercase;
  font-size: 12px;
  letter-spacing: 0.08em;
  margin-bottom: 16px;
}

/* ── Карточки модулей ── */
.modules-section { margin-bottom: 40px; }

.modules-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
  gap: 16px;
}

.module-card {
  border-radius: var(--radius-lg);
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 14px;
  min-height: 200px;
}

.module-info { flex: 1; }
.module-name {
  font-size: 17px;
  font-weight: 700;
  margin-bottom: 6px;
}
.module-desc {
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.6;
}

.module-footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.module-tech {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
}

.tech-tag {
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 99px;
  background: rgba(99,102,241,0.1);
  color: #a5b4fc;
  border: 1px solid rgba(99,102,241,0.2);
  font-weight: 500;
}

/* ── Стек ── */
.stack-section { margin-bottom: 40px; }

.stack-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 10px;
}

.stack-item {
  border-radius: var(--radius-md);
  padding: 14px 16px;
  display: flex;
  align-items: center;
  gap: 12px;
}

.stack-icon { font-size: 20px; flex-shrink: 0; }

.status-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  margin-left: auto;
  flex-shrink: 0;
}
.status-dot.online {
  background: #22c55e;
  box-shadow: 0 0 8px rgba(34,197,94,0.5);
}
.status-dot.offline { background: #475569; }
</style>
