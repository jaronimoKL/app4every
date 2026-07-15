<template>
  <aside class="app-sidebar" :class="{ expanded: isExpanded, 'is-local': isLocal }">
    
    <!-- TOP LOCAL USER PROFILE (Only if isLocal is true) -->
    <div v-if="isLocal" class="local-sidebar-profile" @click="isExpanded = !isExpanded" style="cursor: pointer;" title="Свернуть/Развернуть">
      <div class="local-profile-avatar">{{ userInitial }}</div>
      <div class="local-profile-info" v-if="isExpanded">
        <div class="local-profile-name">{{ authStore.user?.username || authStore.user?.email?.split('@')[0] }}</div>
        <div class="local-profile-id">ID: {{ authStore.user?.id }}</div>
      </div>
    </div>
    <hr v-if="isLocal" class="local-sidebar-divider" />

    <!-- TOP LOGO CONTAINER (Only if not local) -->
    <div v-if="!isLocal" class="sidebar-top-section">
      <div class="logo-wrapper">
        <RouterLink to="/dashboard" class="logo-pod-link">
          <div class="logo-pod-box">
            <span class="logo-pod-text">A4E</span>
          </div>
          <span class="logo-brand-name" v-if="isExpanded">App4Every</span>
        </RouterLink>
      </div>
    </div>

    <!-- MIDDLE NAVIGATION ITEMS -->
    <nav class="sidebar-items-nav">
      <template v-for="item in items" :key="item.key">
        <!-- Render RouterLink if route is provided and it's not a local settings selection button -->
        <RouterLink 
          v-if="item.route" 
          :to="item.route"
          class="sidebar-item-link"
          :class="{ active: modelValue === item.key || $route?.path === item.route }"
          :title="!isExpanded ? item.title : ''"
          @click="onItemClick"
        >
          <span class="sidebar-item-icon">{{ item.icon }}</span>
          <span class="sidebar-item-label" v-if="isExpanded">{{ item.title }}</span>
          
          <!-- Badge -->
          <span v-if="item.badge" class="sidebar-item-badge">
            {{ item.badge }}
          </span>
        </RouterLink>

        <!-- Render regular action button if no route or is custom local action -->
        <button 
          v-else 
          class="sidebar-item-link"
          :class="{ active: modelValue === item.key }"
          @click="emit('update:modelValue', item.key); item.action && item.action(); onItemClick()"
          :title="!isExpanded ? item.title : ''"
        >
          <span class="sidebar-item-icon">{{ item.icon }}</span>
          <span class="sidebar-item-label" v-if="isExpanded">{{ item.title }}</span>
          
          <!-- Badge -->
          <span v-if="item.badge" class="sidebar-item-badge">
            {{ item.badge }}
          </span>
        </button>
      </template>
    </nav>

    <!-- Кнопка свернуть/развернуть -->
    <button 
      v-if="collapsible" 
      class="sidebar-collapse-btn" 
      @click="isExpanded = !isExpanded"
      :title="!isExpanded ? 'Развернуть меню' : ''"
    >
      <span class="collapse-icon">{{ isExpanded ? '❮' : '❯' }}</span>
      <span class="collapse-label" v-if="isExpanded">Свернуть</span>
    </button>

    <!-- BOTTOM SECTION (Only if not local) -->
    <div v-if="!isLocal" class="sidebar-bottom-section">
      <!-- Theme Switcher and Notification Bell inside horizontal row when expanded, or stacked when collapsed -->
      <div class="sidebar-actions-row" :class="{ stacked: !isExpanded }">
        
        <!-- Theme switcher -->
        <button @click="themeStore.toggleTheme()" class="action-circle-btn" title="Сменить тему">
          <svg v-if="themeStore.currentTheme === 'macchiato'" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
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
          <svg v-else xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round">
            <path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z"></path>
          </svg>
        </button>

        <!-- Notification bell -->
        <NotificationBell />
      </div>

      <!-- User Info block -->
      <div class="sidebar-user-block">
        <RouterLink to="/profile" class="user-profile-chip" :title="authStore.user?.username || 'Профиль'" @click="onItemClick">
          <div class="user-avatar-circle">{{ userInitial }}</div>
          <div class="user-text-meta" v-if="isExpanded">
            <span class="user-meta-name">{{ authStore.user?.username || authStore.user?.email?.split('@')[0] }}</span>
          </div>
        </RouterLink>

        <!-- Logout door button -->
        <button v-if="isExpanded" @click="authStore.logout()" class="logout-door-btn" title="Выйти">
          🚪
        </button>
      </div>
    </div>

  </aside>
