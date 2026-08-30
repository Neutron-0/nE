import { create } from 'zustand'
import type { Track } from '../types'
import { api } from '../lib/api'

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

    playTrack: (track: Track, newQueue?: Track[]) => {
      const a = getAudio()
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

      let nextIndex = queueIndex + 1
      if (nextIndex >= queue.length) {
        if (repeatMode === 'all') {
          nextIndex = 0
        } else {
          set({ isPlaying: false })
          return
        }
      }

      playTrack(queue[nextIndex])
    },

    previous: () => {
      const { queue, queueIndex, playTrack, currentTime } = get()
      const a = getAudio()

      if (currentTime > 3) {
        a.currentTime = 0
        set({ currentTime: 0 })
        return
      }

      if (queue.length === 0) return
      let prevIndex = queueIndex - 1
      if (prevIndex < 0) {
        prevIndex = queue.length - 1
      }
      playTrack(queue[prevIndex])
    },

    seek: (seconds: number) => {
      const a = getAudio()
      a.currentTime = seconds
      set({ currentTime: seconds })
    },

    setVolume: (volume: number) => {
      const a = getAudio()
      const safeVol = Math.max(0, Math.min(1, volume))
      a.volume = safeVol
      localStorage.setItem('ne_volume', safeVol.toString())
      set({ volume: safeVol, isMuted: safeVol === 0 })
    },

    toggleMute: () => {
      const a = getAudio()
      const isMuted = !get().isMuted
      a.volume = isMuted ? 0 : get().volume
      set({ isMuted })
    },

    toggleShuffle: () => {
      set({ isShuffle: !get().isShuffle })
    },

    toggleRepeat: () => {
      const modes: ('off' | 'all' | 'one')[] = ['off', 'all', 'one']
      const currentIdx = modes.indexOf(get().repeatMode)
      const nextMode = modes[(currentIdx + 1) % modes.length]
      set({ repeatMode: nextMode })
    },

    addToQueue: (track: Track) => {
      set({ queue: [...get().queue, track] })
    },

    toggleStarCurrentTrack: async () => {
      const { currentTrack } = get()
      if (!currentTrack) return
      const nextStarred = !currentTrack.isStarred
      try {
        await api.starItem('track', currentTrack.id, nextStarred)
        set({ currentTrack: { ...currentTrack, isStarred: nextStarred } })
      } catch (err) {
        console.error('Star failed:', err)
      }
    },
  }
})

function updateMediaSession(track: Track) {
  if ('mediaSession' in navigator) {
    navigator.mediaSession.metadata = new MediaMetadata({
      title: track.title,
      artist: track.rawArtist || track.albumArtist || 'Unknown Artist',
      album: track.albumTitle || 'Unknown Album',
      artwork: [
        {
          src: api.getArtworkUrl('track', track.id, 512),
          sizes: '512x512',
          type: 'image/jpeg',
        },
      ],
    })

    navigator.mediaSession.setActionHandler('play', () => usePlayerStore.getState().resume())
    navigator.mediaSession.setActionHandler('pause', () => usePlayerStore.getState().pause())
    navigator.mediaSession.setActionHandler('previoustrack', () => usePlayerStore.getState().previous())
    navigator.mediaSession.setActionHandler('nexttrack', () => usePlayerStore.getState().next())
  }
}
