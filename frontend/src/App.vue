<template>
  <div class="app-layout">
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

@media (min-width: 769px) {
  .app-main-content.with-sidebar {
    padding-left: 64px;
  }
  .app-main-content.with-sidebar.sidebar-expanded {
    padding-left: 220px;
  }
}
</style>
