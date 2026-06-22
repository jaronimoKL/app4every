<template>
  <div class="anime-search-step">
    <div class="mb-4">
      <h3 class="text-lg font-bold mb-2">Шаг 1: Найди аниме</h3>
      <p class="text-sm text-secondary">
        Найдите аниме на Shikimori, чтобы автоматически заполнить обложку, жанры и найти релиз для совместного просмотра.
      </p>
    </div>

    <div class="search-box glass mb-4 p-2 rounded-lg flex items-center gap-2">
      <input 
        v-model="searchQuery"
        type="text"
        placeholder="Название аниме (например: Наруто)..."
        class="bg-transparent border-none outline-none flex-1 text-sm px-2 text-white"
        @keyup.enter="handleSearch"
      />
      <button @click="handleSearch" class="btn btn-primary px-4 py-1 text-sm" :disabled="isSearching">
        {{ isSearching ? 'Поиск...' : 'Искать' }}
      </button>
    </div>

    <div v-if="shikimoriResults.length > 0" class="results-grid">
      <div 
        v-for="anime in shikimoriResults" 
        :key="anime.id"
        class="anime-card glass glass-hover cursor-pointer flex gap-3 p-3 rounded-lg"
        @click="selectAnime(anime)"
      >
        <img :src="anime.poster" alt="poster" class="w-16 h-20 object-cover rounded-md flex-shrink-0" />
        <div class="flex-1 min-w-0">
          <div class="font-bold text-sm truncate">{{ anime.title }}</div>
          <div class="text-xs text-secondary truncate">{{ anime.titleEn }}</div>
          
          <div class="text-xs mt-2 flex flex-wrap gap-1">
            <span class="badge bg-violet-500/20 text-violet-300 px-1 rounded">{{ anime.kind?.toUpperCase() }}</span>
            <span class="badge bg-green-500/20 text-green-300 px-1 rounded">★ {{ anime.score }}</span>
            <span v-if="anime.episodes" class="badge bg-gray-500/20 text-gray-300 px-1 rounded">{{ anime.episodes }} эп.</span>
          </div>
        </div>
      </div>
    </div>
    
    <div v-else-if="hasSearched && !isSearching" class="text-center text-secondary py-6 text-sm">
      Ничего не найдено 😔
    </div>

    <div class="mt-6 flex justify-end">
      <button @click="$emit('skip')" class="btn btn-ghost text-sm">
        Пропустить → заполнить вручную
      </button>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useAnimeSearch } from '@/composables/useAnimeSearch'

const emit = defineEmits(['select', 'skip'])
const searchQuery = ref('')
const hasSearched = ref(false)

const { shikimoriResults, isSearching, searchShikimori, fetchShikimoriDetails, autoMatchAniliberty } = useAnimeSearch()

async function handleSearch() {
  if (!searchQuery.value.trim()) return
  hasSearched.value = true
  await searchShikimori(searchQuery.value)
}

async function selectAnime(anime) {
  // 1. Fetch full details to get genres and description
  const details = await fetchShikimoriDetails(anime.id)
  
  // 2. Try to auto-match AniLiberty
  // AniLiberty search usually matches by russian name
  const anilibertyRelease = await autoMatchAniliberty(anime.title)

  emit('select', {
    ...anime,
    details, // contains genres, description
    anilibertyRelease // contains alias, episodes_total
  })
}
</script>

<style scoped>
.results-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(250px, 1fr));
  gap: 12px;
  max-height: 400px;
  overflow-y: auto;
}
.badge {
  font-size: 10px;
}
</style>
