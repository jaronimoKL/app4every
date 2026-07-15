<template>
  <div class="dashboard-page">
    <main class="dashboard-main">

      <!-- Сетка Дашборда -->
      <div class="dashboard-grid">
        
        <!-- Левая колонка: Основной функционал -->
        <div class="main-column">
          <!-- Герой-приветствие -->
          <section class="welcome-hero glass">
            <div class="orb-blur" style="background: var(--caramel); top: -20%; right: -10%;"></div>
            <div class="orb-blur" style="background: var(--cinnamon); bottom: -30%; left: -10%;"></div>
            
            <div class="welcome-content">
              <div class="status-pill">
                <span class="pulse-dot"></span>
                <span>Система активна</span>
              </div>
              <h1 class="welcome-title">
                Привет, <span class="gradient-text">{{ auth.user?.username || auth.user?.email?.split('@')[0] }}</span> 👋
              </h1>
              <p class="welcome-subtitle">
                Добро пожаловать в персональную экосистему App4Every. Все модули изолированы и работают независимо друг от друга.
              </p>
              
              <!-- Быстрая статистика -->
              <div class="quick-stats-row">
                <div class="stat-badge">
                  <span class="stat-num">4</span>
                  <span class="stat-label">модуля подключено</span>
                </div>
                <div class="stat-badge">
                  <span class="stat-num">Go</span>
                  <span class="stat-label">на бэкенде</span>
                </div>
                <div class="stat-badge">
                  <span class="stat-num">DB</span>
                  <span class="stat-label">PostgreSQL + Redis</span>
                </div>
              </div>
            </div>
          </section>

          <!-- Секция модулей -->
          <section class="modules-section">
            <h2 class="section-title">Доступные модули</h2>
            
            <div class="modules-grid">
              <component
                v-for="mod in modules"
                :key="mod.name"
                :is="mod.link ? 'RouterLink' : 'div'"
                :to="mod.link || undefined"
                class="module-card glass glass-hover"
                style="text-decoration: none;"
              >
                <!-- Подсвечивающийся задний фон при наведении -->
                <div class="card-glow" :style="{ background: mod.glowColor }"></div>
                
                <div class="card-header-row">
                  <div class="module-icon-box" :style="{ background: mod.iconBg }">
                    <span class="m-icon">{{ mod.icon }}</span>
                  </div>
                  <span v-if="mod.link" class="status-tag active-tag">Доступно</span>
                  <span v-else class="status-tag coming-tag">В разработке</span>
                </div>

                <div class="module-details">
                  <h3 class="module-card-title">{{ mod.name }}</h3>
                  <p class="module-card-desc">{{ mod.desc }}</p>
                </div>

                <div class="module-card-footer">
                  <div class="tech-stack-row">
                    <span v-for="tag in mod.tech" :key="tag" class="tech-capsule">{{ tag }}</span>
                  </div>
                  <span v-if="mod.link" class="action-arrow">Открыть →</span>
                </div>
              </component>
            </div>
          </section>
        </div>

        <!-- Правая колонка: Боковая панель -->
        <aside class="sidebar-column">
          <!-- Статус Стэка -->
          <section class="panel-section glass">
            <h3 class="panel-title">Статус инфраструктуры</h3>
            <div class="stack-list">
              <div v-for="item in stack" :key="item.name" class="stack-item-row">
                <span class="stack-emoji">{{ item.icon }}</span>
                <div class="stack-info-meta">
                  <div class="stack-name">{{ item.name }}</div>
                  <div class="stack-role">{{ item.role }}</div>
                </div>
                <div class="uptime-badge" :class="item.status">
                  <span class="uptime-dot"></span>
                  <span>{{ item.status === 'online' ? 'Uptime' : 'Offline' }}</span>
                </div>
              </div>
            </div>
          </section>

          <!-- Системный лог -->
          <section class="panel-section glass">
            <h3 class="panel-title">Системные события</h3>
            <div class="sys-log-box">
              <div v-for="(log, idx) in sysLogs" :key="idx" class="log-entry">
                <span class="log-time">{{ log.time }}</span>
                <span class="log-text">{{ log.message }}</span>
              </div>
            </div>
          </section>
        </aside>

      </div>
    </main>
  </div>
</template>

<script setup>
import { computed, onMounted, ref } from 'vue'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()

