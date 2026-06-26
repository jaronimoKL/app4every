import { defineStore } from 'pinia'
import { ref, watch } from 'vue'

export const useThemeStore = defineStore('theme', () => {
  const currentTheme = ref(localStorage.getItem('app-theme') || 'macchiato')

  const applyTheme = (theme) => {
    if (theme === 'latte') {
      document.documentElement.setAttribute('data-theme', 'latte')
    } else {
      document.documentElement.removeAttribute('data-theme')
    }
  }

  // Apply immediately on creation
  applyTheme(currentTheme.value)

  watch(currentTheme, (newTheme) => {
    localStorage.setItem('app-theme', newTheme)
    applyTheme(newTheme)
  })

  const toggleTheme = () => {
    currentTheme.value = currentTheme.value === 'macchiato' ? 'latte' : 'macchiato'
  }

  return {
    currentTheme,
    toggleTheme
  }
})
