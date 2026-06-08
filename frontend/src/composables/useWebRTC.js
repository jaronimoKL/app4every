import { reactive } from 'vue'

const iceConfig = {
  iceServers: [
    { urls: 'stun:stun.l.google.com:19302' },
    { urls: 'stun:stun1.l.google.com:19302' },
    { urls: 'stun:stun.cloudflare.com:3478' }
  ],
  iceCandidatePoolSize: 10
}

export function useWebRTC(signaling, localVoiceStream, localScreenStream, myUserID, myUsername) {
  const peerConnections = {}
  const dataChannels = {}
  const remoteScreenStreams = reactive({})
  const remoteVoiceStreams = reactive({})
  const connectionStates = reactive({})
  const peerTypes = reactive({}) // userID -> 'P2P' | 'TURN' | 'unknown'
  const chatMessages = reactive([])
  const iceCandidateQueues = {} // userID -> array of candidates

  function clearAll() {
    Object.keys(peerConnections).forEach(id => {
      closePeerConnection(id)
    })
    Object.keys(remoteScreenStreams).forEach(id => {
      delete remoteScreenStreams[id]
    })
    Object.keys(remoteVoiceStreams).forEach(id => {
      delete remoteVoiceStreams[id]
    })
    Object.keys(connectionStates).forEach(id => {
      delete connectionStates[id]
    })
    Object.keys(peerTypes).forEach(id => {
      delete peerTypes[id]
    })
    Object.keys(iceCandidateQueues).forEach(id => {
      delete iceCandidateQueues[id]
    })
    chatMessages.length = 0
  }

  function closePeerConnection(userID) {
    if (dataChannels[userID]) {
      dataChannels[userID].close()
      delete dataChannels[userID]
    }
    if (peerConnections[userID]) {
      peerConnections[userID].close()
      delete peerConnections[userID]
    }
    delete remoteScreenStreams[userID]
    delete remoteVoiceStreams[userID]
    delete connectionStates[userID]
    delete peerTypes[userID]
    delete iceCandidateQueues[userID]
  }

  // ── Prefer VP9 / H264 & Max Bitrate ──

  function preferHighQualityCodecs(pc) {
    const sender = pc.getSenders().find(s => s.track?.kind === 'video')
    if (!sender) return
    try {
      const params = sender.getParameters()
      params.encodings = [{
        maxBitrate: 20000000, // 20 Mbps max for 4K stream
        maxFramerate: 60,
        scaleResolutionDownBy: 1
      }]
      sender.setParameters(params)
    } catch (e) {
      console.warn('[WebRTC] Could not set encoding parameters:', e)
    }
  }

  function setCodecPreferences(pc) {
    const transceivers = pc.getTransceivers()
    if (transceivers.length === 0) return
    const transceiver = transceivers[0]
    if (typeof transceiver.setCodecPreferences !== 'function') return

    try {
      const { codecs } = RTCRtpSender.getCapabilities('video')
      const preferred = codecs
        .filter(c => c.mimeType === 'video/VP9' || c.mimeType === 'video/H264')
        .sort((a, b) => a.mimeType.includes('VP9') ? -1 : 1)
      transceiver.setCodecPreferences([
        ...preferred,
        ...codecs.filter(c => c.mimeType !== 'video/VP9' && c.mimeType !== 'video/H264')
      ])
    } catch (e) {
      console.warn('[WebRTC] Could not set codec preferences:', e)
    }
  }

  // ── Connection Initialization ──

  async function createPeerConnection(targetUserID, isInitiator) {
    // REUSE existing connection if present to support renegotiation
    let pc = peerConnections[targetUserID]
    if (pc) {
      return pc
    }

    pc = new RTCPeerConnection(iceConfig)
    peerConnections[targetUserID] = pc
    connectionStates[targetUserID] = 'connecting'
    peerTypes[targetUserID] = 'unknown'

    // Add local tracks (microphone)
    if (localVoiceStream.value) {
      localVoiceStream.value.getTracks().forEach(track => {
        pc.addTrack(track, localVoiceStream.value)
      })
    }

    // Add local tracks (screenshare)
    if (localScreenStream.value) {
      localScreenStream.value.getTracks().forEach(track => {
        pc.addTrack(track, localScreenStream.value)
      })
    }

    // ICE candidate gathering
    pc.onicecandidate = ({ candidate }) => {
      if (candidate) {
        signaling.send({
          type: 'ice_candidate',
          target_id: targetUserID,
          candidate: candidate
        })
      }
    }

    // Remote track added by peer
    pc.ontrack = (event) => {
      const stream = event.streams[0]
      if (!stream) return

      const isScreen = event.track.kind === 'video' || stream.getVideoTracks().length > 0

      if (isScreen) {
        remoteScreenStreams[targetUserID] = stream
        // Clean up stream if the tracks stop/end
        event.track.onended = () => {
          if (stream.getVideoTracks().every(t => t.readyState === 'ended')) {
            delete remoteScreenStreams[targetUserID]
          }
        }
      } else {
        remoteVoiceStreams[targetUserID] = stream
        event.track.onended = () => {
          if (stream.getAudioTracks().every(t => t.readyState === 'ended')) {
            delete remoteVoiceStreams[targetUserID]
          }
        }
      }
    }

    // Connection state changes
    pc.onconnectionstatechange = () => {
      connectionStates[targetUserID] = pc.connectionState
      if (pc.connectionState === 'failed') {
        restartIce(targetUserID)
      }
    }

    // Monitor connection candidate type (P2P vs TURN)
    const statsInterval = setInterval(() => {
      if (!peerConnections[targetUserID] || pc.connectionState === 'closed') {
        clearInterval(statsInterval)
        return
      }
      checkConnectionStats(pc, targetUserID)
    }, 5000)

    // Data Channel (P2P Text Chat)
    if (isInitiator) {
      const dc = pc.createDataChannel('chat', { ordered: true })
      setupDataChannel(targetUserID, dc)
    } else {
      pc.ondatachannel = (event) => {
        setupDataChannel(targetUserID, event.channel)
      }
    }

    return pc
  }

  function setupDataChannel(targetUserID, dc) {
    dataChannels[targetUserID] = dc
    dc.onmessage = ({ data }) => {
      try {
        const msg = JSON.parse(data)
        if (msg.type === 'chat') {
          chatMessages.push({
            ...msg,
            own: false
          })
        }
      } catch (e) {
        console.error('[WebRTC] DataChannel parse error:', e)
      }
    }
    dc.onclose = () => {
      delete dataChannels[targetUserID]
    }
  }

  async function checkConnectionStats(pc, userID) {
    try {
      const stats = await pc.getStats()
      let isRelay = false
      stats.forEach(report => {
        if (report.type === 'candidate-pair' && report.state === 'succeeded') {
          const localCand = stats.get(report.localCandidateId)
          const remoteCand = stats.get(report.remoteCandidateId)
          if (localCand?.candidateType === 'relay' || remoteCand?.candidateType === 'relay') {
            isRelay = true
          }
        }
      })
      peerTypes[userID] = isRelay ? 'TURN' : 'P2P'
    } catch (e) {
      // Stats might not be ready
    }
  }

  // ── Signaling Handlers ──

  async function handleUserJoined(targetUserID) {
    const pc = await createPeerConnection(targetUserID, true)
    setCodecPreferences(pc)
    
    const offer = await pc.createOffer()
    await pc.setLocalDescription(offer)
    preferHighQualityCodecs(pc)

    signaling.send({
      type: 'offer',
      target_id: targetUserID,
      sdp: offer.sdp
    })
  }

  async function handleOffer(fromUserID, sdp) {
    const pc = await createPeerConnection(fromUserID, false)
    setCodecPreferences(pc)
    
    await pc.setRemoteDescription(new RTCSessionDescription({ type: 'offer', sdp }))
    
    // Drain queued ICE candidates now that remote description is set
    await drainIceQueue(fromUserID)

    const answer = await pc.createAnswer()
    await pc.setLocalDescription(answer)
    preferHighQualityCodecs(pc)

    signaling.send({
      type: 'answer',
      target_id: fromUserID,
      sdp: answer.sdp
    })
  }

  async function handleAnswer(fromUserID, sdp) {
    const pc = peerConnections[fromUserID]
    if (pc) {
      await pc.setRemoteDescription(new RTCSessionDescription({ type: 'answer', sdp }))
      
      // Drain queued ICE candidates now that remote description is set
      await drainIceQueue(fromUserID)
    }
  }

  async function handleIceCandidate(fromUserID, candidate) {
    const pc = peerConnections[fromUserID]
    if (pc && pc.remoteDescription && pc.remoteDescription.type) {
      try {
        await pc.addIceCandidate(new RTCIceCandidate(candidate))
      } catch (e) {
        console.warn('[WebRTC] Failed to add ICE candidate:', e)
      }
    } else {
      // Queue candidate if connection not ready or remote description not set
      if (!iceCandidateQueues[fromUserID]) {
        iceCandidateQueues[fromUserID] = []
      }
      iceCandidateQueues[fromUserID].push(candidate)
    }
  }

  async function drainIceQueue(userID) {
    const pc = peerConnections[userID]
    const queue = iceCandidateQueues[userID]
    if (pc && queue && queue.length > 0) {
      while (queue.length > 0) {
        const candidate = queue.shift()
        try {
          await pc.addIceCandidate(new RTCIceCandidate(candidate))
        } catch (e) {
          console.warn('[WebRTC] Failed to add queued ICE candidate:', e)
        }
      }
    }
  }

  async function restartIce(targetUserID) {
    const pc = peerConnections[targetUserID]
    if (!pc) return
    console.warn(`[WebRTC] ICE connection failed for peer ${targetUserID}. Restarting ICE...`)
    
    try {
      const offer = await pc.createOffer({ iceRestart: true })
      await pc.setLocalDescription(offer)
      signaling.send({
        type: 'offer',
        target_id: targetUserID,
        sdp: offer.sdp
      })
    } catch (e) {
      console.error('[WebRTC] Failed to restart ICE:', e)
    }
  }

  // ── Track Sync ──

  function updateLocalVoiceStream(newStream) {
    Object.keys(peerConnections).forEach(userID => {
      const pc = peerConnections[userID]
      
      if (!newStream) {
        // Remove microphone sender
        const senders = pc.getSenders()
        senders.forEach(sender => {
          if (sender.track?.kind === 'audio' && (!localScreenStream.value || !localScreenStream.value.getAudioTracks().includes(sender.track))) {
            pc.removeTrack(sender)
          }
        })
        triggerRenegotiation(userID)
        return
      }

      newStream.getAudioTracks().forEach(track => {
        // Find existing voice sender
        const sender = pc.getSenders().find(s => s.track?.kind === 'audio' && (!localScreenStream.value || !localScreenStream.value.getAudioTracks().includes(s.track)))
        if (sender) {
          sender.replaceTrack(track)
        } else {
          pc.addTrack(track, newStream)
          triggerRenegotiation(userID)
        }
      })
    })
  }

  function updateLocalScreenStream(newStream) {
    Object.keys(peerConnections).forEach(userID => {
      const pc = peerConnections[userID]
      
      if (!newStream) {
        // Remove screen share senders
        const senders = pc.getSenders()
        senders.forEach(sender => {
          if (sender.track?.kind === 'video' || (sender.track?.kind === 'audio' && (!localVoiceStream.value || !localVoiceStream.value.getAudioTracks().includes(sender.track)))) {
            pc.removeTrack(sender)
          }
        })
        triggerRenegotiation(userID)
        return
      }

      let negotiationNeeded = false
      newStream.getTracks().forEach(track => {
        const sender = pc.getSenders().find(s => s.track?.kind === track.kind && (track.kind === 'video' || (localVoiceStream.value && !localVoiceStream.value.getAudioTracks().includes(s.track))))
        if (sender) {
          sender.replaceTrack(track)
        } else {
          pc.addTrack(track, newStream)
          negotiationNeeded = true
        }
      })

      if (negotiationNeeded) {
        triggerRenegotiation(userID)
      }
    })
  }

  async function triggerRenegotiation(targetUserID) {
    const pc = peerConnections[targetUserID]
    if (!pc) return
    try {
      const offer = await pc.createOffer()
      await pc.setLocalDescription(offer)
      preferHighQualityCodecs(pc)
      signaling.send({
        type: 'offer',
        target_id: targetUserID,
        sdp: offer.sdp
      })
    } catch (e) {
      console.error('[WebRTC] Renegotiation failed:', e)
    }
  }

  // ── P2P DataChannel Chat ──

  function sendChatMessage(text) {
    const msg = {
      type: 'chat',
      from_id: myUserID,
      from_name: myUsername,
      text,
      ts: Date.now()
    }
    
    const payload = JSON.stringify(msg)
    Object.values(dataChannels).forEach(dc => {
      if (dc.readyState === 'open') {
        dc.send(payload)
      }
    })

    chatMessages.push({
      ...msg,
      own: true
    })
  }

  return {
    remoteScreenStreams,
    remoteVoiceStreams,
    connectionStates,
    peerTypes,
    chatMessages,
    handleUserJoined,
    handleOffer,
    handleAnswer,
    handleIceCandidate,
    closePeerConnection,
    updateLocalVoiceStream,
    updateLocalScreenStream,
    sendChatMessage,
    clearAll
  }
}
