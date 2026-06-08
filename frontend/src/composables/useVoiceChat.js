import { ref } from 'vue'

export function useVoiceChat() {
  const stream = ref(null)
  const active = ref(false) // Whether voice is unmuted and transmitting
  const error = ref(null)

  async function start() {
    error.value = null
    try {
      const localStream = await navigator.mediaDevices.getUserMedia({
        audio: {
          echoCancellation: true,
          noiseSuppression: true,
          autoGainControl: true
        }
      })
      stream.value = localStream
      active.value = true
      return localStream
    } catch (err) {
      error.value = err.message || 'Ошибка захвата микрофона'
      active.value = false
      throw err
    }
  }

  function stop() {
    if (stream.value) {
      stream.value.getTracks().forEach(track => track.stop())
      stream.value = null
    }
    active.value = false
  }

  function toggleMute(mute) {
    if (stream.value) {
      stream.value.getAudioTracks().forEach(track => {
        track.enabled = !mute
      })
      active.value = !mute
    }
  }

  return {
    stream,
    active,
    error,
    start,
    stop,
    toggleMute
  }
}
