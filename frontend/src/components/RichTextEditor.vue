<template>
  <div class="rte-wrap">
    <!-- ── Панель инструментов ── -->
    <div class="rte-toolbar" v-if="editor">
      <!-- Текстовые стили -->
      <div class="tb-group">
        <button
          class="tb-btn"
          :class="{ active: editor.isActive('bold') }"
          @click="editor.chain().focus().toggleBold().run()"
          title="Жирный (Ctrl+B)"
        ><strong>B</strong></button>
        <button
          class="tb-btn"
          :class="{ active: editor.isActive('italic') }"
          @click="editor.chain().focus().toggleItalic().run()"
          title="Курсив (Ctrl+I)"
        ><em>I</em></button>
        <button
          class="tb-btn"
          :class="{ active: editor.isActive('strike') }"
          @click="editor.chain().focus().toggleStrike().run()"
          title="Зачёркивание"
        ><s>S</s></button>
      </div>

      <div class="tb-sep"></div>

      <!-- Заголовки -->
      <div class="tb-group">
        <button
          class="tb-btn tb-btn--label"
          :class="{ active: editor.isActive('heading', { level: 1 }) }"
          @click="editor.chain().focus().toggleHeading({ level: 1 }).run()"
          title="Заголовок 1"
        >H1</button>
        <button
          class="tb-btn tb-btn--label"
          :class="{ active: editor.isActive('heading', { level: 2 }) }"
          @click="editor.chain().focus().toggleHeading({ level: 2 }).run()"
          title="Заголовок 2"
        >H2</button>
      </div>

      <div class="tb-sep"></div>

      <!-- Списки -->
      <div class="tb-group">
        <button
          class="tb-btn"
          :class="{ active: editor.isActive('bulletList') }"
          @click="editor.chain().focus().toggleBulletList().run()"
          title="Маркированный список"
        >
          <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
            <circle cx="2" cy="4" r="1.5"/><rect x="5" y="3" width="10" height="2" rx="1"/>
            <circle cx="2" cy="8" r="1.5"/><rect x="5" y="7" width="10" height="2" rx="1"/>
            <circle cx="2" cy="12" r="1.5"/><rect x="5" y="11" width="10" height="2" rx="1"/>
          </svg>
        </button>
        <button
          class="tb-btn"
          :class="{ active: editor.isActive('orderedList') }"
          @click="editor.chain().focus().toggleOrderedList().run()"
          title="Нумерованный список"
        >
          <svg width="14" height="14" viewBox="0 0 16 16" fill="currentColor">
            <text x="0" y="5" font-size="5" font-family="monospace">1.</text>
            <rect x="6" y="3" width="9" height="2" rx="1"/>
            <text x="0" y="9.5" font-size="5" font-family="monospace">2.</text>
            <rect x="6" y="7.5" width="9" height="2" rx="1"/>
            <text x="0" y="14" font-size="5" font-family="monospace">3.</text>
            <rect x="6" y="12" width="9" height="2" rx="1"/>
          </svg>
        </button>
        <button
          class="tb-btn"
          :class="{ active: editor.isActive('taskList') }"
          @click="editor.chain().focus().toggleTaskList().run()"
          title="Список задач / чеклист"
        >
          <svg width="14" height="14" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5">
            <rect x="1" y="2.5" width="5" height="5" rx="1"/>
            <polyline points="2.5,5 4,6.5 6.5,3.5" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
            <rect x="1" y="9.5" width="5" height="5" rx="1"/>
            <line x1="8" y1="5" x2="15" y2="5"/>
            <line x1="8" y1="12" x2="15" y2="12"/>
          </svg>
        </button>
      </div>

      <div class="tb-sep"></div>

      <!-- История -->
      <div class="tb-group">
        <button
          class="tb-btn"
          :disabled="!editor.can().undo()"
          @click="editor.chain().focus().undo().run()"
          title="Отменить (Ctrl+Z)"
        >↩</button>
        <button
          class="tb-btn"
          :disabled="!editor.can().redo()"
          @click="editor.chain().focus().redo().run()"
          title="Повторить (Ctrl+Y)"
        >↪</button>
      </div>
    </div>

    <!-- ── Редактор ── -->
    <EditorContent :editor="editor" class="rte-body" />
  </div>
</template>

<script setup>
import { watch, onBeforeUnmount } from 'vue'
import { useEditor, EditorContent } from '@tiptap/vue-3'
import StarterKit from '@tiptap/starter-kit'
import TaskList from '@tiptap/extension-task-list'
import TaskItem from '@tiptap/extension-task-item'

const props = defineProps({
  modelValue: { type: String, default: '' },
})
const emit = defineEmits(['update:modelValue', 'input'])

// Парсим контент: поддерживаем старый plain-text и новый Tiptap JSON.
function parseContent(value) {
  if (!value) return ''
  try {
    const parsed = JSON.parse(value)
    // Убеждаемся что это Tiptap-документ
    if (parsed && parsed.type === 'doc') return parsed
    return value
  } catch {
    // Старые заметки в plain-text — Tiptap обработает как HTML
    return value
  }
}

const editor = useEditor({
  content: parseContent(props.modelValue),
  extensions: [
    StarterKit.configure({
      // Headings до уровня 3
      heading: { levels: [1, 2, 3] },
    }),
    TaskList,
    TaskItem.configure({
      nested: true,
    }),
  ],
  editorProps: {
    attributes: {
      class: 'prose-editor',
      spellcheck: 'false',
    },
  },
  onUpdate({ editor }) {
    const json = JSON.stringify(editor.getJSON())
    emit('update:modelValue', json)
    emit('input')
  },
})

