<template>
  <div class="room-page">
    <!-- ══ ГРУППА СКРЫТЫХ АУДИОПЛЕЕРОВ ДЛЯ ГОЛОСА ══ -->
    <div style="display:none;">
      <audio
        v-for="(stream, userID) in webrtc.remoteVoiceStreams"
        :key="'voice-' + userID"
        :ref="el => { 
          if (el) { 
            el.srcObject = stream; 
            el.volume = isAllMuted ? 0 : (userVolumes[userID] !== undefined ? userVolumes[userID] / 100 : 1.0); 
            el.muted = isAllMuted; 
          } 
        }"
        autoplay
      ></audio>
    </div>

    <!-- ══ НАВБАР ══ -->
    <nav class="room-nav glass">
      <div class="room-nav-inner flex justify-between items-center px-6 py-3">
        <div class="flex items-center gap-3">
          <span style="font-size:18px;">📹</span>
          <span style="font-weight:700;font-size:15px;">Комната: <span class="text-indigo-400">{{ roomId }}</span></span>
        </div>
        <div class="flex items-center gap-3">
          <!-- Предупреждение о количестве участников -->
          <span v-if="participantCount > 4" class="badge-warning text-xs px-3 py-1 rounded-full">
            ⚠ Рекомендуется до 4 участников (Mesh P2P)
          </span>
          <span class="online-indicator text-xs text-green-400 flex items-center gap-1.5">
            <span class="ping-dot"></span> Сеть: OK
          </span>
        </div>
      </div>
    </nav>

    <!-- ══ ОСНОВНАЯ ЗОНА ══ -->
    <div class="room-main-layout flex flex-1 overflow-hidden">
      
      <!-- Зона с видео-трансляциями -->
      <div class="video-workspace flex-1 flex flex-col p-6 overflow-y-auto">
        
        <!-- Большой экран (активный стрим) -->
        <div class="primary-screen-holder flex-1 flex items-center justify-center relative mb-4 rounded-2xl overflow-hidden glass border border-[var(--border)] min-h-[360px]">
          <video 
            v-if="activeStream" 
            :ref="el => { 
              if (el) { 
                el.srcObject = activeStream; 
                el.volume = isAllMuted ? 0 : (activeStreamOwnerId && streamVolumes[activeStreamOwnerId] !== undefined ? streamVolumes[activeStreamOwnerId] / 100 : 1.0); 
                el.muted = isAllMuted || (activeStreamOwnerId && streamVolumes[activeStreamOwnerId] === 0);
              } 
            }" 
            autoplay 
            playsinline 
            class="primary-video w-full h-full object-contain"
          ></video>
          
          <div v-else class="no-stream-placeholder text-center p-8">
            <div style="font-size:64px;margin-bottom:16px;">🖥️</div>
            <h3 style="font-weight:700;font-size:18px;margin-bottom:8px;">Трансляция не запущена</h3>
            <p style="font-size:13.5px;color:var(--text-secondary);max-width:380px;margin:0 auto;">
              Нажмите кнопку «Показать экран» внизу, чтобы запустить демонстрацию, или выберите участника в панели ниже.
            </p>
          </div>

          <!-- Название активного стрима -->
          <div v-if="activeStreamOwner" class="active-stream-badge absolute top-4 left-4 glass px-3 py-1.5 rounded-lg text-xs font-semibold">
            🖥️ Демонстрация: {{ activeStreamOwner }}
          </div>
        </div>

        <!-- Сетка миниатюр / Участники -->
        <div class="participants-section">
          <h4 class="section-title">Участники сессии</h4>
          <div class="participants-grid flex flex-wrap gap-4">
            
            <!-- Моя карточка -->
            <div class="participant-card glass p-3 rounded-xl border border-[var(--border)] flex flex-col gap-2 min-w-[160px] relative">
              <div class="flex justify-between items-center">
                <span class="user-name font-semibold text-sm">Вы ({{ authStore.user?.username }})</span>
                <span class="status-dot-indicator connected"></span>
              </div>
              <div v-if="screenShare.active.value" class="text-[11px] text-indigo-300">
                Захвачено: {{ screenShare.resolution.value }}
              </div>
              <div v-else class="text-[11px] text-slate-500 italic">Экран не захвачен</div>
              
              <!-- Мое превью -->
              <div v-if="screenShare.active.value" class="mini-preview rounded-lg overflow-hidden border border-indigo-500/30 aspect-video mt-1 bg-black">
                <video 
                  :ref="el => { if (el) el.srcObject = screenShare.stream.value }" 
                  autoplay 
                  muted 
                  playsinline 
                  class="w-full h-full object-cover"
                ></video>
              </div>
            </div>

            <!-- Карточки удаленных участников -->
            <div 
              v-for="peer in remoteParticipants" 
              :key="peer.user_id" 
              class="participant-card glass p-3 rounded-xl border border-[var(--border)] flex flex-col gap-2 min-w-[190px] cursor-pointer"
              :class="{ 'border-indigo-500': activeStream === webrtc.remoteScreenStreams[peer.user_id] }"
              @click="selectActiveStream(webrtc.remoteScreenStreams[peer.user_id], peer.user_id, peer.username)"
            >
              <div class="flex justify-between items-center">
                <span class="user-name font-semibold text-sm">{{ peer.username }}</span>
                
                <!-- Индикатор статуса WebRTC (connected, failed, TURN, etc.) -->
                <span 
                  class="status-dot-indicator" 
                  :class="getRTCStatusClass(peer.user_id)"
                  :title="'Статус: ' + (webrtc.connectionStates[peer.user_id] || 'connecting')"
                ></span>
              </div>

              <!-- Тип подключения (P2P или TURN) -->
              <div class="flex justify-between text-[10px] text-slate-400">
                <span>Связь:</span>
                <span 
                  class="font-semibold" 
                  :style="{ color: getRTCStatusColor(peer.user_id) }"
                >
                  {{ webrtc.peerTypes[peer.user_id] || 'handshake' }}
                </span>
              </div>

              <!-- Превью удаленного экрана -->
              <div class="mini-preview rounded-lg overflow-hidden border border-[var(--border)] aspect-video mt-1 bg-black flex items-center justify-center relative">
                <video 
                  v-if="webrtc.remoteScreenStreams[peer.user_id]" 
                  :ref="el => { 
                    if (el) { 
                      el.srcObject = webrtc.remoteScreenStreams[peer.user_id]; 
                      el.volume = isAllMuted ? 0 : (streamVolumes[peer.user_id] !== undefined ? streamVolumes[peer.user_id] / 100 : 1.0); 
                      el.muted = isAllMuted || (streamVolumes[peer.user_id] === 0);
                    } 
                  }" 
                  autoplay 
                  playsinline 
                  class="w-full h-full object-cover"
                ></video>
                <div v-else class="text-[10px] text-slate-500 italic">Ждем стрим...</div>
              </div>

              <!-- Ползунки громкости (Voice & Stream) -->
              <div class="volume-controls flex flex-col gap-1.5 mt-2 pt-2 border-t border-[var(--border)]" @click.stop>
                <!-- Громкость голоса -->
                <div class="flex items-center justify-between gap-1 text-[10px] text-slate-300">
                  <span>🎙️ Голос:</span>
                  <input 
                    type="range" 
                    min="0" 
                    max="100" 
                    v-model.number="userVolumes[peer.user_id]" 
                    class="volume-slider w-20"
                  />
                  <span class="w-6 text-right">{{ userVolumes[peer.user_id] ?? 100 }}%</span>
                </div>
                
                <!-- Громкость стрима экрана (если есть) -->
                <div v-if="webrtc.remoteScreenStreams[peer.user_id]" class="flex items-center justify-between gap-1 text-[10px] text-indigo-300">
                  <span>🖥️ Стрим:</span>
                  <input 
                    type="range" 
                    min="0" 
                    max="100" 
                    v-model.number="streamVolumes[peer.user_id]" 
                    class="volume-slider w-20"
                  />
                  <span class="w-6 text-right">{{ streamVolumes[peer.user_id] ?? 100 }}%</span>
                </div>
              </div>

            </div>

          </div>
        </div>

      </div>

      <!-- Правая панель (Чат) -->
      <aside class="chat-sidebar w-80 border-l border-[var(--border)] flex flex-col glass">
        <div class="chat-header p-4 border-b border-[var(--border)] flex justify-between items-center">
          <h3 class="font-bold text-sm flex items-center gap-2">
            <span>💬</span> P2P Текстовый чат
          </h3>
          <span class="text-[10px] text-slate-400 bg-slate-800 px-2 py-0.5 rounded-full">без сервера</span>
        </div>

        <!-- Список сообщений -->
        <div class="chat-messages flex-1 overflow-y-auto p-4 flex flex-col gap-3">
          <div 
            v-for="(msg, idx) in webrtc.chatMessages" 
            :key="idx"
            class="chat-bubble flex flex-col max-w-[85%] rounded-2xl p-3 text-xs"
            :class="msg.own ? 'self-end bg-indigo-600 text-white rounded-tr-none' : 'self-start bg-slate-800 text-slate-100 rounded-tl-none'"
          >
            <span v-if="!msg.own" class="font-semibold text-indigo-300 mb-1" style="font-size:10px;">{{ msg.from_name }}</span>
            <p class="break-words">{{ msg.text }}</p>
            <span class="text-[9px] text-slate-400 mt-1 align-self-end text-right">
              {{ formatTime(msg.ts) }}
            </span>
          </div>
          <div v-if="webrtc.chatMessages.length === 0" class="text-center text-slate-500 italic text-xs py-8">
            Напишите сообщение, чтобы начать P2P чат.
          </div>
        </div>

        <!-- Отправка сообщений -->
        <div class="chat-input-holder p-3 border-t border-[var(--border)] flex gap-2">
          <input 
            v-model="chatInput" 
            type="text" 
            class="form-input flex-1 text-xs" 
            placeholder="Напишите сообщение..." 
            style="padding: 8px 12px;"
            @keyup.enter="sendMessage"
          />
          <button class="btn btn-primary px-3 text-xs" @click="sendMessage">✓</button>
        </div>
      </aside>

    </div>

    <!-- ══ ПАНЕЛЬ УПРАВЛЕНИЯ (НИЖНЯЯ) ══ -->
    <footer class="room-controls-bar glass border-t border-[var(--border)] py-4 px-6 flex justify-between items-center">
      <div class="flex items-center gap-3">
        <!-- Показать / Остановить экран -->
        <button 
          v-if="!screenShare.active.value" 
          class="btn btn-primary flex items-center gap-1.5"
          @click="startScreenSharing"
        >
          🖥️ Показать экран
        </button>
        <button 
          v-else 
          class="btn flex items-center gap-1.5"
          style="background:#ef4444;color:white;border:none;"
          @click="stopScreenSharing"
        >
          ⏹️ Остановить
        </button>

        <!-- Включить / Выключить микрофон -->
        <button 
          class="btn flex items-center gap-1.5"
          :class="isMicMuted ? 'btn-danger-mute' : 'btn-ghost border border-[var(--border)]'"
          @click="toggleMic"
        >
          {{ isMicMuted ? '🎙️ Включить микрофон' : '🎙️ Выключить микрофон' }}
        </button>

        <!-- Заглушить весь звук -->
        <button 
          class="btn flex items-center gap-1.5"
          :class="isAllMuted ? 'btn-danger-mute' : 'btn-ghost border border-[var(--border)]'"
          @click="toggleAllMute"
        >
          {{ isAllMuted ? '🔊 Включить звук' : '🔇 Выключить звук' }}
        </button>
      </div>

      <div class="flex items-center gap-3">
        <button 
          class="btn btn-ghost border border-[var(--border)] text-xs" 
          @click="copyInviteLink"
        >
          📋 Копировать ссылку
        </button>
        <button 
          class="btn border border-red-500/30 text-red-400 text-xs bg-red-950/20 hover:bg-red-500 hover:text-white transition-all"
          @click="leaveCall"
        >
          📴 Выйти из комнаты
        </button>
      </div>
    </footer>

    <!-- Всплывающее уведомление о копировании -->
    <div v-if="copySuccess" class="toast-notification">
      ✓ Ссылка скопирована!
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch, reactive } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { useSignaling } from '@/composables/useSignaling'
import { useScreenShare } from '@/composables/useScreenShare'
import { useVoiceChat } from '@/composables/useVoiceChat'
import { useWebRTC } from '@/composables/useWebRTC'

