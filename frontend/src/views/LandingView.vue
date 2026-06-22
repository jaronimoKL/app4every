<template>
  <div class="landing-page">
    <!-- ── Декоративный фон ── -->
    <div class="orb" style="width:700px;height:700px;background:#6366f1;top:-250px;left:-200px;"></div>
    <div class="orb" style="width:500px;height:500px;background:#8b5cf6;bottom:-150px;right:-150px;"></div>
    <div class="orb" style="width:300px;height:300px;background:#06b6d4;top:40%;right:15%;opacity:0.08;"></div>
    <div class="dot-grid"></div>

    <!-- ── Контент ── -->
    <div class="landing-body">

      <!-- Шапка -->
      <header class="landing-header">
        <div class="flex items-center gap-3">
          <div class="logo-mark">⬡</div>
          <span style="font-weight:700;font-size:18px;letter-spacing:-0.02em;">App4Every</span>
        </div>
        <div v-if="auth.isAuthenticated" class="flex items-center gap-3">
          <RouterLink to="/dashboard" class="user-chip glass" style="text-decoration:none;cursor:pointer;">
            <div class="user-avatar">{{ userInitial }}</div>
            <span class="user-email">В Дашборд</span>
          </RouterLink>
        </div>
        <div v-else class="flex items-center gap-3">
          <RouterLink to="/login" class="btn btn-ghost" style="padding:8px 16px;font-size:13px;">
            Войти
          </RouterLink>
          <RouterLink to="/register" class="btn btn-primary" style="padding:8px 16px;font-size:13px;">
            Регистрация
          </RouterLink>
        </div>
      </header>

      <!-- Герой -->
      <main class="hero-section">
        <div class="badge-chip" style="margin-bottom:28px;">
          <span class="badge-dot"></span>
          Приватная экосистема · только для своих
        </div>

        <h1 class="hero-title">
          Твоё личное<br>
          <span class="gradient-text">цифровое пространство</span>
        </h1>

        <p class="hero-description">
          Заметки, отзывы и видеосвязь — всё в одном месте.<br>
          Никакой рекламы, никакого трекинга. Только ты и близкие.
        </p>

        <div v-if="auth.isAuthenticated" class="hero-actions">
          <RouterLink to="/dashboard" class="btn btn-primary" style="padding:14px 28px;font-size:15px;">
            Открыть дашборд →
          </RouterLink>
        </div>
        <div v-else class="hero-actions">
          <RouterLink to="/register" class="btn btn-primary" style="padding:14px 28px;font-size:15px;">
            Создать аккаунт →
          </RouterLink>
          <RouterLink to="/login" class="btn btn-outline" style="padding:14px 28px;font-size:15px;">
            Уже есть аккаунт
          </RouterLink>
        </div>

        <!-- Карточки модулей -->
        <div class="modules-grid">
          <div v-for="mod in modules" :key="mod.name" class="module-preview glass glass-hover">
            <div class="module-icon" :style="{ background: mod.bg }">{{ mod.icon }}</div>
            <div>
              <div style="font-weight:600;font-size:15px;margin-bottom:4px;">{{ mod.name }}</div>
              <div style="font-size:13px;color:var(--text-secondary);line-height:1.5;">{{ mod.desc }}</div>
            </div>
            <span class="coming-soon" style="margin-top:auto;">Скоро</span>
          </div>
        </div>
      </main>

      <!-- Подвал -->
      <footer class="landing-footer">
        <span style="color:var(--text-muted);font-size:13px;">
          App4Every · Личный проект · Сделано с ❤️ на Go + Vue
        </span>
      </footer>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()

const userInitial = computed(() =>
  (auth.user?.username || auth.user?.email)?.charAt(0).toUpperCase() ?? '?'
)
const modules = [
  {
    icon: '📝',
    name: 'Заметки',
    desc: 'PWA + real-time синхронизация через WebSockets. Работает офлайн.',
    bg: 'rgba(99,102,241,0.15)',
  },
  {
    icon: '⭐',
    name: 'Отзывы',
    desc: 'Оценки фильмов, еды и сериалов. Совместные списки с друзьями.',
    bg: 'rgba(245,158,11,0.15)',
  },
  {
    icon: '📹',
    name: 'Видеочат',
    desc: 'P2P видеозвонки на WebRTC. Демонстрация экрана со звуком.',
    bg: 'rgba(6,182,212,0.15)',
  },
]
</script>

<style scoped>
.landing-page {
  min-height: 100vh;
  position: relative;
  overflow: hidden;
  background: var(--bg-base);
}

.landing-body {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  min-height: 100vh;
  max-width: 1100px;
  margin: 0 auto;
  padding: 0 24px;
}

.landing-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 24px 0;
}

.hero-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  text-align: center;
  padding: 60px 0 80px;
}

.hero-title {
  font-size: clamp(36px, 6vw, 68px);
  font-weight: 800;
  line-height: 1.1;
  letter-spacing: -0.03em;
  margin-bottom: 20px;
}

.hero-description {
  font-size: 17px;
  color: var(--text-secondary);
  line-height: 1.7;
  max-width: 460px;
  margin-bottom: 36px;
}

.hero-actions {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  justify-content: center;
  margin-bottom: 72px;
}

.modules-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 16px;
  width: 100%;
  max-width: 900px;
}

.module-preview {
  border-radius: var(--radius-lg);
  padding: 22px;
  display: flex;
  flex-direction: column;
  gap: 14px;
  text-align: left;
}

.landing-footer {
  padding: 24px 0;
  text-align: center;
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
  font-weight: 500;
}
</style>
