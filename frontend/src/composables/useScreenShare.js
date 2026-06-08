import { ref } from 'vue'

export function useScreenShare() {
  const stream = ref(null)
  const active = ref(false)
  const resolution = ref('')
  const error = ref(null)

  async function start() {
    error.value = null
    try {
      const localStream = await navigator.mediaDevices.getDisplayMedia({
        video: {
          width:     { ideal: 3840, max: 3840 },    // 4K / 2160p
          height:    { ideal: 2160, max: 2160 },
          frameRate: { ideal: 60,   max: 60   },
          displaySurface: 'monitor',
          resizeMode: 'none',
        },
        audio: {
          echoCancellation: false,
          noiseSuppression: false,
          sampleRate: 48000,
        },
        selfBrowserSurface: 'exclude',
        surfaceSwitching: 'include',
      })

      stream.value = localStream
      active.value = true

      // Read real active resolution and framerate
      const videoTrack = localStream.getVideoTracks()[0]
      if (videoTrack) {
        const settings = videoTrack.getSettings()
        resolution.value = `${settings.width || 0}×${settings.height || 0} @ ${Math.round(settings.frameRate || 0)}fps`

        // Listen for track ending (e.g. user clicks "Stop sharing" native bar)
        videoTrack.onended = () => {
          stop()
        }
      }

      return localStream
    } catch (err) {
      error.value = err.message || 'Ошибка захвата экрана'
      active.value = false
      resolution.value = ''
      throw err
    }
  }

  function stop() {
    if (stream.value) {
      stream.value.getTracks().forEach(track => track.stop())
      stream.value = null
    }
    active.value = false
    resolution.value = ''
  }

  return {
    stream,
    active,
    resolution,
    error,
    start,
    stop
  }
}
