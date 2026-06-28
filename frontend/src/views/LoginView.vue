<template>
  <div class="auth-page">
    <!-- Фоновые шары -->
    <div class="orb" style="width:600px;height:600px;background:#6366f1;top:-200px;left:-200px;"></div>
    <div class="orb" style="width:400px;height:400px;background:#8b5cf6;bottom:-100px;right:-100px;"></div>
    <div class="dot-grid"></div>

    <div class="auth-container">
      <!-- Логотип и заголовок -->
      <div class="auth-brand">
        <RouterLink to="/" class="flex items-center gap-3" style="text-decoration:none;color:inherit;">
          <div class="logo-mark">⬡</div>
          <span style="font-weight:700;font-size:18px;">App4Every</span>
        </RouterLink>
      </div>

      <!-- Карточка формы -->
      <div class="auth-card glass">
        <div class="auth-card-header">
          <h1 class="auth-title">Добро пожаловать</h1>
          <p class="auth-subtitle">Войдите в свою экосистему</p>
        </div>

        <!-- Алерт ошибки -->
        <div v-if="errorMsg" class="alert-error">
          <span>⚠</span>
          {{ errorMsg }}
        </div>

        <!-- Форма -->
        <form @submit.prevent="handleLogin" class="auth-form">
          <div class="form-group">
            <label class="form-label" for="identifier">Логин или Email</label>
            <input
              id="identifier"
              v-model="identifier"
              type="text"
              class="form-input"
              :class="{ error: !!errorMsg }"
              placeholder="Логин или Email"
              autocomplete="username"
              required
            />
          </div>

          <div class="form-group">
            <div class="flex items-center justify-between">
              <label class="form-label" for="password">Пароль</label>
              <RouterLink to="/forgot-password" class="forgot-link">Забыли пароль?</RouterLink>
            </div>
            <input
              id="password"
              v-model="password"
              type="password"
              class="form-input"
              :class="{ error: !!errorMsg }"
              placeholder="••••••••"
              autocomplete="current-password"
              required
            />
          </div>

          <button
            type="submit"
            class="btn btn-primary"
            style="width:100%;padding:13px;"
            :disabled="loading"
          >
            <span v-if="loading" class="spinner"></span>
            <span>{{ loading ? 'Входим...' : 'Войти' }}</span>
          </button>
        </form>

        <!-- Ссылка на регистрацию -->
        <p class="auth-switch">
          Нет аккаунта?
          <RouterLink to="/register">Зарегистрироваться →</RouterLink>
        </p>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const auth = useAuthStore()

const identifier = ref('')
const password   = ref('')
const errorMsg   = ref('')
const loading    = ref(false)

async function handleLogin() {
  errorMsg.value = ''
  loading.value = true

  const result = await auth.login(identifier.value, password.value)

  loading.value = false

  if (result.success) {
    router.push({ name: 'dashboard' })
  } else {
    const status = result.error?.status
    if (status === 401) {
      errorMsg.value = 'Неверный логин/email или пароль'
    } else {
      errorMsg.value = 'Ошибка сервера. Попробуйте позже.'
    }
  }
}
</script>

<style scoped>
.auth-page {
  min-height: 100vh;
  position: relative;
  overflow: hidden;
  background: var(--bg-base);
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 24px;
}

.auth-container {
  position: relative;
  z-index: 1;
  width: 100%;
  max-width: 420px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.auth-brand {
  display: flex;
  justify-content: center;
}

.auth-card {
  border-radius: var(--radius-xl);
  padding: 36px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

.auth-card-header {
  text-align: center;
}

.auth-title {
  font-size: 24px;
  font-weight: 700;
  letter-spacing: -0.02em;
  margin-bottom: 6px;
}

.auth-subtitle {
  font-size: 14px;
  color: var(--text-secondary);
}

.auth-form {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.auth-switch {
  text-align: center;
  font-size: 13px;
  color: var(--text-secondary);
}

.auth-switch a {
  color: #a5b4fc;
  text-decoration: none;
  font-weight: 500;
  transition: color 0.2s;
}
.auth-switch a:hover { color: white; }

.forgot-link {
  font-size: 12px;
  color: var(--text-muted);
  text-decoration: none;
  transition: color 0.2s;
}
.forgot-link:hover { color: #a5b4fc; }
</style>