const route = useRoute()
const router = useRouter()
const authStore = useAuthStore()

const roomId = route.params.roomId
const chatInput = ref('')
const copySuccess = ref(false)
const isMicMuted = ref(false)
const isAllMuted = ref(false)

const remoteParticipants = ref([])

// Volumetric state maps (userID -> 0 to 100)
const userVolumes = reactive({})
const streamVolumes = reactive({})

// Active displayed media stream
const activeStream = ref(null)
const activeStreamOwner = ref('')
const activeStreamOwnerId = ref(null)

// Initialize composables
const signaling = useSignaling()
const screenShare = useScreenShare()
const voiceChat = useVoiceChat()

const myUserID = authStore.user?.id
const myUsername = authStore.user?.username || 'Guest'

const webrtc = useWebRTC(signaling, voiceChat.stream, screenShare.stream, myUserID, myUsername)

const participantCount = computed(() => remoteParticipants.value.length + 1)

// Watch remote streams to set active stream automatically if none selected
watch(webrtc.remoteScreenStreams, (streams) => {
  const keys = Object.keys(streams)
  if (keys.length > 0 && !activeStream.value) {
    const firstPeerID = keys[0]
    const p = remoteParticipants.value.find(u => u.user_id == firstPeerID)
    selectActiveStream(streams[firstPeerID], firstPeerID, p ? p.username : 'Unknown')
  }
}, { deep: true })

