import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useNotificationStore = defineStore('notifications', () => {
  const notifications = ref([])
  const unreadCount = ref(0)
  const isConnected = ref(false)
  let ws = null

  const connect = (token) => {
    if (ws) {
      ws.close()
    }
    const wsUrl = `ws://localhost:8080/api/v1/auth/ws/notifications?token=${token}`
    ws = new WebSocket(wsUrl)

    ws.onopen = () => {
      isConnected.value = true
      console.log('Notifications WS connected')
    }

    ws.onmessage = (event) => {
      try {
        const notif = JSON.parse(event.data)
        notifications.value.unshift(notif)
        unreadCount.value++
      } catch (err) {
        console.error('Failed to parse notification:', err)
      }
    }

    ws.onclose = () => {
      isConnected.value = false
      console.log('Notifications WS closed')
    }
  }

  const disconnect = () => {
    if (ws) {
      ws.close()
      ws = null
    }
    notifications.value = []
    unreadCount.value = 0
  }

  const fetchHistory = async (token) => {
    try {
      const resp = await fetch('http://localhost:8080/api/v1/auth/notifications', {
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })
      if (resp.ok) {
        const data = await resp.json()
        notifications.value = data || []
        unreadCount.value = notifications.value.filter(n => !n.read).length
      }
    } catch (err) {
      console.error('Failed to fetch notifications history:', err)
    }
  }

  const markAsRead = async (id, token) => {
    try {
      const resp = await fetch(`http://localhost:8080/api/v1/auth/notifications/${id}/read`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })
      if (resp.ok) {
        const notif = notifications.value.find(n => n.id === id)
        if (notif && !notif.read) {
          notif.read = true
          unreadCount.value = Math.max(0, unreadCount.value - 1)
        }
      }
    } catch (err) {
      console.error('Failed to mark notification as read:', err)
    }
  }

  const markAllAsRead = async (token) => {
    try {
      const resp = await fetch('http://localhost:8080/api/v1/auth/notifications/read-all', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })
      if (resp.ok) {
        notifications.value.forEach(n => n.read = true)
        unreadCount.value = 0
      }
    } catch (err) {
      console.error('Failed to mark all notifications as read:', err)
    }
  }

  const deleteNotification = async (id, token) => {
    try {
      const resp = await fetch(`http://localhost:8080/api/v1/auth/notifications/${id}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${token}`
        }
      })
      if (resp.ok) {
        const index = notifications.value.findIndex(n => n.id === id)
        if (index !== -1) {
          const notif = notifications.value[index]
          if (!notif.read) {
            unreadCount.value = Math.max(0, unreadCount.value - 1)
          }
          notifications.value.splice(index, 1)
        }
      }
    } catch (err) {
      console.error('Failed to delete notification:', err)
    }
  }

  return {
    notifications,
    unreadCount,
    isConnected,
    connect,
    disconnect,
    fetchHistory,
    markAsRead,
    markAllAsRead,
    deleteNotification
  }
})