const modules = [
  {
    icon: '📝',
    name: 'Заметки',
    desc: 'Персональный блокнот с автосохранением и разметкой. Готов к быстрым записям.',
    iconBg: 'rgba(169, 103, 59, 0.15)',
    glowColor: 'rgba(169, 103, 59, 0.05)',
    tech: ['Go', 'PostgreSQL', 'WebSockets'],
    link: '/notes',
  },
  {
    icon: '⭐',
    name: 'Рецензии',
    desc: 'Каталог фильмов, сериалов и аниме. Оценки, статусы, комментарии и ссылки.',
    iconBg: 'rgba(192, 133, 82, 0.15)',
    glowColor: 'rgba(192, 133, 82, 0.06)',
    tech: ['Go', 'PostgreSQL', 'TMDB API'],
    link: '/reviews',
  },
  {
    icon: '📹',
    name: 'Видеочат',
    desc: 'Комнаты для звонков с поддержкой WebRTC и шеринга экрана без лишних задержек.',
    iconBg: 'rgba(124, 154, 110, 0.18)',
    glowColor: 'rgba(124, 154, 110, 0.05)',
    tech: ['Go Signalling', 'WebRTC'],
    link: '/screenshare',
  },
  {
    icon: '📺',
    name: 'Watch Party',
    desc: 'Синхронный просмотр видеофайлов и YouTube совместно с друзьями.',
    iconBg: 'rgba(201, 112, 100, 0.15)',
    glowColor: 'rgba(201, 112, 100, 0.05)',
    tech: ['Go WS', 'YouTube API'],
    link: '/watch',
  },
]

const stack = [
  { icon: '🐹', name: 'Go Backend',   role: 'REST API + WS Signalling', status: 'online' },
  { icon: '🐘', name: 'PostgreSQL',   role: 'Основная база данных',     status: 'online' },
  { icon: '🔴', name: 'Redis',        role: 'Кэш и сессии севера',       status: 'online' },
  { icon: '⚡', name: 'Caddy server', role: 'Реверс-прокси шлюз',       status: 'online' },
  { icon: '🟢', name: 'Vue 3 + Vite', role: 'SPA клиентский бандл',     status: 'online' },
]

const sysLogs = ref([
  { time: '14:02', message: 'Авторизована сессия пользователя' },
  { time: '14:00', message: 'Сигнальные сокеты WebRTC подключены' },
  { time: '13:58', message: 'Синхронизация списков Shikimori завершена' },
  { time: '13:55', message: 'БД PostgreSQL: пул соединений стабилен' },
])

onMounted(async () => {
  if (!auth.user && auth.accessToken) {
    await auth.tryRestoreSession()
  }
})
</script>

<style scoped>
.dashboard-page {
  min-height: 100vh;
  background-color: var(--bg-base);
}

.dashboard-main {
  max-width: 1200px;
  margin: 0 auto;
  padding: 16px 24px 80px;
}

/* Сетка дашборда */
.dashboard-grid {
  display: grid;
  grid-template-columns: 8fr 3fr;
  gap: 24px;
  align-items: start;
}

/* Левая основная колонка */
.main-column {
  display: flex;
  flex-direction: column;
  gap: 32px;
}

/* Герой-приветствие */
.welcome-hero {
  position: relative;
  overflow: hidden;
  padding: 44px;
  border-radius: var(--radius-xl);
  border: 1px solid var(--border);
  background: var(--bg-surface);
  box-shadow: var(--shadow-md);
}

.orb-blur {
  position: absolute;
  width: 320px;
  height: 320px;
  border-radius: 50%;
  filter: blur(100px);
  opacity: 0.1;
  pointer-events: none;
}

.welcome-content {
  position: relative;
  z-index: 1;
}

.status-pill {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 6px 14px;
  border-radius: 99px;
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  background: rgba(124, 154, 110, 0.12);
  border: 1px solid rgba(124, 154, 110, 0.25);
  color: var(--matcha);
  margin-bottom: 24px;
}

.pulse-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--matcha);
  box-shadow: 0 0 8px var(--matcha);
  animation: pulse 2s infinite;
}

@keyframes pulse {
  0% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(124, 154, 110, 0.7); }
  70% { transform: scale(1); box-shadow: 0 0 0 6px rgba(124, 154, 110, 0); }
  100% { transform: scale(0.95); box-shadow: 0 0 0 0 rgba(124, 154, 110, 0); }
}

.welcome-title {
  font-family: var(--font-display);
  font-size: clamp(28px, 4vw, 44px);
  font-weight: 800;
  letter-spacing: -0.02em;
  color: var(--text-primary);
  margin-bottom: 12px;
  line-height: 1.2;
}

.welcome-subtitle {
  color: var(--text-secondary);
  font-size: 15px;
  line-height: 1.6;
  max-width: 520px;
  margin-bottom: 32px;
}

/* Быстрая статистика */
.quick-stats-row {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
  border-top: 1px dashed var(--border);
  padding-top: 24px;
}

