<template>
  <div class="anime-search-step">
    <div class="mb-4">
      <h3 class="text-lg font-bold mb-2">Шаг 1: Найди фильм/сериал</h3>
      <p class="text-sm text-secondary">
        Найдите тайтл в базе TMDB, чтобы автоматически заполнить постер, жанры, рейтинги и количество серий.
      </p>
    </div>

    <div class="search-box glass mb-4 p-2 rounded-lg flex items-center gap-2">
      <input 
        v-model="searchQuery"
        type="text"
        placeholder="Название (например: Интерстеллар)..."
        class="bg-transparent border-none outline-none flex-1 text-sm px-2 text-white"
        @keyup.enter="handleSearch"
      />
      <button @click="handleSearch" class="btn btn-primary px-4 py-1 text-sm" :disabled="isSearching">
        {{ isSearching ? 'Поиск...' : 'Искать' }}
      </button>
    </div>

    <div v-if="tmdbResults.length > 0" class="results-grid">
      <div 
        v-for="item in tmdbResults" 
        :key="item.id"
        class="anime-card glass glass-hover cursor-pointer flex gap-3 p-3 rounded-lg"
        @click="selectItem(item)"
      >
        <img v-if="item.poster_url" :src="item.poster_url" alt="poster" class="w-16 h-20 object-cover rounded-md flex-shrink-0" />
        <div v-else class="w-16 h-20 bg-gray-800 rounded-md flex-shrink-0 flex items-center justify-center text-xs text-gray-500">Нет фото</div>
        
        <div class="flex-1 min-w-0">
          <div class="font-bold text-sm truncate">{{ item.title }}</div>
          <div class="text-xs text-secondary truncate">{{ item.original_title }}</div>
          
          <div class="text-xs mt-2 flex flex-wrap gap-1">
            <span class="badge bg-violet-500/20 text-violet-300 px-1 rounded">{{ item.media_type === 'movie' ? 'ФИЛЬМ' : 'СЕРИАЛ' }}</span>
            <span class="badge bg-green-500/20 text-green-300 px-1 rounded">★ {{ item.vote_average ? item.vote_average.toFixed(1) : '-' }}</span>
            <span class="badge bg-gray-500/20 text-gray-300 px-1 rounded" v-if="item.release_date">{{ item.release_date.split('-')[0] }}</span>
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
import { useTmdbSearch } from '@/composables/useTmdbSearch'

const emit = defineEmits(['select', 'skip'])
const searchQuery = ref('')
const hasSearched = ref(false)

const { tmdbResults, isSearching, searchTmdb, fetchTmdbDetails } = useTmdbSearch()

async function handleSearch() {
  if (!searchQuery.value.trim()) return
  hasSearched.value = true
  await searchTmdb(searchQuery.value)
}

async function selectItem(item) {
  // Запрашиваем детали для жанров и кол-ва серий
  const details = await fetchTmdbDetails(item.id, item.media_type)
  
  emit('select', {
    ...item,
    details
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
