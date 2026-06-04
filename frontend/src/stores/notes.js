import { defineStore } from 'pinia'
import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'

const BASE = '/api/v1'

async function apiRequest(method, path, body = null, token = null) {
  const headers = { 'Content-Type': 'application/json' }
  if (token) headers['Authorization'] = `Bearer ${token}`
  const res = await fetch(`${BASE}${path}`, {
    method,
    headers,
    credentials: 'include',
    body: body ? JSON.stringify(body) : null,
  })
  if (res.status === 204) return null
  const data = await res.json()
  if (!res.ok) throw { status: res.status, ...data }
  return data
}

export const useNotesStore = defineStore('notes', () => {
  const notes     = ref([])
  const active    = ref(null)   // выбранная заметка в редакторе
  const loading   = ref(false)
  const saving    = ref(false)

  function token() {
    return useAuthStore().accessToken
  }

  async function fetchNotes() {
    loading.value = true
    try {
      notes.value = await apiRequest('GET', '/notes', null, token())
    } finally {
      loading.value = false
    }
  }

  async function createNote() {
    saving.value = true
    try {
      const note = await apiRequest('POST', '/notes', { title: '', content: '' }, token())
      notes.value.unshift(note)  // добавить в начало списка
      active.value = { ...note } // открыть сразу в редакторе
      return note
    } finally {
      saving.value = false
    }
  }

  async function saveNote(id, title, content) {
    saving.value = true
    try {
      const updated = await apiRequest('PUT', `/notes/${id}`, { title, content }, token())
      // Обновить в списке
      const idx = notes.value.findIndex(n => n.id === id)
      if (idx !== -1) notes.value[idx] = updated
      if (active.value?.id === id) active.value = { ...updated }
      return updated
    } finally {
      saving.value = false
    }
  }

  async function deleteNote(id) {
    await apiRequest('DELETE', `/notes/${id}`, null, token())
    notes.value = notes.value.filter(n => n.id !== id)
    if (active.value?.id === id) active.value = null
  }

  function selectNote(note) {
    active.value = note ? { ...note } : null
  }

  return { notes, active, loading, saving, fetchNotes, createNote, saveNote, deleteNote, selectNote }
})