.stat-badge {
  display: flex;
  align-items: center;
  gap: 8px;
  background: var(--btn-ghost-bg);
  border: 1px solid var(--border);
  padding: 8px 16px;
  border-radius: var(--radius-md);
}

.stat-num {
  font-family: var(--font-mono);
  font-size: 15px;
  font-weight: 700;
  color: var(--primary);
}

.stat-label {
  font-size: 12px;
  color: var(--text-muted);
  font-weight: 500;
}

/* Секция модулей */
.section-title {
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.08em;
  color: var(--text-muted);
  text-transform: uppercase;
  margin-bottom: 16px;
}

.modules-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
}

.module-card {
  position: relative;
  border-radius: var(--radius-lg);
  padding: 24px;
  background: var(--bg-surface);
  border: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  justify-content: space-between;
  min-height: 220px;
  overflow: hidden;
  transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
  box-shadow: var(--shadow-sm);
}

.module-card:hover {
  transform: translateY(-4px) scale(1.01);
  border-color: var(--border-hover);
  box-shadow: var(--shadow-lg);
}

.card-glow {
  position: absolute;
  inset: 0;
  filter: blur(40px);
  opacity: 0;
  transition: opacity 0.3s;
  pointer-events: none;
}
.module-card:hover .card-glow {
  opacity: 1;
}

.card-header-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  position: relative;
  z-index: 2;
}

.module-icon-box {
  width: 44px;
  height: 44px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: var(--shadow-sm);
}

.m-icon {
  font-size: 20px;
}

.status-tag {
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.05em;
  padding: 4px 8px;
  border-radius: 20px;
}

.active-tag {
  background: rgba(124, 154, 110, 0.12);
  color: var(--matcha);
  border: 1px solid rgba(124, 154, 110, 0.25);
}

.coming-tag {
  background: var(--btn-ghost-bg);
  color: var(--text-muted);
  border: 1px solid var(--border);
}

.module-details {
  position: relative;
  z-index: 2;
  margin: 16px 0;
}

.module-card-title {
  font-family: var(--font-sans);
  font-size: 18px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 6px;
}

.module-card-desc {
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.5;
}

.module-card-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  position: relative;
  z-index: 2;
  border-top: 1px solid var(--border);
  padding-top: 14px;
  margin-top: 8px;
}

.tech-stack-row {
  display: flex;
  gap: 4px;
  flex-wrap: wrap;
}

.tech-capsule {
  font-family: var(--font-mono);
  font-size: 9px;
  font-weight: 600;
  padding: 2px 6px;
  border-radius: 4px;
  background: var(--btn-ghost-bg);
  border: 1px solid var(--border);
  color: var(--text-muted);
}

.action-arrow {
  font-size: 12px;
  font-weight: 700;
  color: var(--primary);
  transition: transform 0.2s;
}

.module-card:hover .action-arrow {
  transform: translateX(3px);
  color: var(--primary-hover);
}

/* Правая колонка / Сайдбар */
.sidebar-column {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

.panel-section {
  padding: 20px;
  border-radius: var(--radius-lg);
  border: 1px solid var(--border);
  background: var(--bg-surface);
}

.panel-title {
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.08em;
  color: var(--text-muted);
  text-transform: uppercase;
  margin-bottom: 16px;
}

.stack-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.stack-item-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding-bottom: 10px;
  border-bottom: 1px solid var(--border);
}
.stack-item-row:last-child {
  border-bottom: none;
  padding-bottom: 0;
}

.stack-emoji {
  font-size: 18px;
  width: 28px;
  height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--btn-ghost-bg);
  border-radius: 50%;
}

.stack-info-meta {
  flex: 1;
}

.stack-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.stack-role {
  font-size: 11px;
  color: var(--text-muted);
}

.uptime-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 10px;
  font-weight: 600;
  color: var(--matcha);
}

.uptime-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--matcha);
  box-shadow: 0 0 6px var(--matcha);
}

/* Логи событий */
.sys-log-box {
  font-family: var(--font-mono);
  font-size: 11px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.log-entry {
  display: flex;
  gap: 8px;
  line-height: 1.4;
}

.log-time {
  color: var(--primary);
  flex-shrink: 0;
}

.log-text {
  color: var(--text-secondary);
}

/* Адаптив */
@media (max-width: 1280px) {
  .dashboard-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 640px) {
  .modules-grid {
    grid-template-columns: 1fr;
  }
  .welcome-hero {
    padding: 24px;
  }
  .quick-stats-row {
    flex-wrap: wrap;
    gap: 8px;
  }
  .stat-badge {
    padding: 6px 12px;
  }
}
</style>
