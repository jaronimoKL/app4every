import re

with open('frontend/src/views/reviews/ReviewsView.vue', 'r') as f:
    content = f.read()

# 1. Add isSaving ref
content = re.sub(
    r'const showModal = ref\(false\)',
    r'const showModal = ref(false)\nconst isSaving = ref(false)',
    content
)

# 2. Update handleSave function
content = re.sub(
    r'async function handleSave\(\) \{\n  if \(\!form\.title\) return',
    r'async function handleSave() {\n  if (!form.title || isSaving.value) return\n  isSaving.value = true',
    content
)

content = re.sub(
    r'closeModal\(\)\n  \} catch \(err\) \{\n    console\.error\(err\)\n  \}',
    r'closeModal()\n  } catch (err) {\n    console.error(err)\n  } finally {\n    isSaving.value = false\n  }',
    content
)

# 3. Update button
content = re.sub(
    r'<button class="btn btn-primary" :disabled="\!form\.title" @click="handleSave">Сохранить</button>',
    r'<button class="btn btn-primary" :disabled="!form.title || isSaving" @click="handleSave">{{ isSaving ? \'Сохранение...\' : \'Сохранить\' }}</button>',
    content
)

with open('frontend/src/views/reviews/ReviewsView.vue', 'w') as f:
    f.write(content)

