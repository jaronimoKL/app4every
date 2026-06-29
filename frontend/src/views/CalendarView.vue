<script setup>
import { ref, computed, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'

const auth = useAuthStore()

const calendarItems = ref([])
const loading = ref(true)
const errorMsg = ref('')
const addingState = ref({}) // shikimori_id -> boolean (is adding)

onMounted(async () => {
  await fetchCalendar()
})

async function fetchCalendar() {
  loading.value = true
  try {
    const res = await fetch('https://shikimori.one/api/calendar')
    if (!res.ok) throw new Error('Failed to fetch calendar')
    const data = await res.json()
    calendarItems.value = data
  } catch (err) {
    console.error(err)
    errorMsg.value = 'Ошибка при загрузке расписания онгоингов.'
  } finally {
    loading.value = false
  }
}

// Group items by day of week
const groupedItems = computed(() => {
  const groups = {}
  // Sort by date first
  const sorted = [...calendarItems.value].sort((a, b) => new Date(a.next_episode_at) - new Date(b.next_episode_at))

  sorted.forEach(item => {
    const date = new Date(item.next_episode_at)
    const day = date.toLocaleDateString('ru-RU', { weekday: 'long', month: 'long', day: 'numeric' })
    if (!groups[day]) {
      groups[day] = []
    }
    groups[day].push(item)
  })
  return groups
})

function formatTime(dateString) {
  const date = new Date(dateString)
  return date.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' })
}

async function addToPlanned(item) {
  const id = item.anime.id
  addingState.value[id] = true

  const requestData = {
    title: item.anime.russian || item.anime.name,
    content_type: 'anime',
    status: 'planned',
    poster_url: 'https://shikimori.one' + item.anime.image.original,
    shikimori_id: id,
    description: item.anime.description || '',
    episodes_total: item.anime.episodes || 0,
    shikimori_score: parseFloat(item.anime.score) || 0
  }

  try {
    const res = await fetch('/api/v1/reviews', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${auth.accessToken}`
      },
      body: JSON.stringify(requestData)
    })
    
    if (!res.ok) throw new Error('Не удалось добавить в локальный список')
    
    // Если Shikimori привязан, добавляем и туда
    if (auth.user?.shikimori_user_id) {
      await addToShikimori(id)
    }

    alert('Успешно добавлено в "Запланированное"!')
  } catch (err) {
    console.error(err)
    alert(err.message)
  } finally {
    addingState.value[id] = false
  }
}

async function addToShikimori(animeId) {
  const payload = {
    user_rate: {
      target_id: animeId,
      target_type: 'Anime',
      status: 'planned'
    }
  }
  
  const res = await fetch('/api/v1/auth/shikimori/rates', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${auth.accessToken}`
    },
    body: JSON.stringify(payload)
  })
  
  if (!res.ok) {
    console.warn('Не удалось синхронизировать с Shikimori, но локально добавлено')
  }
}
</script>

<template>
  <div class="calendar-container glass">
      <h1 class="page-title">Календарь Онгоингов</h1>
      <p class="page-desc">Расписание выхода новых серий аниме с Shikimori.</p>

      <div v-if="loading" class="flex justify-center mt-8">
        <span class="spinner large"></span>
      </div>
      
      <div v-else-if="errorMsg" class="error-msg">
        {{ errorMsg }}
      </div>
      
      <div v-else class="calendar-content">
        <div v-for="(items, day) in groupedItems" :key="day" class="day-group">
          <h2 class="day-title">{{ day.charAt(0).toUpperCase() + day.slice(1) }}</h2>
          <div class="anime-grid">
            <div v-for="item in items" :key="item.anime.id" class="anime-card glass">
              <div class="poster-wrapper">
                <img :src="'https://shikimori.one' + item.anime.image.original" alt="Poster" class="anime-poster" />
                <div class="episode-badge">
                  Эп. {{ item.next_episode }} • {{ formatTime(item.next_episode_at) }}
                </div>
              </div>
              <div class="anime-info">
                <h3 class="anime-title" :title="item.anime.russian || item.anime.name">
                  {{ item.anime.russian || item.anime.name }}
                </h3>
                <div class="anime-meta">
                  <span class="score" v-if="item.anime.score > 0">★ {{ item.anime.score }}</span>
                  <span class="episodes" v-if="item.anime.episodes">
                    Всего: {{ item.anime.episodes }}
                  </span>
                </div>
                <button 
                  class="btn btn-primary add-btn"
                  @click="addToPlanned(item)"
                  :disabled="addingState[item.anime.id]"
                >
                  <span v-if="addingState[item.anime.id]" class="spinner"></span>
                  <span v-else>+ Запланировано</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
</template>

<style scoped>
.calendar-container {
  padding: 32px;
  border-radius: 16px;
  max-width: 1200px;
  margin: 0 auto;
}

.page-title {
  font-size: 2rem;
  font-weight: 700;
  margin-bottom: 8px;
  background: var(--gradient-primary);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
}

.page-desc {
  color: var(--text-secondary);
  margin-bottom: 24px;
}

.day-group {
  margin-bottom: 40px;
}

.day-title {
  font-size: 1.5rem;
  font-weight: 600;
  margin-bottom: 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  padding-bottom: 8px;
}

.anime-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 20px;
}

.anime-card {
  display: flex;
  flex-direction: column;
  border-radius: 12px;
  overflow: hidden;
  transition: transform 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

.anime-card:hover {
  transform: translateY(-4px);
}

.poster-wrapper {
  position: relative;
  aspect-ratio: 2/3;
  width: 100%;
}

.anime-poster {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.episode-badge {
  position: absolute;
  bottom: 8px;
  left: 8px;
  right: 8px;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(4px);
  padding: 4px 8px;
  border-radius: 6px;
  font-size: 0.85rem;
  font-weight: 600;
  text-align: center;
}

.anime-info {
  padding: 12px;
  display: flex;
  flex-direction: column;
  flex: 1;
}

.anime-title {
  font-size: 1rem;
  font-weight: 600;
  margin-bottom: 8px;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  overflow: hidden;
}

.anime-meta {
  display: flex;
  justify-content: space-between;
  font-size: 0.85rem;
  color: var(--text-secondary);
  margin-bottom: 12px;
  margin-top: auto;
}

.score {
  color: #fbbf24;
  font-weight: 600;
}

.add-btn {
  width: 100%;
  font-size: 0.9rem;
  padding: 8px;
}

.error-msg {
  color: var(--error);
  background: rgba(239, 68, 68, 0.1);
  padding: 16px;
  border-radius: 8px;
  text-align: center;
}
</style>
