import { ref } from 'vue'
import { useAuthStore } from '@/stores/auth'

export function useAnimeSearch() {
  const shikimoriResults = ref([])
  const anilibertyResults = ref([])
  const isSearching = ref(false)

  // Поиск через наш бэкенд-прокси
  async function searchShikimori(query) {
    if (!query || query.length < 2) return
    isSearching.value = true
    const auth = useAuthStore()
    try {
      const res = await fetch(`/api/v1/reviews/integrations/shikimori/search?q=${encodeURIComponent(query)}&limit=8`, {
        headers: {
          'Authorization': `Bearer ${auth.accessToken}`
        }
      })
      const data = await res.json()
      shikimoriResults.value = data.map(item => ({
        id: item.id,
        title: item.russian || item.name,
        titleEn: item.name,
        poster: `https://shikimori.io${item.image.preview}`,
        posterFull: `https://shikimori.io${item.image.original}`,
        score: parseFloat(item.score),
        kind: item.kind,        // tv | movie | ova | ona
        episodes: item.episodes,
        status: item.status,
        airedOn: item.aired_on,
      }))
    } catch (e) {
      console.error("Failed to search Shikimori", e)
      shikimoriResults.value = []
    }
    isSearching.value = false
  }

  async function searchAniliberty(query) {
    const auth = useAuthStore()
    try {
      const res = await fetch(`/api/v1/reviews/integrations/aniliberty/search?q=${encodeURIComponent(query)}&limit=5`, {
        headers: {
          'Authorization': `Bearer ${auth.accessToken}`
        }
      })
      const data = await res.json()
      // API возвращает массив напрямую (app/search/releases)
      anilibertyResults.value = Array.isArray(data) ? data : (data.data || [])
    } catch (e) {
      console.error("Failed to search AniLiberty", e)
      anilibertyResults.value = []
    }
  }

  // Авто-поиск в AniLiberty по русскому названию из Shikimori
  async function autoMatchAniliberty(russianName) {
    await searchAniliberty(russianName)
    // Берём первый результат если есть совпадение
    return anilibertyResults.value[0] || null
  }

  async function fetchEpisodes(alias) {
    const auth = useAuthStore()
    try {
      const res = await fetch(`/api/v1/reviews/integrations/aniliberty/episodes/${alias}`, {
        headers: {
          'Authorization': `Bearer ${auth.accessToken}`
        }
      })
      const data = await res.json()
      return data.episodes || []  // массив эпизодов
    } catch (e) {
      console.error("Failed to fetch AniLiberty episodes", e)
      return []
    }
  }

  // Детальная инфа (включая жанры) с Shikimori
  async function fetchShikimoriDetails(id) {
    try {
      // Поскольку мы кэшируем только поиск на бэке, детали можно дергать напрямую или через бэк (если есть эндпоинт).
      // В задании не просили делать прокси для деталей, но CORS может мешать? 
      // Shikimori API без авторизации обычно открыт для CORS, но лучше проверить.
      // Попробуем напрямую с fetch, так как в задании прокси только для /search
      const res = await fetch(`https://shikimori.io/api/animes/${id}`)
      if (!res.ok) throw new Error("Failed")
      return await res.json()
    } catch (e) {
      console.error("Failed to fetch Shikimori details", e)
      return null
    }
  }

  return { shikimoriResults, anilibertyResults, isSearching, searchShikimori, searchAniliberty, autoMatchAniliberty, fetchEpisodes, fetchShikimoriDetails }
}
