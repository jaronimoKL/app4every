import re

with open('frontend/src/views/reviews/ReviewsView.vue', 'r') as f:
    content = f.read()

# 1. Update filter pills
content = re.sub(
    r'<div class="filter-types flex items-center gap-4">.*?</div>',
    r'''<div class="filter-types modern-pills">
          <button class="pill-btn" :class="{active: selectedContentType === 'all'}" @click="selectedContentType = 'all'">Все</button>
          <button v-for="t in contentTypes" :key="t.value" class="pill-btn" :class="{active: selectedContentType === t.value}" @click="selectedContentType = t.value">
            {{ t.icon }} {{ t.label }}
          </button>
        </div>''',
    content,
    flags=re.DOTALL
)

# 2. Update media card info
content = re.sub(
    r'<!-- Оценка -->\s*<div class="card-rating" v-if="rev.rating">\s*<span class="rating-star">★</span>\s*<span class="rating-num">{{ rev.rating }}</span>\s*<span class="rating-max">/10</span>\s*</div>',
    r'''<!-- Оценка и Прогресс -->
            <div class="card-badges-top-right" v-if="rev.rating || rev.episodes_total">
              <div class="card-rating" v-if="rev.rating">
                <span class="rating-star">★</span>
                <span class="rating-num">{{ rev.rating }}</span>
                <span class="rating-max">/10</span>
              </div>
            </div>''',
    content,
    flags=re.DOTALL
)

content = re.sub(
    r'<div class="card-title">{{ rev\.title }}</div>',
    r'''<div class="card-title">{{ rev.title }}</div>
            <div class="card-episode-progress mt-1.5 flex items-center gap-1.5 text-xs text-indigo-300 font-semibold" v-if="rev.episodes_total && rev.content_type !== \'movie\'">
              <span>🎬</span> Серия: {{ rev.current_episode || 0 }} из {{ rev.episodes_total }}
              <button 
                class="btn btn-ghost !p-1 ml-auto hover:text-white"
                v-if="(rev.current_episode || 0) < rev.episodes_total" 
                @click.stop="incrementEpisode(rev)"
                title="Отметить следующую серию просмотренной"
              >
                <span class="text-lg leading-none">+1</span>
              </button>
            </div>''',
    content
)

# 3. Update incrementEpisode script
script_addition = r'''
async function incrementEpisode(rev) {
  if (rev.current_episode >= rev.episodes_total) return
  const newEpisode = (rev.current_episode || 0) + 1
  
  const payload = {
    title: rev.title,
    content_type: rev.content_type,
    status: newEpisode === rev.episodes_total ? 'completed' : rev.status,
    rating: rev.rating,
    notes: rev.notes,
    poster_url: rev.poster_url,
    shikimori_id: rev.shikimori_id,
    description: rev.description,
    episodes_total: rev.episodes_total,
    current_episode: newEpisode,
    aniliberty_alias: rev.aniliberty_alias,
    shikimori_score: rev.shikimori_score
  }
  
  await store.updateReview(rev.id, payload)
}
'''
content = content.replace('function openCreate() {', script_addition + '\nfunction openCreate() {\n  showAnimeSearch.value = true\n')
content = content.replace('showAnimeSearch.value = true\n  isEditing.value    = false', 'isEditing.value    = false')

# 4. Form inputs for current_episode
content = re.sub(
    r'<div class="form-group">\s*<label class="form-label">Заметки</label>',
    r'''<div class="form-group" v-if="form.content_type !== 'movie'">
              <label class="form-label">Просмотрено серий</label>
              <div class="flex items-center gap-2">
                <input v-model.number="form.current_episode" type="number" min="0" class="form-input" placeholder="0" style="width:100px" />
                <span v-if="form.episodes_total">из {{ form.episodes_total }}</span>
              </div>
            </div>
            
            <div class="form-group">
              <label class="form-label">Заметки</label>''',
    content
)

with open('frontend/src/views/reviews/ReviewsView.vue', 'w') as f:
    f.write(content)

