<template>
  <div class="notification-wrapper" @click.stop="toggleDropdown">
    <button class="btn btn-ghost notif-btn relative" title="Уведомления">
      <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
        <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"></path>
        <path d="M13.73 21a2 2 0 0 1-3.46 0"></path>
      </svg>
      <span v-if="notifStore.unreadCount > 0" class="badge">
        {{ notifStore.unreadCount }}
      </span>
    </button>

    <div v-if="isOpen" class="notif-dropdown glass" @click.stop>
      <div class="notif-header">
        <h3 style="font-size:14px;font-weight:600;">Уведомления</h3>
        <button v-if="notifStore.unreadCount > 0" class="text-btn" @click="notifStore.markAllAsRead(authStore.token)">
          Прочитать все
        </button>
      </div>

      <div class="notif-list">
        <div v-if="notifStore.notifications.length === 0" class="empty-state">
          Нет новых уведомлений
        </div>
        <div 
          v-else
          v-for="item in notifStore.notifications" 
          :key="item.id" 
          class="notif-item" 
          :class="{ 'is-unread': !item.read }"
          @click="handleNotificationClick(item)"
        >
          <div class="notif-content">
            <span class="notif-msg">{{ item.message }}</span>
            <span class="notif-time">{{ new Date(item.created_at).toLocaleString() }}</span>
          </div>
          <button class="close-btn" @click.stop="notifStore.deleteNotification(item.id, authStore.token)">
            <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
              <line x1="18" y1="6" x2="6" y2="18"></line>
              <line x1="6" y1="6" x2="18" y2="18"></line>
            </svg>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { useNotificationStore } from '../stores/notifications'

const isOpen = ref(false)
const authStore = useAuthStore()
const notifStore = useNotificationStore()
const router = useRouter()

const toggleDropdown = () => {
  isOpen.value = !isOpen.value
}

const closeDropdown = () => {
  isOpen.value = false
}

onMounted(() => {
  document.addEventListener('click', closeDropdown)
  if (authStore.token) {
    notifStore.fetchHistory(authStore.token)
    notifStore.connect(authStore.token)
  }
})

onUnmounted(() => {
  document.removeEventListener('click', closeDropdown)
  notifStore.disconnect()
})

const handleNotificationClick = (item) => {
  if (!item.read) {
    notifStore.markAsRead(item.id, authStore.token)
  }
  
  if (item.type === 'watchparty_room_created' && item.metadata?.room_id) {
    router.push(`/watchparty?room=${item.metadata.room_id}`)
  } else if (item.type === 'group_invite' && item.metadata?.group_id) {
    router.push(`/reviews/${item.metadata.group_id}`)
  }
  
  isOpen.value = false
}
</script>

<style scoped>
.notification-wrapper {
  position: relative;
}

.notif-btn {
  padding: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 50%;
}

.badge {
  position: absolute;
  top: -2px;
  right: -2px;
  background: #f43f5e;
  color: white;
  font-size: 10px;
  font-weight: bold;
  padding: 2px 5px;
  border-radius: 10px;
  min-width: 16px;
  text-align: center;
}

.notif-dropdown {
  position: absolute;
  top: calc(100% + 8px);
  right: 0;
  width: 320px;
  background: var(--bg-surface);
  border-radius: var(--radius-md);
  box-shadow: 0 10px 40px rgba(0,0,0,0.3);
  z-index: 100;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.notif-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border);
}

.text-btn {
  background: none;
  border: none;
  color: var(--primary);
  font-size: 12px;
  cursor: pointer;
}

.text-btn:hover {
  text-decoration: underline;
}

.notif-list {
  max-height: 400px;
  overflow-y: auto;
}

.empty-state {
  padding: 24px;
  text-align: center;
  color: var(--text-muted);
  font-size: 13px;
}

.notif-item {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border);
  cursor: pointer;
  transition: background 0.2s;
}

.notif-item:hover {
  background: rgba(255, 255, 255, 0.02);
}

.notif-item.is-unread {
  background: rgba(99, 102, 241, 0.08);
}

.notif-content {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.notif-msg {
  font-size: 13px;
  color: var(--text-primary);
  line-height: 1.4;
}

.notif-time {
  font-size: 11px;
  color: var(--text-muted);
}

.close-btn {
  background: none;
  border: none;
  color: var(--text-muted);
  cursor: pointer;
  padding: 4px;
  border-radius: 4px;
  transition: all 0.2s;
}

.close-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-primary);
}
</style>
