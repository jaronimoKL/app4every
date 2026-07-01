import re

with open('frontend/src/views/watchparty/WatchPartyRoom.vue', 'r') as f:
    content = f.read()

script_addition = '''
// Import stores for progress sync
import { useReviewsStore } from '@/stores/reviews'
import { useAuthStore } from '@/stores/auth'
const reviewsStore = useReviewsStore()
const authStore = useAuthStore()

async function syncProgressWithList(episodeNum) {
  if (!shikimoriDetails.value || !authStore.isAuthenticated) return
  const shikiId = shikimoriDetails.value.id
  
  // Try to find in personal reviews
  if (reviewsStore.reviews.length === 0) {
    await reviewsStore.fetchReviews()
  }
  const review = reviewsStore.reviews.find(r => r.shikimori_id === shikiId)
  if (review && review.current_episode < episodeNum) {
    console.log(`Syncing progress to ${episodeNum} for review ${review.id}`)
    const payload = {
      title: review.title,
      content_type: review.content_type,
      status: review.status,
      rating: review.rating,
      notes: review.notes,
      poster_url: review.poster_url,
      shikimori_id: review.shikimori_id,
      description: review.description,
      episodes_total: review.episodes_total,
      current_episode: episodeNum,
      aniliberty_alias: review.aniliberty_alias,
      shikimori_score: review.shikimori_score
    }
    await reviewsStore.updateReview(review.id, payload)
  }
}
'''

content = content.replace('import { useRoute } from \'vue-router\'', 'import { useRoute } from \'vue-router\'\n' + script_addition)

# Update onEpisodeSelect to call syncProgressWithList
content = re.sub(
    r'function onEpisodeSelect\(ep\) \{\n  console\.log\("onEpisodeSelect", ep\)',
    r'function onEpisodeSelect(ep) {\n  console.log("onEpisodeSelect", ep)\n  syncProgressWithList(ep)',
    content
)

with open('frontend/src/views/watchparty/WatchPartyRoom.vue', 'w') as f:
    f.write(content)

