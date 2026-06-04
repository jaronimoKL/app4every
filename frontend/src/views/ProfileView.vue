<template>
  <div class="profile-page">
    <!-- Навбар -->
    <nav class="navbar glass">
      <div class="navbar-inner">
        <div class="flex items-center gap-3">
          <RouterLink to="/dashboard" class="btn btn-ghost" style="padding:7px 12px;font-size:13px;gap:6px;">
            ← Назад
          </RouterLink>
          <div class="nav-sep"></div>
          <div class="logo-mark" style="width:32px;height:32px;font-size:14px;">⬡</div>
          <span style="font-weight:700;font-size:15px;">Профиль</span>
        </div>
        <button @click="auth.logout()" class="btn btn-ghost" style="padding:7px 14px;font-size:13px;">
          Выйти
        </button>
      </div>
    </nav>

    <main class="profile-main">
      <div class="profile-header">
        <div class="profile-avatar">{{ userInitial }}</div>
        <div>
          <div class="profile-name">{{ auth.user?.username || auth.user?.email }}</div>
          <div style="font-size:13px;color:var(--text-secondary);">{{ auth.user?.email }}</div>
        </div>
      </div>

      <!-- ── Секция 1: Основная информация ── -->
      <section class="profile-card glass">
        <div class="card-header">
          <h2 class="card-title">Основная информация</h2>
          <p class="card-desc">Изменение логина и адреса электронной почты</p>
        </div>

        <div v-if="profileMsg.text" :class="['alert-' + profileMsg.type]">
          <span>{{ profileMsg.type === 'success' ? '✓' : '⚠' }}</span>
          {{ profileMsg.text }}
        </div>

        <form @submit.prevent="handleProfileUpdate" class="card-form">
          <div class="form-group">
            <label class="form-label" for="p-username">Логин</label>
            <input
              id="p-username"
              v-model="profileForm.username"
              type="text"
              class="form-input"
              placeholder="your_username"
            />
          </div>
          <div class="form-group">
            <label class="form-label" for="p-email">Email</label>
            <input
              id="p-email"
              v-model="profileForm.email"
              type="email"
              class="form-input"
              placeholder="you@example.com"
            />
          </div>
          <div class="form-actions">
            <button type="submit" class="btn btn-primary" :disabled="profileLoading">
              <span v-if="profileLoading" class="spinner"></span>
              <span>{{ profileLoading ? 'Сохраняем...' : 'Сохранить изменения' }}</span>
            </button>
          </div>
        </form>
      </section>

      <!-- ── Секция 2: Смена пароля ── -->
      <section class="profile-card glass">
        <div class="card-header">
          <h2 class="card-title">Смена пароля</h2>
          <p class="card-desc">Введите текущий пароль для подтверждения</p>
        </div>

        <div v-if="passwordMsg.text" :class="['alert-' + passwordMsg.type]">
          <span>{{ passwordMsg.type === 'success' ? '✓' : '⚠' }}</span>
          {{ passwordMsg.text }}
        </div>

        <form @submit.prevent="handlePasswordChange" class="card-form">
          <div class="form-group">
            <label class="form-label" for="p-current">Текущий пароль</label>
            <input
              id="p-current"
              v-model="passwordForm.current"
              type="password"
              class="form-input"
              :class="{ error: passwordErrors.current }"
              placeholder="••••••••"
              autocomplete="current-password"
            />
            <span v-if="passwordErrors.current" class="field-error">{{ passwordErrors.current }}</span>
          </div>
          <div class="form-group">
            <label class="form-label" for="p-new">Новый пароль</label>
            <input
              id="p-new"
              v-model="passwordForm.newPass"
              type="password"
              class="form-input"
              :class="{ error: passwordErrors.newPass }"
              placeholder="Минимум 8 символов"
              autocomplete="new-password"
            />
            <span v-if="passwordErrors.newPass" class="field-error">{{ passwordErrors.newPass }}</span>
          </div>
          <div class="form-group">
            <label class="form-label" for="p-confirm">Подтверждение нового пароля</label>
            <input
              id="p-confirm"
              v-model="passwordForm.confirm"
              type="password"
              class="form-input"
              :class="{ error: passwordErrors.confirm }"
              placeholder="Повторите новый пароль"
              autocomplete="new-password"
            />
            <span v-if="passwordErrors.confirm" class="field-error">{{ passwordErrors.confirm }}</span>
          </div>
          <div class="form-actions">
            <button type="submit" class="btn btn-primary" :disabled="passwordLoading">
              <span v-if="passwordLoading" class="spinner"></span>
              <span>{{ passwordLoading ? 'Меняем...' : 'Изменить пароль' }}</span>
            </button>
          </div>
        </form>
      </section>

      <!-- ── Секция 3: Безопасность ── -->
      <section class="profile-card glass">
        <div class="card-header">
          <h2 class="card-title">Безопасность</h2>
          <p class="card-desc">Управление сессиями и аккаунтом</p>
        </div>
        <div class="security-info">
          <div class="security-item">
            <span class="security-icon">🔑</span>
            <div>
              <div style="font-size:14px;font-weight:500;">Сессии</div>
              <div style="font-size:12px;color:var(--text-muted);">Refresh-токены хранятся в Redis. При выходе — сессия удаляется.</div>
            </div>
          </div>
          <div class="security-item">
            <span class="security-icon">🍪</span>
            <div>
              <div style="font-size:14px;font-weight:500;">HttpOnly Cookie</div>
              <div style="font-size:12px;color:var(--text-muted);">Refresh-токен недоступен JS — защита от XSS.</div>
            </div>
          </div>
        </div>
      </section>
    </main>
  </div>