onMounted(async () => {
  if (!roomId || !authStore.accessToken) {
    router.push('/screenshare')
    return
  }

  // Set up signaling websocket hooks
  signaling.on('joined', (msg) => {
    remoteParticipants.value = msg.participants.filter(p => p.user_id !== myUserID)
    
    // Set default volumes
    remoteParticipants.value.forEach(p => {
      if (userVolumes[p.user_id] === undefined) userVolumes[p.user_id] = 100
      if (streamVolumes[p.user_id] === undefined) streamVolumes[p.user_id] = 100
    })
  })

  signaling.on('user_joined', async (msg) => {
    console.log(`[Room] User joined room: ${msg.username} (${msg.user_id})`)
    
    // Set default volumes
    if (userVolumes[msg.user_id] === undefined) userVolumes[msg.user_id] = 100
    if (streamVolumes[msg.user_id] === undefined) streamVolumes[msg.user_id] = 100

    // Add to participants list
    if (!remoteParticipants.value.some(p => p.user_id === msg.user_id)) {
      remoteParticipants.value.push({
        user_id: msg.user_id,
        username: msg.username
      })
    }
    
    // Instantiate Peer Connection & Create Offer
    await webrtc.handleUserJoined(msg.user_id)
  })

  signaling.on('user_left', (msg) => {
    console.log(`[Room] User left room: ${msg.user_id}`)
    
    remoteParticipants.value = remoteParticipants.value.filter(p => p.user_id !== msg.user_id)
    webrtc.closePeerConnection(msg.user_id)

    // Reset active stream if it was that user's stream
    if (activeStreamOwnerId.value == msg.user_id || activeStream.value === webrtc.remoteScreenStreams[msg.user_id]) {
      activeStream.value = null
      activeStreamOwner.value = ''
      activeStreamOwnerId.value = null
    }
  })

  signaling.on('offer', async (msg) => {
    console.log(`[Room] Received WebRTC offer from: ${msg.from_id}`)
    await webrtc.handleOffer(msg.from_id, msg.sdp)
  })

  signaling.on('answer', async (msg) => {
    console.log(`[Room] Received WebRTC answer from: ${msg.from_id}`)
    await webrtc.handleAnswer(msg.from_id, msg.sdp)
  })

  signaling.on('ice_candidate', async (msg) => {
    await webrtc.handleIceCandidate(msg.from_id, msg.candidate)
  })

  signaling.on('error', (msg) => {
    alert(`Ошибка сигнального сервера: ${msg.message}`)
    router.push('/screenshare')
  })

  // Connect to Go signaling server
  signaling.connect(roomId, authStore.accessToken)

  // Start voice chat automatically on join
  try {
    const micStream = await voiceChat.start()
    webrtc.updateLocalVoiceStream(micStream)
  } catch (err) {
    console.warn('[Room] Failed to capture microphone:', err)
  }
})

