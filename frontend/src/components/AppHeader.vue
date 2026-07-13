<template>
  <nav class="navbar glass">
    <div class="navbar-inner max-w-[1200px] mx-auto w-full flex justify-between items-center px-6 py-3">
      <RouterLink to="/dashboard" class="flex items-center gap-3" style="text-decoration:none; color:inherit;">
        <div class="logo-mark" style="width:36px;height:36px;font-size:16px;color:var(--latte-foam);">⬡</div>
        <span style="font-family: var(--font-display); font-weight: 800; font-size: 19px; letter-spacing: -0.02em; color: var(--text-primary);">App4Every</span>
      </RouterLink>

      <!-- Навигация (центр) - если нужно, можно добавить ссылки -->
      <div class="nav-links flex gap-4 hidden md:flex">
        <RouterLink to="/dashboard" class="nav-link" active-class="active">Главная</RouterLink>
        <RouterLink to="/calendar" class="nav-link" active-class="active">Календарь</RouterLink>
      </div>

      <!-- Правая часть: переключатель темы + уведомления + пользователь + logout -->
      <div class="flex items-center gap-4">
        <!-- Кнопка смены темы -->
        <button @click="themeStore.toggleTheme()" class="btn btn-ghost !p-2" title="Сменить тему">
          <svg v-if="themeStore.currentTheme === 'macchiato'" xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
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
          <svg v-else xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z"></path>
          </svg>
        </button>

        <NotificationBell />

        <RouterLink to="/profile" class="user-chip glass" style="text-decoration:none;cursor:pointer;" title="Профиль">
          <div class="user-avatar">{{ userInitial }}</div>
          <span class="user-email hidden sm:block">{{ authStore.user?.username || authStore.user?.email }}</span>
        </RouterLink>

        <button @click="authStore.logout()" class="btn btn-ghost" style="padding:8px 14px;font-size:13px;">
          Выйти
        </button>
      </div>
    </div>
  </nav>
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
.navbar {
  position: sticky;
  top: 0;
  z-index: 50;
  border-bottom: 1px solid var(--border);
  transition: background-color 0.3s ease, border-color 0.3s ease;
}

.user-chip {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 6px 14px 6px 6px;
  border-radius: 99px;
  background: var(--glass-bg);
  border: 1px solid var(--border);
  transition: all 0.2s ease;
}

.user-chip:hover {
  background: var(--glass-hover-bg);
  border-color: var(--border-hover);
  transform: translateY(-1px);
}

.nav-link {
  color: var(--text-secondary);
  text-decoration: none;
  font-weight: 600;
  font-size: 0.95rem;
  padding: 6px 0;
  margin: 0 10px;
  position: relative;
  transition: color 0.2s ease;
}

.nav-link::after {
  content: '';
  position: absolute;
  bottom: -2px;
  left: 0;
  width: 0;
  height: 2px;
  background-color: var(--primary);
  transition: width 0.2s ease;
}

.nav-link:hover {
  color: var(--text-primary);
}

.nav-link:hover::after {
  width: 100%;
}

.nav-link.active {
  color: var(--primary);
}

.nav-link.active::after {
  width: 100%;
}

.user-avatar {
  width: 28px;
  height: 28px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--primary), var(--cinnamon));
  color: var(--latte-foam);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 13px;
  font-weight: 700;
}

.user-email {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
}
</style>
