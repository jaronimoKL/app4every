<template>
  <header class="header-container">
    <nav class="navbar glass">
      <div class="navbar-inner w-full flex justify-between items-center px-6 py-2.5">
        
        <!-- Логотип -->
        <RouterLink to="/dashboard" class="logo-link flex items-center gap-3">
          <div class="logo-pod">
            <span class="logo-char">⬡</span>
          </div>
          <span class="logo-text">App4Every</span>
        </RouterLink>

        <!-- Навигация (Центр) -->
        <div class="nav-links flex gap-1 hidden md:flex">
          <RouterLink to="/dashboard" class="nav-link" active-class="active">
            <span>Главная</span>
            <span class="indicator"></span>
          </RouterLink>
          <RouterLink to="/calendar" class="nav-link" active-class="active">
            <span>Календарь</span>
            <span class="indicator"></span>
          </RouterLink>
        </div>

        <!-- Правая часть -->
        <div class="flex items-center gap-3.5">
          <!-- Переключатель темы -->
          <button @click="themeStore.toggleTheme()" class="theme-switcher-btn" title="Сменить тему">
            <svg v-if="themeStore.currentTheme === 'macchiato'" xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
              <circle cx="12" cy="12" r="4"></circle>
              <path d="M12 2v2"></path>
              <path d="M12 20v2"></path>
              <path d="m4.93 4.93 1.41 1.41"></path>
              <path d="m17.66 17.66 1.41 1.41"></path>
              <path d="M2 12h2"></path>
              <path d="M20 12h2"></path>
              <path d="m6.34 17.66-1.41 1.41"></path>
              <path d="m19.07 4.93-1.41 1.41"></path>
            </svg>
            <svg v-else xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
              <path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z"></path>
            </svg>
          </button>

          <!-- Колокольчик -->
          <NotificationBell />

          <!-- Разделитель -->
          <div class="v-sep"></div>

          <!-- Профиль пользователя -->
          <RouterLink to="/profile" class="user-chip" title="Профиль">
            <div class="user-avatar">{{ userInitial }}</div>
            <span class="user-name hidden sm:block">{{ authStore.user?.username || authStore.user?.email?.split('@')[0] }}</span>
          </RouterLink>

          <!-- Кнопка Выйти -->
          <button @click="authStore.logout()" class="logout-btn">
            🚪 <span class="hidden md:inline">Выйти</span>
          </button>
        </div>
      </div>
    </nav>
  </header>
</template>

<script setup>
import { computed } from 'vue'
import { useAuthStore } from '../stores/auth'
import { useThemeStore } from '../stores/theme'
import NotificationBell from './NotificationBell.vue'

const authStore = useAuthStore()
const themeStore = useThemeStore()

const userInitial = computed(() => {
  const name = authStore.user?.username || authStore.user?.email || 'U'
  return name.charAt(0).toUpperCase()
})
</script>

<style scoped>
.header-container {
  position: sticky;
  top: 0;
  z-index: 50;
  padding: 14px 24px;
  width: 100%;
  pointer-events: none;
}

.navbar {
  pointer-events: auto;
  max-width: 1200px;
  margin: 0 auto;
  border-radius: var(--radius-xl);
  box-shadow: var(--shadow-sm);
  border: 1px solid var(--border);
  transition: all 0.3s cubic-bezier(0.16, 1, 0.3, 1);
  background: var(--glass-bg);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
}

.navbar:hover {
  border-color: var(--border-hover);
  box-shadow: var(--shadow-md);
}

.logo-link {
  text-decoration: none;
  color: inherit;
  transition: transform 0.2s ease;
}
.logo-link:hover {
  transform: scale(1.02);
}

.logo-pod {
  width: 34px;
  height: 34px;
  border-radius: var(--radius-md);
  background: linear-gradient(135deg, var(--primary), var(--cinnamon));
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: var(--shadow-sm);
}

.logo-char {
  font-size: 15px;
  font-weight: 800;
  color: var(--latte-foam);
}

.logo-text {
  font-family: var(--font-display);
  font-weight: 800;
  font-size: 19px;
  letter-spacing: -0.02em;
  color: var(--text-primary);
}

.nav-links {
  background: var(--btn-ghost-bg);
  padding: 4px;
  border-radius: 99px;
  border: 1px solid var(--border);
}

.nav-link {
  color: var(--text-secondary);
  text-decoration: none;
  font-weight: 600;
  font-size: 13px;
  padding: 6px 16px;
  border-radius: 99px;
  position: relative;
  transition: all 0.25s cubic-bezier(0.16, 1, 0.3, 1);
}

.nav-link:hover {
  color: var(--text-primary);
  background: var(--btn-ghost-hover-bg);
}

.nav-link.active {
  color: var(--primary);
  background: var(--bg-surface);
  box-shadow: var(--shadow-sm);
}

.v-sep {
  width: 1px;
  height: 20px;
  background: var(--border);
}

.theme-switcher-btn {
  background: var(--btn-ghost-bg);
  border: 1px solid var(--border);
  color: var(--text-secondary);
  width: 34px;
  height: 34px;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}
.theme-switcher-btn:hover {
  background: var(--btn-ghost-hover-bg);
  border-color: var(--border-hover);
  color: var(--text-primary);
  transform: rotate(15deg);
}

.user-chip {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 4px 12px 4px 4px;
  border-radius: 99px;
  background: var(--btn-ghost-bg);
  border: 1px solid var(--border);
  text-decoration: none;
  transition: all 0.2s ease;
}
.user-chip:hover {
  background: var(--btn-ghost-hover-bg);
  border-color: var(--border-hover);
  transform: translateY(-1px);
}

.user-avatar {
  width: 26px;
  height: 26px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--primary), var(--cinnamon));
  color: var(--latte-foam);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 12px;
  font-weight: 700;
}

.user-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.logout-btn {
  background: transparent;
  border: none;
  color: var(--text-muted);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  border-radius: var(--radius-md);
  transition: all 0.2s;
}
.logout-btn:hover {
  color: var(--dusty-rose);
  background: rgba(201, 112, 100, 0.08);
}

@media (max-width: 768px) {
  .header-container {
    padding: 8px 12px;
  }
}
</style>
