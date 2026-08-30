import React, { useState } from 'react'
import {
  Play,
  Pause,
  SkipBack,
  SkipForward,
  Shuffle,
  Repeat,
  Repeat1,
  Volume2,
  VolumeX,
  Music,
  Heart,
} from 'lucide-react'
import { usePlayerStore } from '../store/playerStore'
import { api } from '../lib/api'

export const PlayerBar: React.FC = () => {
  const {
    currentTrack,
    isPlaying,
    currentTime,
    duration,
    volume,
    isMuted,
    isShuffle,
    repeatMode,
    togglePlay,
    next,
    previous,
    seek,
    setVolume,
    toggleMute,
    toggleShuffle,
    toggleRepeat,
    toggleStarCurrentTrack,
  } = usePlayerStore()

  const [imgError, setImgError] = useState(false)

  if (!currentTrack) {
    return (
      <footer className="h-20 bg-zinc-950/90 border-t border-zinc-800/80 px-6 flex items-center justify-between text-zinc-500 text-sm">
        <div className="flex items-center gap-3">
          <div className="w-12 h-12 rounded-lg bg-zinc-900 border border-zinc-800 flex items-center justify-center">
            <Music className="w-5 h-5 text-zinc-600" />
          </div>
          <div>
            <p className="font-medium text-zinc-400">No track selected</p>
            <p className="text-xs text-zinc-600">Select an album or track to start playing</p>
          </div>
        </div>
      </footer>
    )
  }

  const formatTime = (secs: number) => {
    if (isNaN(secs) || secs < 0) return '0:00'
    const m = Math.floor(secs / 60)
    const s = Math.floor(secs % 60)
    return `${m}:${s < 10 ? '0' : ''}${s}`
  }

  const progressPercent = duration > 0 ? (currentTime / duration) * 100 : 0
  const artworkUrl = api.getArtworkUrl('track', currentTrack.id, 120)

  return (
    <footer className="h-20 bg-zinc-950/95 border-t border-zinc-800/80 px-6 flex items-center justify-between gap-6 backdrop-blur-xl shrink-0 z-50">
      {/* Left: Track Details with Real Artwork */}
      <div className="flex items-center gap-3.5 min-w-0 w-1/4">
        <div className="w-12 h-12 rounded-lg bg-zinc-900 border border-zinc-700/60 overflow-hidden shrink-0 shadow-md relative flex items-center justify-center">
          {!imgError ? (
            <img
              src={artworkUrl}
              alt="Cover"
              onError={() => setImgError(true)}
              className="w-full h-full object-cover"
            />
          ) : (
            <Music className="w-5 h-5 text-rose-400" />
          )}
        </div>
        <div className="min-w-0 flex-1">
          <p className="text-sm font-semibold text-white truncate hover:underline cursor-pointer">
            {currentTrack.title}
          </p>
          <p className="text-xs text-zinc-400 truncate hover:underline cursor-pointer">
            {currentTrack.rawArtist || currentTrack.albumArtist || 'Unknown Artist'}
          </p>
        </div>
        <button
          onClick={toggleStarCurrentTrack}
          title={currentTrack.isStarred ? 'Unstar' : 'Star'}
          className={`p-1.5 rounded transition-colors ${
            currentTrack.isStarred ? 'text-rose-500 fill-rose-500' : 'text-zinc-500 hover:text-zinc-300'
          }`}
        >
          <Heart className={`w-4 h-4 ${currentTrack.isStarred ? 'fill-current' : ''}`} />
        </button>
      </div>

      {/* Center: Controls & Seekbar */}
      <div className="flex flex-col items-center gap-1.5 w-2/4 max-w-xl">
        {/* Buttons */}
        <div className="flex items-center gap-4">
          <button
            onClick={toggleShuffle}
            title="Shuffle"
            className={`p-1.5 rounded transition-colors ${
              isShuffle ? 'text-rose-400' : 'text-zinc-400 hover:text-zinc-200'
            }`}
          >
            <Shuffle className="w-4 h-4" />
          </button>

          <button
            onClick={previous}
            title="Previous"
            className="p-1.5 text-zinc-300 hover:text-white transition-colors"
          >
            <SkipBack className="w-5 h-5 fill-current" />
          </button>

          <button
            onClick={togglePlay}
            title={isPlaying ? 'Pause' : 'Play'}
            className="w-9 h-9 rounded-full bg-white text-black flex items-center justify-center shadow-lg hover:scale-105 transition-all"
          >
            {isPlaying ? (
              <Pause className="w-4 h-4 fill-current" />
            ) : (
              <Play className="w-4 h-4 fill-current ml-0.5" />
            )}
          </button>

          <button
            onClick={next}
            title="Next"
            className="p-1.5 text-zinc-300 hover:text-white transition-colors"
          >
            <SkipForward className="w-5 h-5 fill-current" />
          </button>

          <button
            onClick={toggleRepeat}
            title={`Repeat: ${repeatMode}`}
            className={`p-1.5 rounded transition-colors ${
              repeatMode !== 'off' ? 'text-rose-400' : 'text-zinc-400 hover:text-zinc-200'
            }`}
          >
            {repeatMode === 'one' ? (
              <Repeat1 className="w-4 h-4" />
            ) : (
              <Repeat className="w-4 h-4" />
            )}
          </button>
        </div>

        {/* Progress Bar */}
        <div className="w-full flex items-center gap-2.5 text-xs text-zinc-400 font-mono">
          <span className="w-10 text-right">{formatTime(currentTime)}</span>
          <div
            onClick={(e) => {
              const rect = e.currentTarget.getBoundingClientRect()
              const pos = (e.clientX - rect.left) / rect.width
              seek(pos * duration)
            }}
            className="relative flex-1 h-1.5 bg-zinc-800 rounded-full cursor-pointer group py-1 -my-1"
          >
            <div
              className="absolute top-1 left-0 h-1.5 bg-zinc-200 group-hover:bg-rose-500 rounded-full transition-all"
              style={{ width: `${progressPercent}%` }}
            />
          </div>
          <span className="w-10">{formatTime(duration)}</span>
        </div>
      </div>

      {/* Right: Volume Controls */}
      <div className="flex items-center justify-end gap-2.5 w-1/4">
        <button
          onClick={toggleMute}
          title={isMuted ? 'Unmute' : 'Mute'}
          className="text-zinc-400 hover:text-zinc-200 p-1"
        >
          {isMuted || volume === 0 ? (
            <VolumeX className="w-4 h-4" />
          ) : (
            <Volume2 className="w-4 h-4" />
          )}
        </button>
        <input
          type="range"
          min="0"
          max="1"
          step="0.01"
          value={isMuted ? 0 : volume}
          onChange={(e) => setVolume(parseFloat(e.target.value))}
          className="w-24 h-1 bg-zinc-800 accent-rose-500 cursor-pointer rounded-full"
        />
      </div>
    </footer>
  )
}