</template>

<script setup>
import { ref, computed, reactive, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()

const userInitial = computed(() =>
  (auth.user?.username || auth.user?.email || '?').charAt(0).toUpperCase()
)

// ── Форма профиля ──
const profileForm = reactive({
  username: '',
  email: '',
})
const profileLoading = ref(false)
const profileMsg = reactive({ text: '', type: 'success' })

// ── Форма пароля ──
const passwordForm = reactive({ current: '', newPass: '', confirm: '' })
const passwordErrors = reactive({ current: '', newPass: '', confirm: '' })
const passwordLoading = ref(false)
const passwordMsg = reactive({ text: '', type: 'success' })

onMounted(() => {
  // Заполняем форму текущими данными пользователя
  profileForm.username = auth.user?.username || ''
  profileForm.email = auth.user?.email || ''
})

async function handleProfileUpdate() {
  profileMsg.text = ''
  if (!profileForm.username || !profileForm.email) return

  profileLoading.value = true
  const result = await auth.updateProfile(profileForm.username, profileForm.email)
  profileLoading.value = false

  if (result.success) {
    profileMsg.text = 'Данные успешно обновлены'
    profileMsg.type = 'success'
  } else {
    const status = result.error?.status
    profileMsg.type = 'error'
    profileMsg.text = status === 409
      ? 'Этот email или логин уже занят'
      : 'Ошибка обновления. Попробуйте позже.'
  }

  // Скрыть сообщение через 4 секунды
  setTimeout(() => { profileMsg.text = '' }, 4000)
}

function validatePassword() {
  passwordErrors.current = ''
  passwordErrors.newPass = ''
  passwordErrors.confirm = ''
  let ok = true
  if (!passwordForm.current) { passwordErrors.current = 'Введите текущий пароль'; ok = false }
  if (passwordForm.newPass.length < 8) { passwordErrors.newPass = 'Минимум 8 символов'; ok = false }
  if (passwordForm.newPass !== passwordForm.confirm) { passwordErrors.confirm = 'Пароли не совпадают'; ok = false }
  return ok
}

async function handlePasswordChange() {
  passwordMsg.text = ''
  if (!validatePassword()) return

  passwordLoading.value = true
  const result = await auth.changePassword(passwordForm.current, passwordForm.newPass)
  passwordLoading.value = false

  if (result.success) {
    passwordMsg.text = 'Пароль успешно изменён'
    passwordMsg.type = 'success'
    passwordForm.current = ''
    passwordForm.newPass = ''
    passwordForm.confirm = ''
  } else {
    passwordMsg.type = 'error'
    const status = result.error?.status
    passwordMsg.text = status === 401
      ? 'Текущий пароль неверен'
      : 'Ошибка изменения пароля. Попробуйте позже.'
  }

  setTimeout(() => { passwordMsg.text = '' }, 4000)
}
</script>

<style scoped>
.profile-page {
  min-height: 100vh;
  background: var(--bg-base);
}

/* Навбар */
.navbar {
  position: sticky;
  top: 0;
  z-index: 100;
  border-bottom: 1px solid var(--border);
  border-radius: 0;
  border-left: none;
  border-right: none;
  border-top: none;
}
.navbar-inner {
  max-width: 800px;
  margin: 0 auto;
  padding: 12px 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.nav-sep {
  width: 1px;
  height: 20px;
  background: var(--border);
}

/* Основной контент */
.profile-main {
  max-width: 680px;
  margin: 0 auto;
  padding: 40px 24px 80px;
  display: flex;
  flex-direction: column;
  gap: 24px;
}

/* Шапка профиля */
.profile-header {
  display: flex;
  align-items: center;
  gap: 20px;
  padding-bottom: 8px;
}
.profile-avatar {
  width: 64px;
  height: 64px;
  border-radius: 50%;
  background: linear-gradient(135deg, var(--primary), var(--violet));
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 28px;
  font-weight: 700;
  flex-shrink: 0;
  box-shadow: 0 8px 25px var(--primary-glow);
}
.profile-name {
  font-size: 22px;
  font-weight: 700;
  letter-spacing: -0.02em;
}

/* Карточка секции */
.profile-card {
  border-radius: var(--radius-xl);
  padding: 28px;
  display: flex;
  flex-direction: column;
  gap: 20px;
}
.card-header { border-bottom: 1px solid var(--border); padding-bottom: 16px; }
.card-title {
  font-size: 16px;
  font-weight: 700;
  margin-bottom: 4px;
}
.card-desc { font-size: 13px; color: var(--text-secondary); }
.card-form { display: flex; flex-direction: column; gap: 16px; }
.form-actions { display: flex; justify-content: flex-end; }
.field-error { font-size: 12px; color: #f87171; margin-top: 2px; }

/* Безопасность */
.security-info { display: flex; flex-direction: column; gap: 14px; }
.security-item {
  display: flex;
  align-items: flex-start;
  gap: 14px;
  padding: 12px 14px;
  background: rgba(255,255,255,0.02);
  border-radius: var(--radius-md);
  border: 1px solid var(--border);
}
.security-icon { font-size: 20px; flex-shrink: 0; margin-top: 1px; }
</style>
