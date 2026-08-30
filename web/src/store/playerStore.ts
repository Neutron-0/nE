import { create } from 'zustand'
import type { Track } from '../types'
import { api } from '../lib/api'
import { audioEngine, type PresetName, type DSPState } from '../lib/audioEngine'

interface PlayerState {
  currentTrack: Track | null
  isPlaying: boolean
  currentTime: number
  duration: number
  volume: number
  isMuted: boolean
  isShuffle: boolean
  repeatMode: 'off' | 'all' | 'one'
  queue: Track[]
  queueIndex: number
  hasScrobbled: boolean

  // Studio Audiophile DSP State
  dspState: DSPState

  // Actions
  playTrack: (track: Track, newQueue?: Track[]) => void
  togglePlay: () => void
  pause: () => void
  resume: () => void
  next: () => void
  previous: () => void
  seek: (seconds: number) => void
  setVolume: (volume: number) => void
  toggleMute: () => void
  toggleShuffle: () => void
  toggleRepeat: () => void
  addToQueue: (track: Track) => void
  toggleStarCurrentTrack: () => Promise<void>

  // DSP Actions
  setDspEnabled: (enabled: boolean) => void
  setDspPreset: (preset: PresetName) => void
  setPreAmpGain: (gain: number) => void
  setEQBand: (index: number, gain: number) => void
  setCompressorEnabled: (enabled: boolean) => void
}

// Singleton Audio Element
let audio: HTMLAudioElement | null = null

export function getAudio(): HTMLAudioElement {
  if (!audio) {
    audio = new Audio()
    audio.preload = 'auto'
    audio.crossOrigin = 'anonymous'
  }
  return audio
}