onUnmounted(() => {
  voiceChat.stop()
  screenShare.stop()
  webrtc.clearAll()
  signaling.disconnect()
})

// ── Voice / Mic Controls ──

function toggleMic() {
  isMicMuted.value = !isMicMuted.value
  voiceChat.toggleMute(isMicMuted.value)
}

function toggleAllMute() {
  isAllMuted.value = !isAllMuted.value
}

// ── Screensharing Triggers ──

async function startScreenSharing() {
  try {
    const stream = await screenShare.start()
    webrtc.updateLocalScreenStream(stream)
    
    // Automatically display own screenshare in main screen if no remote shares
    if (!activeStream.value) {
      selectActiveStream(stream, myUserID, 'Вы')
    }
  } catch (err) {
    console.error('Failed to share screen:', err)
  }
}

function stopScreenSharing() {
  screenShare.stop()
  webrtc.updateLocalScreenStream(null)
  
  if (activeStream.value === screenShare.stream.value) {
    activeStream.value = null
    activeStreamOwner.value = ''
    activeStreamOwnerId.value = null
  }
}

// ── Control Bar Helpers ──

function selectActiveStream(stream, ownerId, ownerName) {
  activeStream.value = stream
  activeStreamOwner.value = ownerName
  activeStreamOwnerId.value = ownerId
}

