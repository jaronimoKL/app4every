import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { useAuthStore } from '@/stores/auth'

const BASE = '/api/v1'

async function api(method, path, body = null, token = null) {
  const headers = { 'Content-Type': 'application/json' }
  if (token) headers['Authorization'] = `Bearer ${token}`
  const res = await fetch(`${BASE}${path}`, {
    method, headers,
    credentials: 'include',
    body: body ? JSON.stringify(body) : null,
  })
  if (res.status === 204) return null
  const data = await res.json()
  if (!res.ok) throw { status: res.status, ...data }
  return data
}

export const useReviewsStore = defineStore('reviews', () => {
  const reviews = ref([])
  const loading = ref(false)
  const saving  = ref(false)

  function token() { return useAuthStore().accessToken }

  // ── CRUD рецензий ──

  async function fetchReviews() {
    loading.value = true
    try {
      const localReviews = await api('GET', '/reviews', null, token())
      
      // Попытка загрузить с Shikimori
      let shikiReviews = []
      try {
        const auth = useAuthStore()
        if (auth.user?.shikimori_user_id) {
          const res = await fetch('/api/v1/auth/shikimori/rates', {
            headers: { 'Authorization': `Bearer ${token()}` }
          })
          if (res.ok) {
            const rates = await res.json()
            // Мапим rates в формат наших reviews
            shikiReviews = rates.map(rate => ({
              id: 'shiki_' + rate.id, // Временный ID
              title: rate.anime?.russian || rate.anime?.name,
              content_type: 'anime',
              status: rate.status === 'on_hold' ? 'planned' : rate.status, // Shikimori 'on_hold' -> planned (или можно оставить)
              rating: rate.score || null,
              notes: rate.text || '',
              poster_url: rate.anime?.image?.original ? 'https://shikimori.one' + rate.anime.image.original : '',
              shikimori_id: rate.anime?.id,
              episodes_total: rate.anime?.episodes || null,
              shikimori_score: parseFloat(rate.anime?.score) || null,
              isShikimoriOnly: true
            }))
            
            // Фильтруем: берем только Смотрю(watching), Просмотрено(completed), Брошено(dropped)
            // Исключаем запланированное, как просил пользователь, чтобы не засорять список
            shikiReviews = shikiReviews.filter(r => ['watching', 'completed', 'dropped'].includes(r.status))
          }
        }
      } catch (err) {
        console.warn('Не удалось загрузить списки Shikimori', err)
      }

      // Объединяем, удаляя дубликаты (если аниме есть и локально, и на shikimori, оставляем локальное)
      const merged = [...localReviews]
      for (const sr of shikiReviews) {
        if (!merged.find(mr => mr.shikimori_id === sr.shikimori_id)) {
          merged.push(sr)
        }
      }
      reviews.value = merged

    } finally {
      loading.value = false
    }
  }

  async function createReview(data) {
    saving.value = true
    try {
      const rev = await api('POST', '/reviews', data, token())
      reviews.value.unshift(rev)
      return rev
    } finally {
      saving.value = false
    }
  }

  async function updateReview(id, data) {
    saving.value = true
    try {
      const updated = await api('PUT', `/reviews/${id}`, data, token())
      const idx = reviews.value.findIndex(r => r.id === id)
      if (idx !== -1) reviews.value[idx] = updated
      return updated
    } finally {
      saving.value = false
    }
  }

  async function deleteReview(id) {
    await api('DELETE', `/reviews/${id}`, null, token())
    reviews.value = reviews.value.filter(r => r.id !== id)
  }

  // ── Ссылки ──

  async function addLink(reviewId, label, url) {
    const link = await api('POST', `/reviews/${reviewId}/links`, { label, url }, token())
    // Обновляем links в локальном состоянии
    const rev = reviews.value.find(r => r.id === reviewId)
    if (rev) rev.links = [...(rev.links || []), link]
    return link
  }

  async function deleteLink(reviewId, linkId) {
    await api('DELETE', `/reviews/${reviewId}/links/${linkId}`, null, token())
    const rev = reviews.value.find(r => r.id === reviewId)
    if (rev) rev.links = rev.links.filter(l => l.id !== linkId)
  }

  // ── Жанры ──

  async function addGenre(reviewId, name) {
    const genre = await api('POST', `/reviews/${reviewId}/genres`, { name }, token())
    const rev = reviews.value.find(r => r.id === reviewId)
    if (rev) rev.genres = [...(rev.genres || []), genre]
    return genre
  }

  async function deleteGenre(reviewId, genreId) {
    await api('DELETE', `/reviews/${reviewId}/genres/${genreId}`, null, token())
    const rev = reviews.value.find(r => r.id === reviewId)
    if (rev) rev.genres = rev.genres.filter(g => g.id !== genreId)
  }

  // ── Computed: группировка по статусам ──

  const byStatus = computed(() => ({
    watching:  reviews.value.filter(r => r.status === 'watching'),
    completed: reviews.value.filter(r => r.status === 'completed'),
    planned:   reviews.value.filter(r => r.status === 'planned'),
    dropped:   reviews.value.filter(r => r.status === 'dropped'),
  }))

  return {
    reviews, loading, saving, byStatus,
    fetchReviews, createReview, updateReview, deleteReview,
    addLink, deleteLink,
    addGenre, deleteGenre,
  }
})
