<template>
  <div class="notes-page">

    <!-- ── Навбар ── -->
    <nav class="notes-nav glass">
      <div class="notes-nav-inner">
        <div class="flex items-center gap-3">
          <RouterLink to="/dashboard" class="btn btn-ghost" style="padding:7px 12px;font-size:13px;">
            ← Назад
          </RouterLink>
          <div class="nav-sep"></div>
          <div class="logo-mark" style="width:32px;height:32px;font-size:14px;">📝</div>
          <span style="font-weight:700;font-size:15px;">Заметки</span>
        </div>
        <div class="flex items-center gap-2">
          <span v-if="notesStore.saving" class="saving-hint">Сохраняется...</span>
          <span v-else-if="savedMsg" class="saved-hint">✓ Сохранено</span>
        </div>
      </div>
    </nav>

    <!-- ── Тело страницы ── -->
    <div class="notes-body">

      <!-- ─── Левая панель: список заметок ─── -->
      <aside class="notes-sidebar" :class="{ 'sidebar-hidden': mobileShowEditor }">
        <div class="sidebar-header">
          <div class="search-wrap">
            <span class="search-icon">🔍</span>
            <input
              v-model="searchQuery"
              type="text"
              placeholder="Поиск..."
              class="search-input"
            />
          </div>
          <button @click="handleCreate" class="create-btn" title="Новая заметка">＋</button>
        </div>

        <div class="note-list">
          <div v-if="notesStore.loading" class="list-empty">
            <span class="spinner" style="width:20px;height:20px;"></span>
          </div>
          <div v-else-if="filteredNotes.length === 0" class="list-empty">
            <div style="font-size:32px;margin-bottom:8px;">📄</div>
            <div style="font-size:13px;color:var(--text-muted);">
              {{ searchQuery ? 'Ничего не найдено' : 'Нет заметок. Создайте первую!' }}
            </div>
          </div>
          <div
            v-for="note in filteredNotes"
            :key="note.id"
            class="note-card"
            :class="{ active: notesStore.active?.id === note.id }"
            @click="selectNote(note)"
          >
            <div class="note-card-title">{{ note.title || 'Без названия' }}</div>
            <div class="note-card-preview">{{ extractText(note.content) }}</div>
            <div class="note-card-date">{{ formatDate(note.updated_at) }}</div>
          </div>
        </div>
      </aside>

      <!-- ─── Правая панель: редактор ─── -->
      <main class="notes-editor" :class="{ 'editor-visible': mobileShowEditor }">
        <!-- Пустое состояние -->
        <div v-if="!notesStore.active" class="editor-empty">
          <div style="font-size:48px;margin-bottom:16px;">✍️</div>
          <p style="color:var(--text-secondary);">Выберите заметку или создайте новую</p>
          <button @click="handleCreate" class="btn btn-primary" style="margin-top:20px;padding:10px 24px;">
            Создать заметку
          </button>
        </div>

        <!-- Редактор заметки -->
        <div v-else class="editor-content">
          <!-- Мобильная кнопка назад -->
          <button
            class="btn btn-ghost mobile-back"
            @click="mobileShowEditor = false"
          >← К списку</button>

          <!-- Заголовок -->
          <input
            v-model="editorTitle"
            type="text"
            class="editor-title-input"
            placeholder="Заголовок заметки..."
            @input="debouncedSave"
          />

          <!-- Rich Text Editor -->
          <RichTextEditor
            v-model="editorContent"
            @input="debouncedSave"
          />

          <!-- Нижняя панель -->
          <div class="editor-footer">
            <span style="font-size:12px;color:var(--text-muted);">
              Изменено: {{ formatDate(notesStore.active.updated_at) }}
            </span>
            <button
              class="btn btn-ghost"
              style="padding:6px 12px;font-size:12px;color:#f87171;"
              @click="handleDelete"
            >🗑 Удалить</button>
          </div>
        </div>
      </main>
    </div>

    <!-- Диалог подтверждения удаления -->
    <div v-if="showDeleteConfirm" class="confirm-overlay" @click.self="showDeleteConfirm = false">
      <div class="confirm-dialog glass">
        <div style="font-size:32px;margin-bottom:12px;">🗑</div>
        <h3 style="font-weight:700;margin-bottom:8px;">Удалить заметку?</h3>
        <p style="font-size:13px;color:var(--text-secondary);margin-bottom:20px;">
          «{{ notesStore.active?.title || 'Без названия' }}» будет удалена без возможности восстановления.
        </p>
        <div class="flex items-center gap-3" style="justify-content:flex-end;">
          <button class="btn btn-ghost" @click="showDeleteConfirm = false">Отмена</button>
          <button class="btn btn-primary" style="background:#ef4444;" @click="confirmDelete">Удалить</button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted } from 'vue'
