import re

with open('frontend/src/views/reviews/ReviewsView.vue', 'r') as f:
    content = f.read()

# Add to form
content = re.sub(
    r'shikimori_score: null,',
    r'shikimori_score: null,\n  current_episode: 0,',
    content
)

# Add to resetForm
content = re.sub(
    r'form\.shikimori_score = null',
    r'form.shikimori_score = null\n  form.current_episode = 0',
    content
)

# Add to openEdit
content = re.sub(
    r'form\.shikimori_score = rev\.shikimori_score',
    r'form.shikimori_score = rev.shikimori_score\n  form.current_episode = rev.current_episode || 0',
    content
)

with open('frontend/src/views/reviews/ReviewsView.vue', 'w') as f:
    f.write(content)

