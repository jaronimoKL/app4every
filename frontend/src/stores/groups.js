import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useAuthStore } from '@/stores/auth'

const BASE = '/api/v1'

async function api(method, path, body = null, token = null) {
  const headers = { 'Content-Type': 'application/json' }
  if (token) headers['Authorization'] = `Bearer ${token}`
  const res = await fetch(`${BASE}${path}`, {
    method, headers,
    credentials: 'include',
    body: body ? JSON.stringify(body) : null,
  })
  if (res.status === 204) return null
  const data = await res.json()
  if (!res.ok) throw { status: res.status, ...data }
  return data
}

export const useGroupsStore = defineStore('groups', () => {
  const groups = ref([])
  const invites = ref([])
  const activeGroup = ref(null)
  const loading = ref(false)
  const saving = ref(false)
  const ws = ref(null)

  function token() { return useAuthStore().accessToken }

  // ── Группы (REST) ──

  async function fetchGroups() {
    loading.value = true
    try {
      const res = await api('GET', '/groups', null, token())
      groups.value = res || []
    } finally {
      loading.value = false
    }
  }

  async function createGroup(name, inviteIds = []) {
    saving.value = true
    try {
      const g = await api('POST', '/groups', { name, invite_ids: inviteIds }, token())
      groups.value.unshift(g)
      return g
    } finally {
      saving.value = false
    }
  }

  async function fetchGroupDetail(id) {
    loading.value = true
    try {
      activeGroup.value = await api('GET', `/groups/${id}`, null, token())
    } finally {
      loading.value = false
    }
  }

  async function deleteGroup(id) {
    await api('DELETE', `/groups/${id}`, null, token())
    groups.value = groups.value.filter(g => g.id !== id)
    if (activeGroup.value && activeGroup.value.id === id) {
      activeGroup.value = null
      disconnectWS()
    }
  }

  async function inviteUser(groupId, identifier) {
    return await api('POST', `/groups/${groupId}/invite`, { identifier }, token())
  }

  async function leaveGroup(groupId) {
    await api('POST', `/groups/${groupId}/leave`, null, token())
    groups.value = groups.value.filter(g => g.id !== groupId)
    if (activeGroup.value && activeGroup.value.id === groupId) {
      activeGroup.value = null
      disconnectWS()
    }
  }

  // ── Приглашения (REST) ──

  async function fetchInvites() {
    invites.value = await api('GET', '/groups/invites', null, token())
  }

  async function acceptInvite(inviteId) {
    const res = await api('POST', `/groups/invites/${inviteId}/accept`, null, token())
    invites.value = invites.value.filter(inv => inv.id !== inviteId)
    await fetchGroups()
    return res.group_id
  }

  async function declineInvite(inviteId) {
    await api('POST', `/groups/invites/${inviteId}/decline`, null, token())
    invites.value = invites.value.filter(inv => inv.id !== inviteId)
  }

  // ── Записи в группе (REST) ──

  async function addGroupItem(groupId, data) {
    const item = await api('POST', `/groups/${groupId}/items`, data, token())
    if (activeGroup.value && activeGroup.value.id === groupId) {
      const idx = activeGroup.value.items.findIndex(it => it.id === item.id)
      if (idx === -1) {
        activeGroup.value.items.unshift(item)
      }
    }
    return item
  }

  async function updateGroupItem(groupId, itemId, data) {
    const item = await api('PUT', `/groups/${groupId}/items/${itemId}`, data, token())
    if (activeGroup.value && activeGroup.value.id === groupId) {
      const idx = activeGroup.value.items.findIndex(it => it.id === itemId)
      if (idx !== -1) {
        activeGroup.value.items[idx] = item
      }
    }
    return item
  }

  async function deleteGroupItem(groupId, itemId) {
    await api('DELETE', `/groups/${groupId}/items/${itemId}`, null, token())
    if (activeGroup.value && activeGroup.value.id === groupId) {
      activeGroup.value.items = activeGroup.value.items.filter(it => it.id !== itemId)
    }
  }

  async function rateGroupItem(groupId, itemId, rating) {
    const item = await api('POST', `/groups/${groupId}/items/${itemId}/rating`, { rating }, token())
    if (activeGroup.value && activeGroup.value.id === groupId) {
      const idx = activeGroup.value.items.findIndex(it => it.id === itemId)
      if (idx !== -1) {
        activeGroup.value.items[idx] = item
      }
    }
    return item
  }

  async function addGroupItemLink(groupId, itemId, label, url) {
    const item = await api('POST', `/groups/${groupId}/items/${itemId}/links`, { label, url }, token())
    if (activeGroup.value && activeGroup.value.id === groupId) {
      const idx = activeGroup.value.items.findIndex(it => it.id === itemId)
      if (idx !== -1) {
        activeGroup.value.items[idx] = item
      }
    }
    return item
  }

  async function deleteGroupItemLink(groupId, itemId, linkId) {
    const item = await api('DELETE', `/groups/${groupId}/items/${itemId}/links/${linkId}`, null, token())
    if (activeGroup.value && activeGroup.value.id === groupId) {
      const idx = activeGroup.value.items.findIndex(it => it.id === itemId)
      if (idx !== -1) {
        activeGroup.value.items[idx] = item
      }
    }
    return item
  }

  // ── WebSocket Соединение ──

  function connectWS(groupId) {
    disconnectWS()

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/api/v1/groups/${groupId}/ws?token=${token()}`
    
    const socket = new WebSocket(wsUrl)
    ws.value = socket

    socket.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        if (!activeGroup.value || activeGroup.value.id !== groupId) return

        if (msg.event === 'item_added') {
          // Исключаем дубликаты
          const idx = activeGroup.value.items.findIndex(it => it.id === msg.data.id)
          if (idx === -1) {
            activeGroup.value.items.unshift(msg.data)
          }
        } else if (msg.event === 'item_updated') {
          const idx = activeGroup.value.items.findIndex(it => it.id === msg.data.id)
          if (idx !== -1) {
            activeGroup.value.items[idx] = msg.data
          }
        } else if (msg.event === 'item_deleted') {
          activeGroup.value.items = activeGroup.value.items.filter(it => it.id !== msg.data.id)
        }
      } catch (err) {
        console.error('WS parse error:', err)
      }
    }

    socket.onclose = () => {
      console.log(`WS connection closed for group ${groupId}`)
    }

    socket.onerror = (err) => {
      console.error('WS error:', err)
    }
  }

  function disconnectWS() {
    if (ws.value) {
      ws.value.close()
      ws.value = null
    }
  }

  return {
    groups, invites, activeGroup, loading, saving,
    fetchGroups, createGroup, fetchGroupDetail, deleteGroup,
    inviteUser, leaveGroup,
    fetchInvites, acceptInvite, declineInvite,
    addGroupItem, updateGroupItem, deleteGroupItem,
    rateGroupItem, addGroupItemLink, deleteGroupItemLink,
    connectWS, disconnectWS,
  }
})
