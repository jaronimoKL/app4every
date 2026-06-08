import { ref } from 'vue'

export function useSignaling() {
  const socket = ref(null)
  const connected = ref(false)
  const reconnectDelay = ref(1000)
  const messageQueue = []
  let pingInterval = null
  let reconnectTimeout = null
  let activeRoomId = null
  let activeToken = null

  // Callback handlers for different event types
  const handlers = {
    joined: [],
    user_joined: [],
    user_left: [],
    offer: [],
    answer: [],
    ice_candidate: [],
    error: [],
    close: []
  }

  function on(event, callback) {
    if (handlers[event]) {
      handlers[event].push(callback)
    }
  }

  function off(event, callback) {
    if (handlers[event]) {
      handlers[event] = handlers[event].filter(cb => cb !== callback)
    }
  }

  function trigger(event, data) {
    if (handlers[event]) {
      handlers[event].forEach(cb => cb(data))
    }
  }

  function connect(roomId, token) {
    disconnect()
    activeRoomId = roomId
    activeToken = token

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/api/v1/screenshare/ws?token=${token}&room_id=${roomId}`

    socket.value = new WebSocket(wsUrl)

    socket.value.onopen = () => {
      connected.value = true
      reconnectDelay.value = 1000
      console.log('[Signaling] WebSocket connected')

      // Drain message queue
      while (messageQueue.length > 0) {
        const msg = messageQueue.shift()
        send(msg)
      }

      // Start ping keep-alive
      startPing()
    }

    socket.value.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        if (msg.type === 'pong') return

        if (handlers[msg.type]) {
          trigger(msg.type, msg)
        }
      } catch (err) {
        console.error('[Signaling] Error parsing WS message:', err)
      }
    }

    socket.value.onclose = (event) => {
      connected.value = false
      stopPing()
      trigger('close', event)

      // Only reconnect if not disconnected intentionally via disconnect()
      if (activeRoomId) {
        console.warn(`[Signaling] WS disconnected, reconnecting in ${reconnectDelay.value}ms...`)
        reconnectTimeout = setTimeout(() => {
          reconnectDelay.value = Math.min(reconnectDelay.value * 2, 30000)
          connect(activeRoomId, activeToken)
        }, reconnectDelay.value)
      }
    }

    socket.value.onerror = (err) => {
      console.error('[Signaling] WS error:', err)
    }
  }

  function disconnect() {
    activeRoomId = null
    activeToken = null
    stopPing()
    if (reconnectTimeout) {
      clearTimeout(reconnectTimeout)
      reconnectTimeout = null
    }
    if (socket.value) {
      socket.value.close()
      socket.value = null
    }
    connected.value = false
  }

  function send(msg) {
    if (socket.value && socket.value.readyState === WebSocket.OPEN) {
      socket.value.send(JSON.stringify(msg))
    } else {
      messageQueue.push(msg)
    }
  }

  function startPing() {
    stopPing()
    pingInterval = setInterval(() => {
      send({ type: 'ping' })
    }, 25000)
  }

  function stopPing() {
    if (pingInterval) {
      clearInterval(pingInterval)
      pingInterval = null
    }
  }

  return {
    connected,
    connect,
    disconnect,
    send,
    on,
    off
  }
}
