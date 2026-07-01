import re

with open('frontend/src/views/reviews/ReviewsView.vue', 'r') as f:
    content = f.read()

content = re.sub(
    r'form\.shikimori_score = anime\.score',
    r'form.shikimori_score = anime.score\n  form.current_episode = anime.episodes || 0',
    content
)

with open('frontend/src/views/reviews/ReviewsView.vue', 'w') as f:
    f.write(content)