import { useNotesStore } from '@/stores/notes'
import RichTextEditor from '@/components/RichTextEditor.vue'

const notesStore = useNotesStore()

// ── Состояние ──
const searchQuery       = ref('')
const editorTitle       = ref('')
const editorContent     = ref('')
const savedMsg          = ref(false)
const showDeleteConfirm = ref(false)
const mobileShowEditor  = ref(false)
let saveTimer = null

onMounted(async () => {
  await notesStore.fetchNotes()
})

// Когда меняется активная заметка — обновляем поля редактора
watch(() => notesStore.active, (note) => {
  if (note) {
    editorTitle.value   = note.title
    editorContent.value = note.content
  } else {
    editorTitle.value   = ''
    editorContent.value = ''
  }
}, { immediate: true })

// ── Фильтрация ──
const filteredNotes = computed(() => {
  const q = searchQuery.value.toLowerCase()
  if (!q) return notesStore.notes
  return notesStore.notes.filter(n =>
    n.title.toLowerCase().includes(q) ||
    extractText(n.content).toLowerCase().includes(q)
  )
})

// ── Действия ──
async function handleCreate() {
  await notesStore.createNote()
  mobileShowEditor.value = true
}

function selectNote(note) {
  notesStore.selectNote(note)
  mobileShowEditor.value = true
}

// Автосохранение с debounce 800мс
function debouncedSave() {
  clearTimeout(saveTimer)
  saveTimer = setTimeout(async () => {
    if (!notesStore.active) return
    await notesStore.saveNote(
      notesStore.active.id,
      editorTitle.value,
      editorContent.value,
    )
    savedMsg.value = true
    setTimeout(() => { savedMsg.value = false }, 2000)
  }, 800)
}

function handleDelete() {
  showDeleteConfirm.value = true
}

async function confirmDelete() {
  showDeleteConfirm.value = false
  mobileShowEditor.value  = false
  await notesStore.deleteNote(notesStore.active.id)
}

// ── Утилиты ──

/**
 * Извлекает plain text из Tiptap JSON или возвращает строку как есть.
 * Используется для превью в боковой панели.
 */
function extractText(content) {
  if (!content) return ''
  try {
    const doc = JSON.parse(content)
    if (!doc || doc.type !== 'doc') return content
    const parts = []
    function walk(node) {
      if (node.type === 'text') {
        parts.push(node.text || '')
      }
      if (node.type === 'taskItem') {
        parts.push(node.attrs?.checked ? '✓ ' : '○ ')
      }
      if (node.content) node.content.forEach(walk)
    }
    walk(doc)
    const text = parts.join('').replace(/\s+/g, ' ').trim()
    return text.length > 80 ? text.slice(0, 80) + '…' : text
  } catch {
    // Старые заметки в plain text
    const clean = content.replace(/\n/g, ' ').trim()
    return clean.length > 80 ? clean.slice(0, 80) + '…' : clean
  }
}

