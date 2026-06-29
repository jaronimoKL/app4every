import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi, userApi } from '@/services/api'
import router from '@/router'

export const useAuthStore = defineStore('auth', () => {
  // ── Состояние ──
  const accessToken = ref(null)  // JWT в памяти — безопаснее localStorage (XSS)
  const user = ref(null)
  const loading = ref(false)

  // initialized = true после первой попытки восстановить сессию.
  // _initPromise обеспечивает что tryRestoreSession вызывается ровно один раз.
  const initialized = ref(false)
  let _initPromise = null

  // ── Геттеры ──
  const isAuthenticated = computed(() => !!accessToken.value)

  // ── Действия ──

  async function register(username, email, password, inviteCode) {
    loading.value = true
    try {
      await authApi.register(username, email, password, inviteCode)
      return { success: true }
    } catch (err) {
      return { success: false, error: err }
    } finally {
      loading.value = false
    }
  }

  async function login(identifier, password) {
    loading.value = true
    try {
      const data = await authApi.login(identifier, password)
      accessToken.value = data.access_token
      user.value = data.user
      return { success: true }
    } catch (err) {
      return { success: false, error: err }
    } finally {
      loading.value = false
    }
  }

  async function refresh() {
    try {
      const data = await authApi.refresh()
      accessToken.value = data.access_token
      return true
    } catch {
      clearAuth()
      return false
    }
  }

  async function logout() {
    try {
      await authApi.logout(accessToken.value)
    } catch { /* даже если упало — чистим */ } finally {
      clearAuth()
      router.push({ name: 'landing' })
    }
  }

  // Восстановление сессии при перезагрузке страницы через refresh_token из куки
  async function tryRestoreSession() {
    const ok = await refresh()
    if (ok) {
      try {
        const userData = await authApi.me(accessToken.value)
        user.value = userData
      } catch {
        clearAuth()
      }
    }
    return ok
  }

  /**
   * ensureInitialized() — вызывается из router.beforeEach перед каждой навигацией.
   * Гарантирует что сессия восстановлена из HttpOnly refresh-cookie ровно один раз.
   * После первого вызова возвращается мгновенно.
   */
  async function ensureInitialized() {
    if (initialized.value) return
    if (!_initPromise) {
      _initPromise = tryRestoreSession().finally(() => {
        initialized.value = true
      })
    }
    return _initPromise
  }

  // Обновление профиля (username + email)
  async function updateProfile(username, email) {
    try {
      const updatedUser = await userApi.updateProfile(username, email, accessToken.value)
      user.value = updatedUser
      return { success: true }
    } catch (err) {
      return { success: false, error: err }
    }
  }

  // Смена пароля
  async function changePassword(currentPassword, newPassword) {
    try {
      await userApi.changePassword(currentPassword, newPassword, accessToken.value)
      return { success: true }
    } catch (err) {
      return { success: false, error: err }
    }
  }

  // Интеграция Shikimori
  async function fetchShikimoriWhoami() {
    try {
      const res = await fetch('/api/v1/auth/shikimori/whoami', {
        headers: { 'Authorization': `Bearer ${accessToken.value}` }
      })
      if (!res.ok) throw new Error('Failed to fetch whoami')
      return await res.json()
    } catch {
      return null
    }
  }

  async function unlinkShikimori() {
    try {
      const res = await fetch('/api/v1/auth/shikimori/unlink', {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${accessToken.value}` }
      })
      if (!res.ok) throw new Error('Failed to unlink')
      // Обновляем локального пользователя чтобы убрать ID
      if (user.value) user.value.shikimori_user_id = null
      return true
    } catch {
      return false
    }
  }

  function clearAuth() {
    accessToken.value = null
    user.value = null
  }

  return {
    accessToken, user, loading,
    isAuthenticated, initialized,
    register, login, logout, refresh, tryRestoreSession, ensureInitialized,
    updateProfile, changePassword,
    fetchShikimoriWhoami, unlinkShikimori,
    clearAuth,
  }
})
