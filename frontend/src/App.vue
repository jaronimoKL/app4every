<template>
  <div class="app-layout">
    <AppHeader v-if="authStore.isAuthenticated" />
    <RouterView v-slot="{ Component }">
      <Transition name="page" mode="out-in">
        <component :is="Component" />
      </Transition>
    </RouterView>
  </div>
</template>

<script setup>
import { onMounted } from 'vue'
import { useAuthStore } from './stores/auth'
import { useThemeStore } from './stores/theme'
import AppHeader from './components/AppHeader.vue'

const authStore = useAuthStore()
const themeStore = useThemeStore() // initialises theme

onMounted(() => {
  // Theme is applied by store creation
})
</script>

<style>
.app-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}
</style>