function formatDate(dateStr) {
  if (!dateStr) return ''
  const d    = new Date(dateStr)
  const now  = new Date()
  const diff = now - d
  if (diff < 60000)    return 'только что'
  if (diff < 3600000)  return `${Math.floor(diff / 60000)} мин. назад`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)} ч. назад`
  return d.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' })
}
</script>

<style scoped>
.notes-page {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--bg-base);
  overflow: hidden;
}

/* ── Навбар ── */
.notes-nav {
  flex-shrink: 0;
  border-radius: 0;
  border: none;
  border-bottom: 1px solid var(--border);
  z-index: 50;
}
.notes-nav-inner {
  padding: 11px 20px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.nav-sep   { width: 1px; height: 20px; background: var(--border); }
.saving-hint { font-size: 12px; color: var(--text-muted); }
.saved-hint  { font-size: 12px; color: #22c55e; }

/* ── Тело ── */
.notes-body {
  flex: 1;
  display: flex;
  overflow: hidden;
  min-height: 0;
}

/* ── Сайдбар ── */
.notes-sidebar {
  width: 260px;
  flex-shrink: 0;
  border-right: 1px solid var(--border);
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: rgba(255,255,255,0.02);
}
.sidebar-header {
  padding: 10px;
  border-bottom: 1px solid var(--border);
  display: flex;
  gap: 8px;
  align-items: center;
  flex-shrink: 0;
}
.search-wrap {
  flex: 1;
  position: relative;
  display: flex;
  align-items: center;
}
.search-icon {
  position: absolute;
  left: 9px;
  font-size: 12px;
  opacity: 0.5;
}
.search-input {
  width: 100%;
  background: rgba(255,255,255,0.04);
  border: 1px solid var(--border);
  border-radius: var(--radius-md);
  color: var(--text-primary);
  font-size: 13px;
  padding: 7px 10px 7px 30px;
  outline: none;
  transition: border-color 0.2s;
}
.search-input:focus { border-color: var(--primary); }
.search-input::placeholder { color: var(--text-muted); }
.create-btn {
  width: 32px; height: 32px;
  border-radius: var(--radius-md);
  background: var(--primary);
  border: none; color: white;
  font-size: 20px; line-height: 1;
  cursor: pointer;
  display: flex; align-items: center; justify-content: center;
  transition: opacity 0.2s, transform 0.15s;
  flex-shrink: 0;
}
.create-btn:hover { opacity: 0.85; transform: scale(1.05); }

.note-list {
  flex: 1;
  overflow-y: auto;
  padding: 8px;
  display: flex;
  flex-direction: column;
  gap: 3px;
}
.note-list::-webkit-scrollbar { width: 4px; }
.note-list::-webkit-scrollbar-thumb { background: var(--border); border-radius: 2px; }
.list-empty {
  display: flex; flex-direction: column;
  align-items: center; justify-content: center;
  padding: 40px 16px; text-align: center;
}
.note-card {
  padding: 9px 11px;
  border-radius: var(--radius-md);
  cursor: pointer;
  border: 1px solid transparent;
  transition: background 0.15s, border-color 0.15s;
}
.note-card:hover { background: rgba(255,255,255,0.04); border-color: var(--border); }
.note-card.active { background: rgba(99,102,241,0.12); border-color: rgba(99,102,241,0.3); }
.note-card-title {
  font-size: 13px; font-weight: 600; color: var(--text-primary);
  white-space: nowrap; overflow: hidden; text-overflow: ellipsis;
  margin-bottom: 3px;
}
.note-card-preview {
  font-size: 12px; color: var(--text-muted);
  overflow: hidden;
  display: -webkit-box;
  -webkit-line-clamp: 2;
  -webkit-box-orient: vertical;
  line-height: 1.5;
}
.note-card-date { font-size: 11px; color: var(--text-muted); margin-top: 5px; opacity: 0.65; }

/* ── Редактор ── */
.notes-editor {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-height: 0;
}
.editor-empty {
  flex: 1; display: flex; flex-direction: column;
  align-items: center; justify-content: center; color: var(--text-muted);
}
.editor-content {
  flex: 1; display: flex; flex-direction: column;
  overflow: hidden; padding: 20px 28px 12px; gap: 0;
}
.editor-title-input {
  font-size: 24px; font-weight: 700; letter-spacing: -0.02em;
  color: var(--text-primary); background: transparent;
  border: none; outline: none; width: 100%;
  padding: 0 0 10px; border-bottom: 1px solid var(--border);
  margin-bottom: 0; flex-shrink: 0;
}
.editor-title-input::placeholder { color: var(--text-muted); }
.editor-footer {
  display: flex; align-items: center; justify-content: space-between;
  padding-top: 10px; border-top: 1px solid var(--border);
  margin-top: 8px; flex-shrink: 0;
}

/* ── Диалог ── */
.confirm-overlay {
  position: fixed; inset: 0;
  background: rgba(0,0,0,0.5); backdrop-filter: blur(4px);
  display: flex; align-items: center; justify-content: center;
  z-index: 1000; padding: 24px;
}
.confirm-dialog {
  border-radius: var(--radius-xl); padding: 28px 32px;
  max-width: 380px; width: 100%; text-align: center;
}

/* ── Мобильный ── */
.mobile-back { display: none; }
@media (max-width: 640px) {
  .notes-sidebar { width: 100%; }
  .notes-editor  { display: none; position: absolute; inset: 0; top: 50px; z-index: 10; }
  .sidebar-hidden { display: none; }
  .editor-visible { display: flex !important; background: var(--bg-base); }
  .mobile-back { display: block; margin-bottom: 10px; font-size: 13px; padding: 7px 12px; }
  .editor-content { padding: 12px; }
}
</style>