</template>

<script setup>
import { ref, computed } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { useThemeStore } from '@/stores/theme'
import { useSidebarStore } from '@/stores/sidebar'
import NotificationBell from './NotificationBell.vue'

const props = defineProps({
  items: { type: Array, required: true },
  modelValue: { type: String, default: '' },
  collapsible: { type: Boolean, default: true },
  isLocal: { type: Boolean, default: false },
  forceCollapse: { type: Boolean, default: false },
  expanded: { type: Boolean, default: null }
})

const emit = defineEmits(['update:modelValue', 'update:expanded'])

const authStore = useAuthStore()
const themeStore = useThemeStore()
const sidebarStore = useSidebarStore()

// For local settings sidebars, we manage state via a local ref.
const localExpanded = ref(true)

const isExpanded = computed({
  get() {
    if (props.forceCollapse) return false
    if (props.expanded !== null) return props.expanded
    if (props.isLocal) return localExpanded.value
    return sidebarStore.isGlobalExpanded
  },
  set(val) {
    if (props.forceCollapse) return
    if (props.expanded !== null) {
      emit('update:expanded', val)
    } else if (props.isLocal) {
      localExpanded.value = val
    } else {
      sidebarStore.isGlobalExpanded = val
      localStorage.setItem('sidebar-expanded', val)
    }
  }
})

const userInitial = computed(() => {
  const name = authStore.user?.username || authStore.user?.email || 'U'
  return name.charAt(0).toUpperCase()
})

function onItemClick() {
  if (window.innerWidth < 769) {
    if (!props.isLocal) {
      sidebarStore.isGlobalExpanded = false
    } else {
      emit('update:expanded', false)
    }
  }
}
</script>

<style scoped>
.app-sidebar {
  width: 64px;
  height: 100vh;
  position: fixed;
  top: 0;
  left: 0;
  bottom: 0;
  z-index: 100;
  background: var(--bg-surface);
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  padding: 16px 8px;
  transition: width 0.3s cubic-bezier(0.16, 1, 0.3, 1), left 0.3s cubic-bezier(0.16, 1, 0.3, 1);
}

.app-sidebar.expanded {
  width: 220px;
}

.app-sidebar.is-local {
  position: static;
  height: auto;
  min-height: 400px;
  border-radius: var(--radius-xl);
  border: 1px solid var(--border);
  background: var(--bg-surface);
  padding: 16px 8px;
}

.local-sidebar-profile {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 4px;
}

.local-profile-avatar {
  width: 44px;
  height: 44px;
  border-radius: 50%;
  background: var(--primary);
  color: var(--espresso);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  font-weight: 800;
  box-shadow: var(--shadow-sm);
  flex-shrink: 0;
}
[data-theme='latte'] .local-profile-avatar {
  color: var(--latte-foam);
}

