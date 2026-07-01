import re

with open('frontend/src/views/reviews/GroupReviewsView.vue', 'r') as f:
    content = f.read()

# 1. Add isSaving ref
content = re.sub(
    r'const showItemModal = ref\(false\)',
    r'const showItemModal = ref(false)\nconst isSaving = ref(false)',
    content
)

# 2. Update saveItem function
content = re.sub(
    r'async function saveItem\(\) \{\n  if \(\!itemForm\.title \|\| \!activeGroup\.value\) return',
    r'async function saveItem() {\n  if (!itemForm.title || !activeGroup.value || isSaving.value) return\n  isSaving.value = true',
    content
)

content = re.sub(
    r'closeItemModal\(\)\n  \} catch \(err\) \{\n    console\.error\(err\)\n  \}',
    r'closeItemModal()\n  } catch (err) {\n    console.error(err)\n  } finally {\n    isSaving.value = false\n  }',
    content
)

# 3. Update button
content = re.sub(
    r'<button class="btn btn-primary" :disabled="\!itemForm\.title" @click="saveItem">Сохранить</button>',
    r'<button class="btn btn-primary" :disabled="!itemForm.title || isSaving" @click="saveItem">{{ isSaving ? \'Сохранение...\' : \'Сохранить\' }}</button>',
    content
)

with open('frontend/src/views/reviews/GroupReviewsView.vue', 'w') as f:
    f.write(content)

