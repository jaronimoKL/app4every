import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useSidebarStore = defineStore('sidebar', () => {
  const isGlobalExpanded = ref(localStorage.getItem('sidebar-expanded') !== 'false')

  const toggleGlobal = () => {
    isGlobalExpanded.value = !isGlobalExpanded.value
    localStorage.setItem('sidebar-expanded', isGlobalExpanded.value)
  }

  const collapseGlobal = () => {
    isGlobalExpanded.value = false
  }

  const expandGlobal = () => {
    isGlobalExpanded.value = true
  }

  return {
    isGlobalExpanded,
    toggleGlobal,
    collapseGlobal,
    expandGlobal
  }
})