.local-profile-info {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.local-profile-name {
  font-size: 15px;
  font-weight: 700;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.local-profile-id {
  font-size: 11.5px;
  color: var(--text-muted);
  margin-top: 2px;
}

.local-sidebar-divider {
  border: none;
  border-top: 1px solid var(--border);
  margin: 12px 0 16px 0;
}

/* ── Top Section ── */
.sidebar-top-section {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 24px;
  padding: 0 4px;
}

.logo-wrapper {
  display: flex;
  align-items: center;
}

.logo-pod-link {
  display: flex;
  align-items: center;
  gap: 12px;
  text-decoration: none;
  color: inherit;
}

.logo-pod-box {
  width: 44px;
  height: 44px;
  border-radius: var(--radius-md);
  background: linear-gradient(135deg, var(--primary), var(--cinnamon));
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: var(--shadow-sm);
  flex-shrink: 0;
}

.logo-pod-text {
  font-family: var(--font-display);
  font-size: 15px;
  font-weight: 900;
  color: var(--latte-foam);
  letter-spacing: -0.02em;
}

.logo-brand-name {
  font-family: var(--font-sans);
  font-weight: 800;
  font-size: 16px;
  color: var(--text-primary);
  white-space: nowrap;
}

.sidebar-collapse-btn {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px;
  background: transparent;
  border: 1px solid transparent;
  color: var(--text-secondary);
  cursor: pointer;
  border-radius: var(--radius-md);
  transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  width: 100%;
  white-space: nowrap;
  text-align: left;
  margin-top: 12px;
  margin-bottom: 8px;
}
.sidebar-collapse-btn:hover {
  background: var(--btn-ghost-hover-bg);
  color: var(--text-primary);
  border-color: var(--border);
}

.collapse-icon {
  font-size: 11px;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.collapse-label {
  font-size: 13.5px;
}

/* ── Menu Navigation Items ── */
.sidebar-items-nav {
  display: flex;
  flex-direction: column;
  gap: 6px;
  flex: 1;
}

.sidebar-item-link {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px;
  background: transparent;
  border: 1px solid transparent;
  color: var(--text-secondary);
  cursor: pointer;
  border-radius: var(--radius-md);
  transition: all 0.2s cubic-bezier(0.16, 1, 0.3, 1);
  width: 100%;
  text-decoration: none;
  white-space: nowrap;
  position: relative;
  text-align: left;
}

.sidebar-item-link:hover {
  background: var(--btn-ghost-hover-bg);
  color: var(--text-primary);
  border-color: var(--border);
}

.sidebar-item-link.active {
  background: var(--espresso);
  color: var(--primary);
  border-color: var(--primary);
  font-weight: 700;
}
[data-theme='latte'] .sidebar-item-link.active {
  background: var(--cream);
}

.sidebar-item-icon {
  font-size: 16px;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.sidebar-item-label {
  font-size: 13.5px;
}

.sidebar-item-badge {
  background: var(--dusty-rose);
  color: var(--latte-foam);
  font-size: 9px;
  font-weight: 700;
  border-radius: 20px;
  padding: 1px 6px;
  margin-left: auto;
}

/* ── Bottom Section ── */
.sidebar-bottom-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
  border-top: 1px solid var(--border);
  padding-top: 16px;
  margin-top: auto;
}

.sidebar-actions-row {
  display: flex;
  align-items: center;
  justify-content: space-around;
  gap: 8px;
  position: relative;
}
.sidebar-actions-row.stacked {
  flex-direction: column;
  gap: 12px;
}

.action-circle-btn {
  background: var(--btn-ghost-bg);
  border: 1px solid var(--border);
  color: var(--text-secondary);
  width: 36px;
  height: 36px;
  border-radius: 50%;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.2s;
}
.action-circle-btn:hover {
  background: var(--btn-ghost-hover-bg);
  color: var(--primary);
  border-color: var(--primary);
}

.sidebar-user-block {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 4px;
  gap: 8px;
}

.user-profile-chip {
  display: flex;
  align-items: center;
  gap: 10px;
  text-decoration: none;
  color: inherit;
  min-width: 0;
}

.user-avatar-circle {
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: var(--primary);
  color: var(--espresso);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 15px;
  font-weight: 800;
  box-shadow: var(--shadow-sm);
  flex-shrink: 0;
}
[data-theme='latte'] .user-avatar-circle {
  color: var(--latte-foam);
}

.user-text-meta {
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.user-meta-name {
  font-size: 13px;
  font-weight: 700;
  color: var(--text-primary);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.logout-door-btn {
  background: transparent;
  border: none;
  font-size: 16px;
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: transform 0.2s;
}
.logout-door-btn:hover {
  transform: scale(1.1);
}

/* Position notification dropdown correctly next to the sidebar */
.app-sidebar :deep(.notif-dropdown) {
  left: 56px;
  right: auto;
  top: auto;
  bottom: 0;
}

/* ── Mobile Responsive Overrides ── */
@media (max-width: 768px) {
  .app-sidebar {
    position: fixed;
    left: -240px;
    top: 0;
    bottom: 0;
    height: 100vh;
    width: 220px;
    z-index: 200;
    border-radius: 0;
    box-shadow: var(--shadow-xl);
  }
  .app-sidebar.expanded {
    left: 0;
  }
  
  .app-sidebar.is-local {
    position: fixed;
    left: -240px;
    top: 0;
    bottom: 0;
    height: 100vh;
    width: 220px;
    z-index: 200;
    border-radius: 0;
    box-shadow: var(--shadow-xl);
  }
  .app-sidebar.is-local.expanded {
    left: 0;
  }
}
</style>
