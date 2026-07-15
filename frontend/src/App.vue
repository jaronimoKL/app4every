<template>
  <div class="app-layout">
    <!-- Мобильная шапка -->
    <header v-if="authStore.isAuthenticated" class="mobile-top-navbar">
      <div class="flex items-center">
        <div class="mobile-logo-pod-box">
          <span class="mobile-logo-pod-text">A4E</span>
        </div>
        <span class="mobile-brand-name">App4Every</span>
      </div>
      
      <!-- Кнопка гамбургера -->
      <button 
        class="mobile-hamburger-btn" 
        :class="{ 'is-active': sidebarStore.isGlobalExpanded }"
        @click="sidebarStore.isGlobalExpanded = !sidebarStore.isGlobalExpanded"
      >
        <span class="hamburger-line"></span>
        <span class="hamburger-line"></span>
        <span class="hamburger-line"></span>
      </button>
    </header>

    <!-- Оверлей для мобильного сайдбара -->
    <div 
      v-if="authStore.isAuthenticated && sidebarStore.isGlobalExpanded" 
      class="sidebar-mobile-backdrop" 
      @click="sidebarStore.isGlobalExpanded = false"
    ></div>

    <AppSidebar 
      v-if="authStore.isAuthenticated" 
      :items="globalMenuItems"
      :collapsible="true"
      :force-collapse="isSidebarLockedCollapsed"
    />
    <div 
      class="app-main-content" 
      :class="{ 
        'with-sidebar': authStore.isAuthenticated, 
        'sidebar-expanded': sidebarStore.isGlobalExpanded && !isSidebarLockedCollapsed 
      }"
    >
      <RouterView v-slot="{ Component }">
        <Transition name="page" mode="out-in">
          <component :is="Component" />
        </Transition>
      </RouterView>
    </div>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { useAuthStore } from './stores/auth'
import { useThemeStore } from './stores/theme'
import { useSidebarStore } from './stores/sidebar'
import AppSidebar from './components/AppSidebar.vue'

const authStore = useAuthStore()
const themeStore = useThemeStore() // initialises theme
const sidebarStore = useSidebarStore()
const route = useRoute()

// Define global modules navigation menu items
const globalMenuItems = [
  { key: 'dashboard', title: 'Главная', icon: '🏠', route: '/dashboard' },
  { key: 'notes', title: 'Заметки', icon: '📝', route: '/notes' },
  { key: 'reviews', title: 'Рецензии', icon: '⭐', route: '/reviews' },
  { key: 'watchparty', title: 'Watch Party', icon: '📺', route: '/watch' },
  { key: 'screenshare', title: 'Видеочат', icon: '💬', route: '/screenshare' },
  { key: 'calendar', title: 'Календарь', icon: '📅', route: '/calendar' },
]

// Automatically lock global sidebar to collapsed rail mode when on the profile page
const isSidebarLockedCollapsed = computed(() => {
  return route.path === '/profile'
})
</script>

<style>
.app-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: row;
}

.app-main-content {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  transition: padding-left 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

/* ── Мобильная шапка ── */
.mobile-top-navbar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: 56px;
  background: rgba(255, 248, 240, 0.85); /* latte foam transparent */
  backdrop-filter: blur(12px);
  -webkit-backdrop-filter: blur(12px);
  border-bottom: 1px solid var(--border);
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 16px;
  z-index: 150;
}
[data-theme='espresso'] .mobile-top-navbar {
  background: rgba(20, 16, 13, 0.85); /* obsidian transparent */
}

.mobile-logo-pod-box {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: var(--primary);
  display: flex;
  align-items: center;
  justify-content: center;
}
.mobile-logo-pod-text {
  font-family: var(--font-mono);
  font-size: 11px;
  font-weight: 800;
  color: var(--latte-foam);
}
[data-theme='espresso'] .mobile-logo-pod-text {
  color: var(--espresso);
}

.mobile-brand-name {
  font-family: var(--font-sans);
  font-size: 15px;
  font-weight: 800;
  color: var(--text-primary);
  margin-left: 8px;
}

.mobile-hamburger-btn {
  background: transparent;
  border: none;
  display: flex;
  flex-direction: column;
  gap: 5px;
  cursor: pointer;
  padding: 8px;
}
.hamburger-line {
  width: 20px;
  height: 2px;
  background-color: var(--text-primary);
  border-radius: 1px;
  transition: all 0.2s;
}

.mobile-hamburger-btn.is-active .hamburger-line:nth-child(1) {
  transform: translateY(7px) rotate(45deg);
}
.mobile-hamburger-btn.is-active .hamburger-line:nth-child(2) {
  opacity: 0;
}
.mobile-hamburger-btn.is-active .hamburger-line:nth-child(3) {
  transform: translateY(-7px) rotate(-45deg);
}

.sidebar-mobile-backdrop {
  position: fixed;
  inset: 0;
  background: rgba(20, 16, 13, 0.4);
  backdrop-filter: blur(4px);
  -webkit-backdrop-filter: blur(4px);
  z-index: 180;
}

@media (max-width: 768px) {
  .app-main-content {
    padding-top: 56px; /* Offset for mobile header */
  }
}

@media (min-width: 769px) {
  .mobile-top-navbar, .sidebar-mobile-backdrop {
    display: none !important;
  }
  .app-main-content.with-sidebar {
    padding-left: 64px;
  }
  .app-main-content.with-sidebar.sidebar-expanded {
    padding-left: 220px;
  }
}
</style>
