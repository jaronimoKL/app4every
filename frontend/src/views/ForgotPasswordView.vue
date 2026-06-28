<template>
  <div class="auth-page">
    <div class="orb" style="width:500px;height:500px;background:#6366f1;top:-150px;left:-150px;"></div>
    <div class="orb" style="width:300px;height:300px;background:#06b6d4;bottom:-100px;right:-50px;"></div>
    <div class="dot-grid"></div>

    <div class="auth-container">
      <div class="auth-brand">
        <RouterLink to="/" class="flex items-center gap-3" style="text-decoration:none;color:inherit;">
          <div class="logo-mark">⬡</div>
          <span style="font-weight:700;font-size:18px;">App4Every</span>
        </RouterLink>
      </div>

      <div class="auth-card glass">
        <!-- Состояние после отправки -->
        <template v-if="sent">
          <div class="success-state">
            <div class="success-icon">📬</div>
            <h1 class="auth-title">Письмо отправлено</h1>
            <p class="auth-subtitle" style="margin-top:8px;line-height:1.6;">
              Если аккаунт с логином или email<br>
              <strong style="color:var(--text-primary);">{{ sentIdentifier }}</strong><br>
              существует — мы отправили инструкции по сбросу пароля.
            </p>
            <div class="alert-stub glass" style="margin-top:16px;">
              <span>🔧</span>
              <span>Заглушка: SMTP не подключён. Сброс пароля будет доступен после интеграции с почтовым сервисом.</span>
            </div>
            <RouterLink to="/login" class="btn btn-ghost" style="width:100%;padding:12px;margin-top:20px;text-align:center;">
              ← Вернуться к входу
            </RouterLink>
          </div>
        </template>

        <!-- Форма -->
        <template v-else>
          <div class="auth-card-header">
            <div style="font-size:36px;margin-bottom:12px;">🔐</div>
            <h1 class="auth-title">Забыли пароль?</h1>
            <p class="auth-subtitle">Введите логин или email — пришлём ссылку для сброса на связанную почту</p>
          </div>

          <form @submit.prevent="handleSubmit" class="auth-form">
            <div class="form-group">
              <label class="form-label" for="fp-identifier">Email или Логин</label>
              <input
                id="fp-identifier"
                v-model="identifier"
                type="text"
                class="form-input"
                placeholder="you@example.com или логин"
                autocomplete="username"
                required
              />
            </div>

            <button type="submit" class="btn btn-primary" style="width:100%;padding:13px;" :disabled="loading">
              <span v-if="loading" class="spinner"></span>
              <span>{{ loading ? 'Отправляем...' : 'Отправить ссылку' }}</span>
            </button>
          </form>

          <p class="auth-switch">
            Вспомнили пароль?
            <RouterLink to="/login">Войти →</RouterLink>
          </p>
        </template>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { authApi } from '@/services/api'

const identifier     = ref('')
const sentIdentifier = ref('')
const loading   = ref(false)
const sent      = ref(false)

async function handleSubmit() {
  loading.value = true
  try {
    await authApi.forgotPassword(identifier.value)
  } catch { /* намеренно игнорируем — не раскрываем существование email */ }
  finally {
    sentIdentifier.value = identifier.value
    loading.value = false
    sent.value = true
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
.auth-switch {
  text-align: center;
  font-size: 13px;
  color: var(--text-secondary);
}
.auth-switch a { color: #a5b4fc; text-decoration: none; font-weight: 500; }
.auth-switch a:hover { color: white; }
.success-state { display: flex; flex-direction: column; align-items: center; text-align: center; gap: 8px; }
.success-icon { font-size: 48px; line-height: 1; margin-bottom: 8px; }
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
