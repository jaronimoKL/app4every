import { reactive, ref, onUnmounted } from 'vue'
import { useAuthStore } from '@/stores/auth'

export function useWatchParty() {
  const ws = ref(null)
  const roomState = reactive({
    isConnected: false,
    roomId: '',
    yourId: 0,
    videoUrl: '',
    videoType: '',
    isPlaying: false,
    currentTime: 0,
    isOwner: false,
    participants: [],
    knockRequests: [], // for owner
    error: null,
  })

  const playerRef = ref(null)

  function connect(roomId, token) {
    if (ws.value) ws.value.close()

    roomState.roomId = roomId
    roomState.error = null
    roomState.isConnected = false
    roomState.knockRequests = []

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const authStore = useAuthStore()
    const username = encodeURIComponent(authStore.user?.username || '')
    const wsUrl = `${protocol}//${window.location.host}/api/v1/watchparty/ws?room_id=${roomId}&token=${token}&username=${username}`

    ws.value = new WebSocket(wsUrl)

    ws.value.onopen = () => {
      // Not fully connected until we get 'joined'
    }

    ws.value.onmessage = (event) => {
      const msg = JSON.parse(event.data)
      handleMessage(msg)
    }

    ws.value.onclose = () => {
      roomState.isConnected = false
      if (!roomState.error) {
        roomState.error = 'Connection lost'
      }
    }

    ws.value.onerror = () => {
      roomState.error = 'WebSocket error'
    }
  }

  function handleMessage(msg) {
    switch (msg.type) {
      case 'joined':
        roomState.isConnected = true
        roomState.yourId = msg.your_id
        roomState.isOwner = msg.is_owner
        roomState.videoUrl = msg.video_url || ''
        roomState.videoType = msg.video_type || ''
        roomState.isPlaying = msg.is_playing || false
        roomState.currentTime = msg.current_time || 0
        roomState.participants = msg.participants || []
        
        // Sync player if it exists
        if (roomState.isPlaying) {
          playerRef.value?.syncPlay(roomState.currentTime)
        } else {
          playerRef.value?.syncPause(roomState.currentTime)
        }
        break

      case 'play':
        roomState.isPlaying = true
        roomState.currentTime = msg.current_time
        playerRef.value?.syncPlay(msg.current_time)
        break

      case 'pause':
        roomState.isPlaying = false
        roomState.currentTime = msg.current_time
        playerRef.value?.syncPause(msg.current_time)
        break

      case 'seek':
        roomState.currentTime = msg.current_time
        playerRef.value?.syncSeek(msg.current_time)
        break

      case 'video_changed':
        roomState.videoUrl = msg.video_url
        roomState.videoType = msg.video_type
        roomState.isPlaying = false
        roomState.currentTime = 0
        playerRef.value?.syncPause(0)
        break

      case 'user_joined':
        if (!roomState.participants.some(p => p.user_id === msg.user_id)) {
          roomState.participants.push({
            user_id: msg.user_id,
            username: msg.username,
            is_owner: false
          })
        }
        break

      case 'user_left':
        roomState.participants = roomState.participants.filter(p => p.user_id !== msg.user_id)
        break

      case 'knock_request':
        roomState.knockRequests.push({
          user_id: msg.user_id,
          username: msg.username
        })
        break

      case 'knock_rejected':
        roomState.error = 'Your request to join was rejected.'
        ws.value?.close()
        break

      case 'kicked':
        roomState.error = 'You were kicked from the room.'
        ws.value?.close()
        break

      case 'error':
        if (msg.code === 'waiting_for_approval') {
          roomState.error = 'Waiting for owner approval...'
        } else {
          roomState.error = msg.code
        }
        break
    }
  }

  function disconnect() {
    if (ws.value) {
      ws.value.close()
      ws.value = null
    }
  }

  onUnmounted(() => {
    disconnect()
  })

  function onLocalPlay(currentTime) {
    if (ws.value?.readyState === WebSocket.OPEN) {
      ws.value.send(JSON.stringify({ type: 'play', current_time: currentTime }))
    }
  }

  function onLocalPause(currentTime) {
    if (ws.value?.readyState === WebSocket.OPEN) {
      ws.value.send(JSON.stringify({ type: 'pause', current_time: currentTime }))
    }
  }

  function onLocalSeek(currentTime) {
    if (ws.value?.readyState === WebSocket.OPEN) {
      ws.value.send(JSON.stringify({ type: 'seek', current_time: currentTime }))
    }
  }

  function changeVideo(url, videoType) {
    if (roomState.isOwner && ws.value?.readyState === WebSocket.OPEN) {
      ws.value.send(JSON.stringify({ type: 'change_video', video_url: url, video_type: videoType }))
    }
  }

  function admitUser(userId) {
    if (roomState.isOwner && ws.value?.readyState === WebSocket.OPEN) {
      ws.value.send(JSON.stringify({ type: 'admit', target_id: userId }))
      roomState.knockRequests = roomState.knockRequests.filter(r => r.user_id !== userId)
    }
  }

  function rejectUser(userId) {
    if (roomState.isOwner && ws.value?.readyState === WebSocket.OPEN) {
      ws.value.send(JSON.stringify({ type: 'reject', target_id: userId }))
      roomState.knockRequests = roomState.knockRequests.filter(r => r.user_id !== userId)
    }
  }

  function kickUser(userId) {
    if (roomState.isOwner && ws.value?.readyState === WebSocket.OPEN) {
      ws.value.send(JSON.stringify({ type: 'kick', target_id: userId }))
    }
  }

  function updateMetadata(shikimoriId, anilibertyAlias) {
    if (roomState.isOwner && ws.value?.readyState === WebSocket.OPEN) {
      ws.value.send(JSON.stringify({
        type: 'update_metadata',
        shikimori_id: shikimoriId || '',
        aniliberty_alias: anilibertyAlias || ''
      }))
    }
  }

  return {
    roomState,
    playerRef,
    connect,
    disconnect,
    onLocalPlay,
    onLocalPause,
    onLocalSeek,
    changeVideo,
    admitUser,
    rejectUser,
    kickUser,
    updateMetadata
  }
}
