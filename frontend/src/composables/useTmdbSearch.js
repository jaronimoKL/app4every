import { ref } from 'vue'

const TMDB_API_KEY = import.meta.env.VITE_TMDB_API_KEY
const TMDB_BASE_URL = 'https://api.themoviedb.org/3'

export function useTmdbSearch() {
  const tmdbResults = ref([])
  const isSearching = ref(false)

  async function searchTmdb(query) {
    if (!query || query.length < 2) {
      tmdbResults.value = []
      return
    }

    if (!TMDB_API_KEY) {
      console.warn("VITE_TMDB_API_KEY is not defined in .env")
      alert("Ошибка: Не задан ключ VITE_TMDB_API_KEY")
      return
    }

    isSearching.value = true
    try {
      const url = `${TMDB_BASE_URL}/search/multi?api_key=${TMDB_API_KEY}&language=ru-RU&query=${encodeURIComponent(query)}`
      console.log('TMDB Request URL:', url.replace(TMDB_API_KEY, 'HIDDEN_KEY'))
      
      const res = await fetch(url)
      
      if (!res.ok) {
        const errorText = await res.text()
        console.error('TMDB Error Response:', res.status, errorText)
        alert(`Ошибка TMDB: ${res.status}\nОтвет: ${errorText.substring(0, 100)}`)
        throw new Error(`TMDB error: ${res.status} - ${errorText}`)
      }
      
      const data = await res.json()
      console.log('TMDB Response Data:', data)
      
      // Фильтруем только фильмы и сериалы
      const filtered = (data.results || []).filter(r => r.media_type === 'movie' || r.media_type === 'tv')
      
      tmdbResults.value = filtered.map(item => ({
        id: item.id,
        title: item.title || item.name,
        original_title: item.original_title || item.original_name,
        media_type: item.media_type,
        poster_url: item.poster_path ? `https://image.tmdb.org/t/p/w500${item.poster_path}` : null,
        release_date: item.release_date || item.first_air_date,
        overview: item.overview,
        vote_average: item.vote_average
      }))
    } catch (err) {
      console.error('Failed to search TMDB:', err)
      alert(`Сетевая ошибка при запросе к TMDB:\n${err.message}`)
      tmdbResults.value = []
    } finally {
      isSearching.value = false
    }
  }

  async function fetchTmdbDetails(id, mediaType) {
    if (!TMDB_API_KEY) return null
    try {
      const url = `${TMDB_BASE_URL}/${mediaType}/${id}?api_key=${TMDB_API_KEY}&language=ru-RU`
      console.log('TMDB Details Request URL:', url.replace(TMDB_API_KEY, 'HIDDEN_KEY'))
      const res = await fetch(url)
      
      if (!res.ok) {
        const errorText = await res.text()
        console.error('TMDB Details Error:', res.status, errorText)
        alert(`Ошибка TMDB Details: ${res.status}\nОтвет: ${errorText.substring(0, 100)}`)
        throw new Error(`TMDB error: ${res.status}`)
      }
      return await res.json()
    } catch (err) {
      console.error('Failed to fetch TMDB details:', err)
      alert(`Сетевая ошибка при запросе деталей TMDB:\n${err.message}`)
      return null
    }
  }

  return {
    tmdbResults,
    isSearching,
    searchTmdb,
    fetchTmdbDetails
  }
}
