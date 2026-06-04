<template>
  <div class="auth-page">
    <div class="orb" style="width:500px;height:500px;background:#6366f1;top:-150px;right:-150px;"></div>
    <div class="orb" style="width:300px;height:300px;background:#8b5cf6;bottom:-100px;left:-50px;"></div>
    <div class="dot-grid"></div>

    <div class="auth-container">
      <div class="auth-brand">
        <RouterLink to="/" class="flex items-center gap-3" style="text-decoration:none;color:inherit;">
          <div class="logo-mark">⬡</div>
          <span style="font-weight:700;font-size:18px;">App4Every</span>
        </RouterLink>
      </div>

      <div class="auth-card glass">
        <!-- Успех -->
        <template v-if="done">
          <div class="success-state">
            <div class="success-icon">✅</div>
            <h1 class="auth-title">Пароль изменён!</h1>
            <p class="auth-subtitle" style="margin-top:8px;">
              Теперь войдите с новым паролем.
            </p>
            <RouterLink to="/login" class="btn btn-primary" style="width:100%;padding:13px;margin-top:20px;text-align:center;">
              Войти →
            </RouterLink>
          </div>
        </template>

        <!-- Форма -->
        <template v-else>
          <div class="auth-card-header">
            <div style="font-size:36px;margin-bottom:12px;">🔑</div>
            <h1 class="auth-title">Новый пароль</h1>
            <p class="auth-subtitle">Придумайте новый пароль для аккаунта</p>
          </div>

          <!-- Заглушка (показываем только если нет токена) -->
          <div v-if="!token" class="alert-stub glass">
            <span>🔧</span>
            <span>
              Заглушка: для полноценной работы нужна ссылка из письма (?token=...).
              SMTP-интеграция будет добавлена позже.
            </span>
          </div>

          <div v-if="errorMsg" class="alert-error">
            <span>⚠</span> {{ errorMsg }}
          </div>

          <form @submit.prevent="handleSubmit" class="auth-form">
            <div class="form-group">
              <label class="form-label" for="rp-new">Новый пароль</label>
              <input
                id="rp-new"
                v-model="newPassword"
                type="password"
                class="form-input"
                :class="{ error: fieldErrors.newPassword }"
                placeholder="Минимум 8 символов"
                autocomplete="new-password"
                required
              />
              <span v-if="fieldErrors.newPassword" class="field-error">{{ fieldErrors.newPassword }}</span>
            </div>
            <div class="form-group">
              <label class="form-label" for="rp-confirm">Подтверждение</label>
              <input
                id="rp-confirm"
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

            <button type="submit" class="btn btn-primary" style="width:100%;padding:13px;" :disabled="loading">
              <span v-if="loading" class="spinner"></span>
              <span>{{ loading ? 'Сохраняем...' : 'Установить новый пароль' }}</span>
            </button>
          </form>

          <p class="auth-switch">
            <RouterLink to="/login">← Вернуться к входу</RouterLink>
          </p>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { authApi } from '@/services/api'

const route = useRoute()

// Токен берём из URL: /reset-password?token=xxxxx
const token      = ref('')
const newPassword = ref('')
const confirm    = ref('')
const errorMsg   = ref('')
const loading    = ref(false)
const done       = ref(false)
const fieldErrors = reactive({ newPassword: '', confirm: '' })

onMounted(() => {
  token.value = route.query.token || ''
})

function validate() {
  fieldErrors.newPassword = ''
  fieldErrors.confirm     = ''
  let ok = true
  if (newPassword.value.length < 8) { fieldErrors.newPassword = 'Минимум 8 символов'; ok = false }
  if (newPassword.value !== confirm.value) { fieldErrors.confirm = 'Пароли не совпадают'; ok = false }
  return ok
}

async function handleSubmit() {
  errorMsg.value = ''
  if (!validate()) return

  loading.value = true
  try {
    // Если токена нет (заглушка без письма) — всё равно вызываем API для демонстрации
    await authApi.resetPassword(token.value || 'stub-token', newPassword.value)
    done.value = true
  } catch {
    errorMsg.value = 'Ссылка недействительна или устарела. Запросите новую.'
  } finally {
    loading.value = false
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
}
.auth-switch a { color: #a5b4fc; text-decoration: none; font-weight: 500; }
.success-state { display: flex; flex-direction: column; align-items: center; text-align: center; gap: 8px; }
.success-icon { font-size: 48px; margin-bottom: 8px; }
.alert-stub {
  border-radius: var(--radius-md);
  padding: 10px 14px;
  font-size: 12px;
  color: var(--text-secondary);
  display: flex;
  align-items: flex-start;
  gap: 8px;
  border: 1px solid rgba(255,255,255,0.06);
}
</style>
