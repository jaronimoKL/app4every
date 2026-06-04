<template>
  <div class="flex flex-col items-center justify-center min-h-screen p-6 text-center">
    <div class="max-w-md p-8 bg-slate-800 rounded-2xl shadow-xl border border-slate-700">
      <h1 class="text-3xl font-bold bg-gradient-to-r from-blue-400 to-indigo-500 bg-clip-text text-transparent mb-4">
        App4Every Ecosystem
      </h1>
      <p class="text-slate-300 mb-6">
        Добро пожаловать в вашу приватную экосистему. Фронтенд успешно запущен и работает в связке с Caddy!
      </p>
      <div class="p-4 bg-slate-900 rounded-lg text-sm text-slate-400 font-mono">
        Статус API: <span :class="apiStatus === 'ok' ? 'text-green-400' : 'text-red-400'">{{ apiStatusText }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'

const apiStatus = ref('loading')
const apiStatusText = ref('Загрузка...')

onMounted(async () => {
  try {
    const res = await fetch('/api/v1/health')
    const data = await res.json()
    if (data.status === 'ok') {
      apiStatus.value = 'ok'
      apiStatusText.value = 'Подключено к бэкенду Go'
    } else {
      apiStatus.value = 'error'
      apiStatusText.value = 'Ошибка бэкенда'
    }
  } catch (err) {
    apiStatus.value = 'error'
    apiStatusText.value = 'Бэкенд недоступен'
  }
})
</script>

<style>
@import 'tailwindcss/base';
@import 'tailwindcss/components';
@import 'tailwindcss/utilities';
</style>
