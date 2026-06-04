import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi, userApi } from '@/services/api'
import router from '@/router'

export const useAuthStore = defineStore('auth', () => {
  // ── Состояние ──
  const accessToken = ref(null)  // JWT в памяти — безопаснее localStorage (XSS)
  const user = ref(null)
  const loading = ref(false)

  // ── Геттеры ──
  const isAuthenticated = computed(() => !!accessToken.value)

  // ── Действия ──

  async function register(username, email, password) {
    loading.value = true
    try {
      await authApi.register(username, email, password)
      return { success: true }
    } catch (err) {
      return { success: false, error: err }
    } finally {
      loading.value = false
    }
  }

  async function login(email, password) {
    loading.value = true
    try {
      const data = await authApi.login(email, password)
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

  function clearAuth() {
    accessToken.value = null
    user.value = null
  }

  return {
    accessToken, user, loading,
    isAuthenticated,
    register, login, logout, refresh, tryRestoreSession,
    updateProfile, changePassword,
    clearAuth,
  }
})
