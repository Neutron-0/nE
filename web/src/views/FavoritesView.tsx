import React, { useEffect, useState } from 'react'
import { Heart, Play, Clock, Music } from 'lucide-react'
import type { Track } from '../types'
import { api } from '../lib/api'
import { usePlayerStore } from '../store/playerStore'

export const FavoritesView: React.FC = () => {
  const [tracks, setTracks] = useState<Track[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const { playTrack, currentTrack, isPlaying } = usePlayerStore()

  useEffect(() => {
    loadFavorites()
  }, [])

  const loadFavorites = async () => {
    try {
      setIsLoading(true)
      const res = await api.getFavorites()
      setTracks(res.items)
    } catch (err) {
      console.error('Failed to load favorites:', err)
    } finally {
      setIsLoading(false)
    }
  }

  const formatDuration = (secs: number) => {
    if (!secs || isNaN(secs)) return '0:00'
    const m = Math.floor(secs / 60)
    const s = Math.floor(secs % 60)
    return `${m}:${s < 10 ? '0' : ''}${s}`
  }

  return (
    <div className="p-8">
      {/* Header */}
      <div className="flex items-center gap-4 mb-8 pb-6 border-b border-zinc-800/60">
        <div className="w-20 h-20 rounded-2xl bg-gradient-to-tr from-rose-500 to-amber-500 flex items-center justify-center text-white shadow-xl shadow-rose-500/20">
          <Heart className="w-10 h-10 fill-current" />
        </div>
        <div>
          <span className="text-xs font-semibold uppercase tracking-wider text-rose-400 font-mono">
            Collection
          </span>
          <h2 className="text-3xl font-extrabold text-white tracking-tight mt-0.5">Starred Tracks</h2>
          <p className="text-xs text-zinc-400 mt-1 font-mono">{tracks.length} favorites</p>
        </div>
      </div>

      {/* Tracks Table */}
      {isLoading ? (
        <div className="p-8 text-center text-zinc-500 font-medium animate-pulse">
          Loading starred tracks...
        </div>
      ) : tracks.length === 0 ? (
        <div className="p-12 text-center text-zinc-500">
          <Heart className="w-12 h-12 mx-auto mb-3 text-zinc-700 stroke-1" />
          <h3 className="text-lg font-semibold text-zinc-400">No starred tracks</h3>
          <p className="text-sm text-zinc-600 mt-1">Star tracks while listening to save them here</p>
        </div>
      ) : (
        <div className="w-full">
          <div className="grid grid-cols-12 text-xs font-semibold text-zinc-500 uppercase tracking-wider pb-3 border-b border-zinc-800/40 px-3">
            <span className="col-span-1 text-center">#</span>
            <span className="col-span-5">Title</span>
            <span className="col-span-3">Artist</span>
            <span className="col-span-2">Album</span>
            <span className="col-span-1 text-right flex items-center justify-end gap-1">
              <Clock className="w-3.5 h-3.5" />
            </span>
          </div>

          <div className="divide-y divide-zinc-900/40">
            {tracks.map((track, idx) => {
              const isCurrent = currentTrack?.id === track.id
              return (
                <div
                  key={track.id}
                  onClick={() => playTrack(track, tracks)}
                  className={`grid grid-cols-12 items-center py-2.5 px-3 rounded-lg cursor-pointer transition-colors group ${
                    isCurrent
                      ? 'bg-zinc-800/80 text-rose-400'
                      : 'hover:bg-zinc-900/60 text-zinc-300'
                  }`}
                >
                  <div className="col-span-1 text-center text-xs font-mono text-zinc-500 group-hover:text-white">
                    {isCurrent && isPlaying ? (
                      <Music className="w-3.5 h-3.5 mx-auto animate-pulse text-rose-400" />
                    ) : (
                      <span className="group-hover:hidden">{idx + 1}</span>
                    )}
                    <Play className="w-3.5 h-3.5 mx-auto fill-current hidden group-hover:inline-block" />
                  </div>

                  <div className="col-span-5 min-w-0 pr-4">
                    <p className={`text-sm font-medium truncate ${isCurrent ? 'text-rose-400 font-semibold' : 'text-white'}`}>
                      {track.title}
                    </p>
                  </div>

                  <div className="col-span-3 min-w-0 pr-4 text-xs text-zinc-400 truncate">
                    {track.rawArtist || track.albumArtist || 'Unknown Artist'}
                  </div>

                  <div className="col-span-2 min-w-0 pr-4 text-xs text-zinc-500 truncate">
                    {track.albumTitle || 'Unknown Album'}
                  </div>

                  <div className="col-span-1 text-right text-xs font-mono text-zinc-500">
                    {formatDuration(track.duration)}
                  </div>
                </div>
              )
            })}
          </div>
        </div>
      )}
    </div>
  )
}
