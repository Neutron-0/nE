import React, { useEffect, useState, useRef } from 'react'
import { X, Mic2, Music2 } from 'lucide-react'
import { usePlayerStore } from '../store/playerStore'
import { api } from '../lib/api'
import type { LyricsResult } from '../types'

interface LyricsOverlayProps {
  isOpen: boolean
  onClose: () => void
}

export const LyricsOverlay: React.FC<LyricsOverlayProps> = ({ isOpen, onClose }) => {
  const { currentTrack, currentTime, seek } = usePlayerStore()
  const [lyrics, setLyrics] = useState<LyricsResult | null>(null)
  const [loading, setLoading] = useState(false)
  const [activeLineIdx, setActiveLineIdx] = useState<number>(-1)
  const lineRefs = useRef<(HTMLDivElement | null)[]>([])

  useEffect(() => {
    if (!isOpen || !currentTrack) return

    let cancelled = false
    setLoading(true)
    api.getLyrics(currentTrack.id)
      .then((res) => {
        if (!cancelled) setLyrics(res)
      })
      .catch((err) => {
        console.warn('Could not fetch lyrics:', err)
        if (!cancelled) setLyrics(null)
      })
      .finally(() => {
        if (!cancelled) setLoading(false)
      })

    return () => {
      cancelled = true
    }
  }, [isOpen, currentTrack?.id])

  // Track active line based on current playback time
  useEffect(() => {
    if (!lyrics?.isSynced || !lyrics.lines.length) {
      setActiveLineIdx(-1)
      return
    }

    let active = -1
    for (let i = 0; i < lyrics.lines.length; i++) {
      if (currentTime >= lyrics.lines[i].time) {
        active = i
      } else {
        break
      }
    }

    setActiveLineIdx(active)

    // Smooth scroll to active line
    if (active >= 0 && lineRefs.current[active]) {
      lineRefs.current[active]?.scrollIntoView({
        behavior: 'smooth',
        block: 'center',
      })
    }
  }, [currentTime, lyrics])

  if (!isOpen || !currentTrack) return null

  const artworkUrl = api.getArtworkUrl('track', currentTrack.id, 600)

  return (
    <div className="fixed inset-0 z-50 bg-black/90 backdrop-blur-2xl flex flex-col transition-all duration-300 animate-in fade-in">
      {/* Blurred Album Artwork Background */}
      <div
        className="absolute inset-0 bg-cover bg-center opacity-25 blur-3xl pointer-events-none scale-125 transition-all duration-700"
        style={{ backgroundImage: `url(${artworkUrl})` }}
      />

      {/* Top Bar */}
      <header className="relative z-10 h-20 px-8 flex items-center justify-between border-b border-white/10 shrink-0">
        <div className="flex items-center gap-4">
          <div className="w-10 h-10 rounded-lg bg-rose-500/20 text-rose-400 flex items-center justify-center border border-rose-500/30">
            <Mic2 className="w-5 h-5" />
          </div>
          <div>
            <h2 className="text-lg font-bold text-white leading-none">{currentTrack.title}</h2>
            <p className="text-sm text-zinc-400 mt-1">{currentTrack.rawArtist || currentTrack.albumArtist}</p>
          </div>
        </div>

        <button
          onClick={onClose}
          className="w-10 h-10 rounded-full bg-white/10 hover:bg-white/20 text-white flex items-center justify-center transition-colors cursor-pointer"
          title="Close Lyrics"
        >
          <X className="w-5 h-5" />
        </button>
      </header>

      {/* Lyrics Content Container */}
      <div className="relative z-10 flex-1 overflow-y-auto px-6 py-16 flex flex-col items-center select-none scroll-smooth">
        {loading ? (
          <div className="flex flex-col items-center justify-center m-auto text-zinc-400 gap-3">
            <div className="w-8 h-8 border-2 border-rose-500 border-t-transparent rounded-full animate-spin" />
            <p className="text-sm font-medium">Searching for synchronized lyrics...</p>
          </div>
        ) : lyrics?.isSynced && lyrics.lines.length > 0 ? (
          <div className="max-w-2xl w-full flex flex-col gap-6 py-20 text-center">
            {lyrics.lines.map((line, idx) => {
              const isActive = idx === activeLineIdx
              const isPast = idx < activeLineIdx
              return (
                <div
                  key={idx}
                  ref={(el) => {
                    lineRefs.current[idx] = el
                  }}
                  onClick={() => seek(line.time)}
                  className={`cursor-pointer transition-all duration-300 font-bold tracking-tight rounded-xl px-4 py-2 ${
                    isActive
                      ? 'text-white text-3xl md:text-4xl scale-105 filter drop-shadow-[0_0_20px_rgba(255,255,255,0.4)]'
                      : isPast
                      ? 'text-zinc-500 text-xl md:text-2xl hover:text-zinc-300'
                      : 'text-zinc-600 text-xl md:text-2xl hover:text-zinc-400'
                  }`}
                >
                  {line.text || '♪'}
                </div>
              )
            })}
          </div>
        ) : lyrics?.plainLyrics ? (
          <div className="max-w-xl w-full my-auto text-center">
            <pre className="font-sans text-xl md:text-2xl text-zinc-300 whitespace-pre-wrap leading-relaxed">
              {lyrics.plainLyrics}
            </pre>
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center m-auto text-zinc-500 gap-3">
            <Music2 className="w-12 h-12 stroke-1 text-zinc-600" />
            <p className="text-lg font-medium text-zinc-400">No lyrics available</p>
            <p className="text-xs text-zinc-600">You can add a .lrc file in the track folder to show lyrics</p>
          </div>
        )}
      </div>
    </div>
  )
}