function copyInviteLink() {
  const inviteUrl = window.location.href
  navigator.clipboard.writeText(inviteUrl).then(() => {
    copySuccess.value = true
    setTimeout(() => {
      copySuccess.value = false
    }, 2000)
  })
}

function leaveCall() {
  voiceChat.stop()
  stopScreenSharing()
  webrtc.clearAll()
  signaling.disconnect()
  router.push('/screenshare')
}

// ── WebRTC Connection Classes ──

function getRTCStatusClass(userID) {
  const state = webrtc.connectionStates[userID]
  if (state === 'connected') return 'connected'
  if (state === 'connecting' || state === 'checking') return 'connecting'
  if (state === 'failed' || state === 'disconnected') return 'failed'
  return ''
}

function getRTCStatusColor(userID) {
  const type = webrtc.peerTypes[userID]
  if (type === 'P2P') return '#4ade80' // Green
  if (type === 'TURN') return '#fbbf24' // Yellow (delay warning)
  return '#94a3b8' // Slate (connecting)
}

// ── Chat Helpers ──

function sendMessage() {
  const text = chatInput.value.trim()
  if (!text) return
  
  webrtc.sendChatMessage(text)
  chatInput.value = ''
}

function formatTime(timestamp) {
  const d = new Date(timestamp)
  return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}
</script>

<style scoped>
.room-page {
  height: 100vh;
  display: flex;
  flex-direction: column;
  background: var(--bg-base);
  color: var(--text-primary);
  overflow: hidden;
}

/* Навбар */
.room-nav {
  border-radius: 0;
  border: none;
  border-bottom: 1px solid var(--border);
  flex-shrink: 0;
}
.room-nav-inner {
  height: 48px;
}
.ping-dot {
  width: 8px;
  height: 8px;
  background: #10b981;
  border-radius: 50%;
  box-shadow: 0 0 8px #10b981;
}

.badge-warning {
  background: rgba(245, 158, 11, 0.15);
  color: #fbbf24;
  border: 1px solid rgba(245, 158, 11, 0.3);
}

/* Рабочая область видео */
.video-workspace {
  background: rgba(0, 0, 0, 0.2);
}

.primary-screen-holder {
  background: rgba(0, 0, 0, 0.4);
}
.primary-video {
  max-height: 70vh;
}
.active-stream-badge {
  background: rgba(0,0,0,0.65);
  backdrop-filter: blur(4px);
  border: 1px solid var(--border);
  color: white;
}

/* Карточки участников */
.section-title {
  font-size: 11px;
  font-weight: 700;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
  margin-bottom: 12px;
}
.participants-grid {
  display: flex;
}
.participant-card {
  width: 200px;
  background: rgba(255, 255, 255, 0.02);
  transition: border-color 0.15s, transform 0.15s;
}
.participant-card:hover {
  transform: translateY(-2px);
  border-color: rgba(99, 102, 241, 0.3);
}
.status-dot-indicator {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #94a3b8;
}
.status-dot-indicator.connected {
  background: #22c55e;
  box-shadow: 0 0 6px #22c55e;
}
.status-dot-indicator.connecting {
  background: #eab308;
  box-shadow: 0 0 6px #eab308;
}
.status-dot-indicator.failed {
  background: #ef4444;
  box-shadow: 0 0 6px #ef4444;
}

/* Чат-сайдбар */
.chat-sidebar {
  background: rgba(255, 255, 255, 0.015);
  flex-shrink: 0;
}
.chat-messages {
  scrollbar-width: thin;
}

/* Контрол-бар */
.room-controls-bar {
  border-radius: 0;
  border-left: none;
  border-right: none;
  border-bottom: none;
  flex-shrink: 0;
}

.btn-danger-mute {
  background: rgba(239, 68, 68, 0.15);
  border: 1px solid rgba(239, 68, 68, 0.35);
  color: #fca5a5;
  transition: all 0.15s;
}
.btn-danger-mute:hover {
  background: #ef4444;
  color: white;
  border-color: #ef4444;
}

.volume-slider {
  accent-color: var(--primary);
  height: 4px;
}

.toast-notification {
  position: fixed;
  bottom: 80px;
  right: 24px;
  background: rgba(99, 102, 241, 0.9);
  color: white;
  padding: 10px 20px;
  border-radius: 8px;
  font-size: 13px;
  font-weight: 600;
  z-index: 1000;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
  animation: slideUp 0.2s ease-out;
}

@keyframes slideUp {
  from { transform: translateY(20px); opacity: 0; }
  to { transform: translateY(0); opacity: 1; }
}
</style>