// Когда меняется активная заметка — загружаем новый контент в редактор
watch(() => props.modelValue, (newVal) => {
  if (!editor.value) return
  const currentJson = JSON.stringify(editor.value.getJSON())
  if (newVal !== currentJson) {
    editor.value.commands.setContent(parseContent(newVal), false)
  }
})

onBeforeUnmount(() => {
  editor.value?.destroy()
})
</script>

<style scoped>
.rte-wrap {
  display: flex;
  flex-direction: column;
  flex: 1;
  overflow: hidden;
  min-height: 0;
}

/* ── Toolbar ── */
.rte-toolbar {
  display: flex;
  align-items: center;
  gap: 2px;
  padding: 6px 8px;
  border-bottom: 1px solid var(--border);
  background: rgba(255,255,255,0.02);
  flex-wrap: wrap;
  flex-shrink: 0;
}
.tb-group {
  display: flex;
  align-items: center;
  gap: 1px;
}
.tb-sep {
  width: 1px;
  height: 18px;
  background: var(--border);
  margin: 0 4px;
  opacity: 0.6;
}
.tb-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  height: 28px;
  border-radius: 6px;
  border: none;
  background: transparent;
  color: var(--text-secondary);
  cursor: pointer;
  font-size: 13px;
  font-family: inherit;
  transition: background 0.15s, color 0.15s;
  padding: 0;
}
.tb-btn--label { font-size: 11px; font-weight: 700; letter-spacing: -0.02em; }
.tb-btn:hover { background: rgba(255,255,255,0.07); color: var(--text-primary); }
.tb-btn.active { background: rgba(99,102,241,0.2); color: #a5b4fc; }
.tb-btn:disabled { opacity: 0.25; cursor: default; }
.tb-btn:disabled:hover { background: transparent; color: var(--text-secondary); }

/* ── Editor body ── */
.rte-body {
  flex: 1;
  overflow-y: auto;
  padding: 16px 0 8px;
}
.rte-body::-webkit-scrollbar { width: 4px; }
.rte-body::-webkit-scrollbar-thumb { background: var(--border); border-radius: 2px; }
</style>

<!-- Глобальные стили ProseMirror — без scoped, иначе не применятся к .ProseMirror -->
<style>
.prose-editor {
  outline: none;
  min-height: 180px;
  line-height: 1.75;
  color: var(--text-primary);
  font-size: 15px;
  caret-color: #a5b4fc;
}

/* Placeholder */
.prose-editor p.is-empty:first-child::before {
  content: 'Начните писать...';
  color: var(--text-muted, #555);
  pointer-events: none;
  float: left;
  height: 0;
}

/* Headings */
.prose-editor h1 {
  font-size: 24px; font-weight: 700; letter-spacing: -0.025em;
  margin: 20px 0 10px; line-height: 1.3;
  color: var(--text-primary);
}
.prose-editor h2 {
  font-size: 19px; font-weight: 600; letter-spacing: -0.015em;
  margin: 16px 0 8px; line-height: 1.4;
}
.prose-editor h3 {
  font-size: 16px; font-weight: 600;
  margin: 14px 0 6px;
}

/* Paragraphs */
.prose-editor p { margin: 4px 0; }

/* Lists */
.prose-editor ul,
.prose-editor ol {
  padding-left: 22px;
  margin: 4px 0;
}
.prose-editor li { margin: 3px 0; }
.prose-editor li > p { margin: 0; }

/* ── Task list / Checklist ── */
.prose-editor ul[data-type="taskList"] {
  list-style: none;
  padding-left: 0;
}
.prose-editor ul[data-type="taskList"] li {
  display: flex;
  align-items: flex-start;
  gap: 9px;
  margin: 5px 0;
}
.prose-editor ul[data-type="taskList"] li > label {
  display: flex;
  align-items: center;
  margin-top: 3px;
  flex-shrink: 0;
  cursor: pointer;
}
.prose-editor ul[data-type="taskList"] li > label > input[type="checkbox"] {
  width: 15px;
  height: 15px;
  border-radius: 3px;
  accent-color: #6366f1;
  cursor: pointer;
}
/* Зачёркнутый текст выполненной задачи */
.prose-editor ul[data-type="taskList"] li[data-checked="true"] > div {
  text-decoration: line-through;
  opacity: 0.45;
}
/* Вложенные task-list */
.prose-editor ul[data-type="taskList"] ul[data-type="taskList"] {
  padding-left: 22px;
}

/* Bold, italic, strike */
.prose-editor strong { font-weight: 700; }
.prose-editor em     { font-style: italic; }
.prose-editor s      { text-decoration: line-through; opacity: 0.6; }

/* Inline code */
.prose-editor code {
  background: rgba(255,255,255,0.08);
  border-radius: 4px;
  padding: 2px 6px;
  font-family: 'JetBrains Mono', 'Fira Code', monospace;
  font-size: 13px;
  color: #c4b5fd;
}

/* Code block */
.prose-editor pre {
  background: rgba(0,0,0,0.3);
  border: 1px solid var(--border, rgba(255,255,255,0.08));
  border-radius: 8px;
  padding: 14px 16px;
  overflow-x: auto;
  margin: 8px 0;
}
.prose-editor pre code {
  background: none;
  padding: 0;
  font-size: 13px;
  color: var(--text-primary);
}

/* Blockquote */
.prose-editor blockquote {
  border-left: 3px solid #6366f1;
  padding-left: 16px;
  margin: 8px 0;
  color: var(--text-secondary, #999);
  font-style: italic;
}

/* Horizontal rule */
.prose-editor hr {
  border: none;
  border-top: 1px solid var(--border, rgba(255,255,255,0.08));
  margin: 14px 0;
}

/* Selected text */
.prose-editor ::selection { background: rgba(99,102,241,0.3); }
</style>