export const usePlayerStore = create<PlayerState>((set, get) => {
  const initialVolume = parseFloat(localStorage.getItem('ne_volume') || '0.8')

  return {
    currentTrack: null,
    isPlaying: false,
    currentTime: 0,
    duration: 0,
    volume: initialVolume,
    isMuted: false,
    isShuffle: false,
    repeatMode: 'off',
    queue: [],
    queueIndex: -1,
    hasScrobbled: false,

    dspState: audioEngine.getState(),

    setDspEnabled: (enabled: boolean) => {
      audioEngine.setEnabled(enabled)
      set({ dspState: audioEngine.getState() })
    },

    setDspPreset: (preset: PresetName) => {
      audioEngine.setPreset(preset)
      set({ dspState: audioEngine.getState() })
    },

    setPreAmpGain: (gain: number) => {
      audioEngine.setPreAmpGain(gain)
      set({ dspState: audioEngine.getState() })
    },

    setEQBand: (index: number, gain: number) => {
      audioEngine.setEQBand(index, gain)
      set({ dspState: audioEngine.getState() })
    },

    setCompressorEnabled: (enabled: boolean) => {
      audioEngine.setCompressorEnabled(enabled)
      set({ dspState: audioEngine.getState() })
    },

    playTrack: (track: Track, newQueue?: Track[]) => {
      const a = getAudio()
      audioEngine.init(a)
      audioEngine.resume()

      const streamUrl = api.getStreamUrl(track.id)

      let queue = get().queue
      let queueIndex = get().queueIndex

      if (newQueue) {
        queue = newQueue
        queueIndex = newQueue.findIndex((t) => t.id === track.id)
        if (queueIndex === -1) queueIndex = 0
      } else {
        const existingIdx = queue.findIndex((t) => t.id === track.id)
        if (existingIdx !== -1) {
          queueIndex = existingIdx
        } else {
          queue = [...queue, track]
          queueIndex = queue.length - 1
        }
      }

      a.src = streamUrl
      a.volume = get().isMuted ? 0 : get().volume

      a.onloadedmetadata = () => {
        set({ duration: a.duration || track.duration, hasScrobbled: false })
      }

      a.ontimeupdate = () => {
        const curTime = a.currentTime
        set({ currentTime: curTime })

        // Scrobble trigger: 50% or 240 seconds
        const { currentTrack, duration, hasScrobbled } = get()
        if (currentTrack && duration > 0 && !hasScrobbled) {
          if (curTime >= duration * 0.5 || curTime >= 240) {
            set({ hasScrobbled: true })
            api.scrobble(currentTrack.id, curTime, false)
          }
        }
      }

      a.onended = () => {
        const { currentTrack, duration, repeatMode, next } = get()
        if (currentTrack) {
          api.scrobble(currentTrack.id, duration, true)
        }
        if (repeatMode === 'one') {
          a.currentTime = 0
          a.play()
        } else {
          next()
        }
      }

      a.play()
        .then(() => {
          set({ currentTrack: track, isPlaying: true, queue, queueIndex, hasScrobbled: false })
          updateMediaSession(track)
        })
        .catch((err) => {
          console.error('Audio play error:', err)
          set({ isPlaying: false })
        })
    },

    togglePlay: () => {
      const a = getAudio()
      audioEngine.init(a)
      audioEngine.resume()

      if (get().isPlaying) {
        a.pause()
        set({ isPlaying: false })
      } else {
        if (a.src) {
          a.play().then(() => set({ isPlaying: true }))
        } else if (get().currentTrack) {
          get().playTrack(get().currentTrack!)
        }
      }
    },

    pause: () => {
      const a = getAudio()
      a.pause()
      set({ isPlaying: false })
    },

    resume: () => {
      const a = getAudio()
      audioEngine.init(a)
      audioEngine.resume()
      if (a.src) {
        a.play().then(() => set({ isPlaying: true }))
      }
    },

    next: () => {
      const { queue, queueIndex, isShuffle, repeatMode, playTrack } = get()
      if (queue.length === 0) return

      if (isShuffle) {
        const randomIdx = Math.floor(Math.random() * queue.length)
        playTrack(queue[randomIdx])
        return
      }

      if (queueIndex < queue.length - 1) {
        playTrack(queue[queueIndex + 1])
      } else if (repeatMode === 'all') {
        playTrack(queue[0])
      }
    },

    previous: () => {
      const { queue, queueIndex, playTrack, currentTime, seek } = get()
      if (currentTime > 3) {
        seek(0)
        return
      }

      if (queueIndex > 0) {
        playTrack(queue[queueIndex - 1])
      } else if (queue.length > 0) {
        playTrack(queue[queue.length - 1])
      }
    },

    seek: (seconds: number) => {
      const a = getAudio()
      a.currentTime = seconds
      set({ currentTime: seconds })
    },

    setVolume: (volume: number) => {
      const a = getAudio()
      const clamped = Math.max(0, Math.min(1, volume))
      a.volume = clamped
      localStorage.setItem('ne_volume', clamped.toString())
      set({ volume: clamped, isMuted: clamped === 0 })
    },

    toggleMute: () => {
      const a = getAudio()
      const { isMuted, volume } = get()
      if (isMuted) {
        a.volume = volume > 0 ? volume : 0.8
        set({ isMuted: false })
      } else {
        a.volume = 0
        set({ isMuted: true })
      }
    },

    toggleShuffle: () => {
      set((state) => ({ isShuffle: !state.isShuffle }))
    },

    toggleRepeat: () => {
      set((state) => {
        if (state.repeatMode === 'off') return { repeatMode: 'all' }
        if (state.repeatMode === 'all') return { repeatMode: 'one' }
        return { repeatMode: 'off' }
      })
    },

    addToQueue: (track: Track) => {
      set((state) => ({ queue: [...state.queue, track] }))
    },

    toggleStarCurrentTrack: async () => {
      const { currentTrack } = get()
      if (!currentTrack) return

      const newStarred = !currentTrack.isStarred
      try {
        await api.starItem('track', currentTrack.id, newStarred)
        set((state) => ({
          currentTrack: state.currentTrack ? { ...state.currentTrack, isStarred: newStarred } : null,
          queue: state.queue.map((t) => (t.id === currentTrack.id ? { ...t, isStarred: newStarred } : t)),
        }))
      } catch (err) {
        console.error('Failed to star track:', err)
      }
    },
  }
})

function updateMediaSession(track: Track) {
  if ('mediaSession' in navigator) {
    navigator.mediaSession.metadata = new MediaMetadata({
      title: track.title,
      artist: track.rawArtist || track.albumArtist || 'Unknown Artist',
      album: track.albumTitle || 'nE Audio',
      artwork: [
        { src: api.getArtworkUrl('track', track.id, 96), sizes: '96x96', type: 'image/jpeg' },
        { src: api.getArtworkUrl('track', track.id, 256), sizes: '256x256', type: 'image/jpeg' },
        { src: api.getArtworkUrl('track', track.id, 512), sizes: '512x512', type: 'image/jpeg' },
      ],
    })

    navigator.mediaSession.setActionHandler('play', () => usePlayerStore.getState().togglePlay())
    navigator.mediaSession.setActionHandler('pause', () => usePlayerStore.getState().pause())
    navigator.mediaSession.setActionHandler('previoustrack', () => usePlayerStore.getState().previous())
    navigator.mediaSession.setActionHandler('nexttrack', () => usePlayerStore.getState().next())
  }
}
