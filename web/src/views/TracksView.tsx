import React, { useEffect, useState } from 'react'
import { Music, Clock, Play } from 'lucide-react'
import type { Track } from '../types'
import { api } from '../lib/api'
import { usePlayerStore } from '../store/playerStore'

interface TracksViewProps {
  searchQuery: string
}

export const TracksView: React.FC<TracksViewProps> = ({ searchQuery }) => {
  const [tracks, setTracks] = useState<Track[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const { playTrack, currentTrack, isPlaying } = usePlayerStore()

  useEffect(() => {
    loadTracks()
  }, [])

  const loadTracks = async () => {
    try {
      setIsLoading(true)
      const res = await api.getTracks(200, 0)
      setTracks(res.items)
    } catch (err) {
      console.error('Error fetching tracks:', err)
    } finally {
      setIsLoading(false)
    }
  }

  const filteredTracks = tracks.filter((t) => {
    const q = searchQuery.toLowerCase()
    return (
      t.title.toLowerCase().includes(q) ||
      (t.rawArtist && t.rawArtist.toLowerCase().includes(q)) ||
      (t.albumTitle && t.albumTitle.toLowerCase().includes(q))
    )
  })

  const formatDuration = (secs: number) => {
    if (!secs || isNaN(secs)) return '0:00'
    const m = Math.floor(secs / 60)
    const s = Math.floor(secs % 60)
    return `${m}:${s < 10 ? '0' : ''}${s}`
  }

  if (isLoading) {
    return (
      <div className="p-8 flex items-center justify-center h-64 text-zinc-500 font-medium animate-pulse">
        Loading tracks...
      </div>
    )
  }

  if (filteredTracks.length === 0) {
    return (
      <div className="p-12 text-center text-zinc-500">
        <Music className="w-12 h-12 mx-auto mb-3 text-zinc-700 stroke-1" />
        <h3 className="text-lg font-semibold text-zinc-400">No tracks found</h3>
        <p className="text-sm text-zinc-600 mt-1">Add audio files to your library and scan</p>
      </div>
    )
  }

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-bold text-white tracking-tight">Tracks</h2>
        <span className="text-xs text-zinc-500 font-mono">{filteredTracks.length} tracks</span>
      </div>

      <div className="w-full">
        <div className="grid grid-cols-12 text-xs font-semibold text-zinc-500 uppercase tracking-wider pb-3 border-b border-zinc-800/40 px-3">
          <span className="col-span-1 text-center">#</span>
          <span className="col-span-4">Title</span>
          <span className="col-span-3">Artist</span>
          <span className="col-span-3">Album</span>
          <span className="col-span-1 text-right flex items-center justify-end gap-1">
            <Clock className="w-3.5 h-3.5" />
          </span>
        </div>

        <div className="divide-y divide-zinc-900/40">
          {filteredTracks.map((track, idx) => {
            const isCurrent = currentTrack?.id === track.id
            return (
              <div
                key={track.id}
                onClick={() => playTrack(track, filteredTracks)}
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

                <div className="col-span-4 min-w-0 pr-4">
                  <p className={`text-sm font-medium truncate ${isCurrent ? 'text-rose-400 font-semibold' : 'text-white'}`}>
                    {track.title}
                  </p>
                  <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-zinc-800 text-zinc-400 uppercase">
                    {track.format}
                  </span>
                </div>

                <div className="col-span-3 min-w-0 pr-4 text-xs text-zinc-400 truncate">
                  {track.rawArtist || track.albumArtist || 'Unknown Artist'}
                </div>

                <div className="col-span-3 min-w-0 pr-4 text-xs text-zinc-500 truncate">
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
    </div>
  )
}
