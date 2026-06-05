<template>
  <div class="auth-page">
    <div class="orb" style="width:600px;height:600px;background:#8b5cf6;top:-200px;right:-200px;"></div>
    <div class="orb" style="width:400px;height:400px;background:#6366f1;bottom:-100px;left:-100px;"></div>
    <div class="dot-grid"></div>

    <div class="auth-container">
      <div class="auth-brand">
        <RouterLink to="/" class="flex items-center gap-3" style="text-decoration:none;color:inherit;">
          <div class="logo-mark">⬡</div>
          <span style="font-weight:700;font-size:18px;">App4Every</span>
        </RouterLink>
      </div>

      <div class="auth-card glass">
        <!-- Успешная регистрация -->
        <template v-if="success">
          <div class="success-state">
            <div class="success-icon">✉️</div>
            <h1 class="auth-title">Проверьте почту</h1>
            <p class="auth-subtitle" style="margin-top:8px;">
              Аккаунт создан! Мы отправили письмо на<br>
              <strong style="color:var(--text-primary);">{{ email }}</strong><br>
              для подтверждения адреса.
            </p>
            <div class="alert-stub glass" style="margin-top:16px;">
              <span>🔧</span>
              <span>Заглушка: отправка email будет подключена позже. Вы уже можете войти.</span>
            </div>
            <RouterLink to="/login" class="btn btn-primary" style="width:100%;padding:13px;margin-top:20px;text-align:center;">
              Войти в аккаунт →
            </RouterLink>
          </div>
        </template>

        <!-- Форма регистрации -->
        <template v-else>
          <div class="auth-card-header">
            <h1 class="auth-title">Создать аккаунт</h1>
            <p class="auth-subtitle">Присоединитесь к экосистеме</p>
          </div>

          <div v-if="errorMsg" class="alert-error">
            <span>⚠</span> {{ errorMsg }}
          </div>

          <form @submit.prevent="handleRegister" class="auth-form">
            <div class="form-group">
              <label class="form-label" for="username">Логин</label>
              <input
                id="username"
                v-model="username"
                type="text"
                class="form-input"
                :class="{ error: fieldErrors.username }"
                placeholder="your_username"
                autocomplete="username"
                required
              />
              <span v-if="fieldErrors.username" class="field-error">{{ fieldErrors.username }}</span>
            </div>

            <div class="form-group">
              <label class="form-label" for="password">Пароль</label>
              <input
                id="password"
                v-model="password"
                type="password"
                class="form-input"
                :class="{ error: fieldErrors.password }"
                placeholder="Минимум 8 символов"
                autocomplete="new-password"
                required
              />
              <span v-if="fieldErrors.password" class="field-error">{{ fieldErrors.password }}</span>
            </div>

            <div class="form-group">
              <label class="form-label" for="confirm">Подтверждение пароля</label>
              <input
                id="confirm"
                v-model="confirm"
                type="password"
                class="form-input"
                :class="{ error: fieldErrors.confirm }"
                placeholder="Повторите пароль"
                autocomplete="new-password"
                required
              />
              <span v-if="fieldErrors.confirm" class="field-error">{{ fieldErrors.confirm }}</span>
            </div>

            <div class="form-group">
              <label class="form-label" for="email">Email</label>
              <input
                id="email"
                v-model="email"
                type="email"
                class="form-input"
                :class="{ error: fieldErrors.email }"
                placeholder="you@example.com"
                autocomplete="email"
                required
              />
              <span v-if="fieldErrors.email" class="field-error">{{ fieldErrors.email }}</span>
            </div>

            <button
              type="submit"
              class="btn btn-primary"
              style="width:100%;padding:13px;"
              :disabled="loading"
            >
              <span v-if="loading" class="spinner"></span>
              <span>{{ loading ? 'Создаём аккаунт...' : 'Зарегистрироваться' }}</span>
            </button>
          </form>

          <p class="auth-switch">
            Уже есть аккаунт?
            <RouterLink to="/login">Войти →</RouterLink>
          </p>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive } from 'vue'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()

const username = ref('')
const email    = ref('')
const password = ref('')
const confirm  = ref('')
const errorMsg = ref('')
const success  = ref(false)
const loading  = ref(false)

const fieldErrors = reactive({ username: '', email: '', password: '', confirm: '' })

function validate() {
  fieldErrors.username = ''
  fieldErrors.email    = ''
  fieldErrors.password = ''
  fieldErrors.confirm  = ''
  let ok = true

  if (username.value.length < 3) {
    fieldErrors.username = 'Минимум 3 символа'
    ok = false
  }
  if (!/^[a-zA-Z0-9_]+$/.test(username.value)) {
    fieldErrors.username = 'Только латиница, цифры и подчёркивание'
    ok = false
  }
  if (!email.value) {
    fieldErrors.email = 'Введите email'
    ok = false
  } else if (!email.value.includes('@')) {
    fieldErrors.email = 'Введите корректный email'
    ok = false
  }
  if (password.value.length < 8) {
    fieldErrors.password = 'Минимум 8 символов'
    ok = false
  }
  if (password.value !== confirm.value) {
    fieldErrors.confirm = 'Пароли не совпадают'
    ok = false
  }
  return ok
}

async function handleRegister() {
  errorMsg.value = ''
  if (!validate()) return

  loading.value = true
  const result = await auth.register(username.value, email.value, password.value)
  loading.value = false

  if (result.success) {
    success.value = true
  } else {
    const status = result.error?.status
    if (status === 409) {
      errorMsg.value = 'Этот email или логин уже занят'
    } else {
      errorMsg.value = 'Ошибка регистрации. Попробуйте позже.'
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
.auth-brand { display: flex; justify-content: center; }
.auth-card {
  border-radius: var(--radius-xl);
  padding: 36px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}
.auth-card-header { text-align: center; }
.auth-title {
  font-size: 24px;
  font-weight: 700;
  letter-spacing: -0.02em;
  margin-bottom: 6px;
}
.auth-subtitle { font-size: 14px; color: var(--text-secondary); }
.auth-form { display: flex; flex-direction: column; gap: 16px; }
.field-error { font-size: 12px; color: #f87171; margin-top: 2px; }
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

/* Состояние успеха */
.success-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: 8px;
}
.success-icon { font-size: 48px; line-height: 1; margin-bottom: 8px; }

/* Заглушка */
.alert-stub {
  border-radius: var(--radius-md);
  padding: 10px 14px;
  font-size: 12px;
  color: var(--text-secondary);
  display: flex;
  align-items: flex-start;
  gap: 8px;
  text-align: left;
  border: 1px solid rgba(255,255,255,0.06);
}
</style>
