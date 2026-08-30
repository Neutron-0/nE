import React, { useEffect, useState } from 'react'
import { Disc, Play, Clock, ArrowLeft, Music, Heart } from 'lucide-react'
import type { Album, Track } from '../types'
import { api } from '../lib/api'
import { usePlayerStore } from '../store/playerStore'

interface AlbumDetailViewProps {
  album: Album
  onBack: () => void
}

export const AlbumDetailView: React.FC<AlbumDetailViewProps> = ({ album, onBack }) => {
  const [tracks, setTracks] = useState<Track[]>(album.tracks || [])
  const [isLoading, setIsLoading] = useState(!album.tracks)
  const [isStarred, setIsStarred] = useState(album.isStarred || false)
  const { playTrack, currentTrack, isPlaying } = usePlayerStore()

  useEffect(() => {
    if (!album.tracks) {
      loadFullAlbum()
    }
  }, [album.id])

  const loadFullAlbum = async () => {
    try {
      setIsLoading(true)
      const res = await api.getAlbum(album.id)
      setTracks(res.tracks || [])
      setIsStarred(res.isStarred || false)
    } catch (err) {
      console.error('Error fetching album details:', err)
    } finally {
      setIsLoading(false)
    }
  }

  const handleToggleStarAlbum = async () => {
    const nextVal = !isStarred
    setIsStarred(nextVal)
    try {
      await api.starItem('album', album.id, nextVal)
    } catch (err) {
      console.error('Star failed:', err)
    }
  }

  const handleToggleStarTrack = async (e: React.MouseEvent, track: Track) => {
    e.stopPropagation()
    const nextVal = !track.isStarred
    setTracks(tracks.map((t) => (t.id === track.id ? { ...t, isStarred: nextVal } : t)))
    try {
      await api.starItem('track', track.id, nextVal)
    } catch (err) {
      console.error('Star failed:', err)
    }
  }

  const formatDuration = (secs: number) => {
    if (!secs || isNaN(secs)) return '0:00'
    const m = Math.floor(secs / 60)
    const s = Math.floor(secs % 60)
    return `${m}:${s < 10 ? '0' : ''}${s}`
  }

  const totalDurationMinutes = Math.round(album.duration / 60)

  return (
    <div className="p-8">
      {/* Back Button */}
      <button
        onClick={onBack}
        className="flex items-center gap-2 text-xs font-semibold text-zinc-400 hover:text-white mb-6 transition-colors"
      >
        <ArrowLeft className="w-4 h-4" /> Back to Albums
      </button>

      {/* Album Header Hero */}
      <div className="flex flex-col md:flex-row items-center md:items-end gap-6 mb-8 pb-8 border-b border-zinc-800/60">
        <div className="w-48 h-48 rounded-2xl bg-zinc-900 border border-zinc-700/60 overflow-hidden shrink-0 shadow-2xl flex items-center justify-center relative">
          <img
            src={api.getArtworkUrl('album', album.id, 500)}
            alt={album.title}
            onError={(e) => {
              (e.target as HTMLElement).style.display = 'none'
            }}
            className="w-full h-full object-cover"
          />
          <Disc className="w-24 h-24 text-zinc-700 stroke-1 absolute -z-10" />
        </div>

        <div className="flex-1 text-center md:text-left min-w-0">
          <span className="text-xs font-semibold uppercase tracking-wider text-rose-400 font-mono">
            {album.isCompilation ? 'Compilation' : 'Album'}
          </span>
          <h1 className="text-3xl sm:text-4xl font-extrabold text-white tracking-tight mt-1 mb-2">
            {album.title}
          </h1>
          <p className="text-base font-medium text-zinc-300 mb-4">{album.albumArtist}</p>

          <div className="flex items-center justify-center md:justify-start gap-3 text-xs text-zinc-400 font-mono">
            {album.year > 0 && <span>{album.year}</span>}
            {album.year > 0 && <span>•</span>}
            <span>{tracks.length} tracks</span>
            {totalDurationMinutes > 0 && <span>•</span>}
            {totalDurationMinutes > 0 && <span>{totalDurationMinutes} mins</span>}
          </div>

          <div className="mt-5 flex items-center justify-center md:justify-start gap-3">
            <button
              onClick={() => tracks.length > 0 && playTrack(tracks[0], tracks)}
              disabled={tracks.length === 0}
              className="inline-flex items-center gap-2 px-5 py-2.5 rounded-full bg-rose-500 hover:bg-rose-600 text-white font-semibold text-sm shadow-lg shadow-rose-500/25 hover:scale-105 transition-all disabled:opacity-50"
            >
              <Play className="w-4 h-4 fill-current ml-0.5" /> Play Album
            </button>

            <button
              onClick={handleToggleStarAlbum}
              className={`p-2.5 rounded-full border transition-all ${
                isStarred
                  ? 'bg-rose-500/10 border-rose-500/40 text-rose-500'
                  : 'bg-zinc-900 border-zinc-800 text-zinc-400 hover:text-white'
              }`}
            >
              <Heart className={`w-4 h-4 ${isStarred ? 'fill-current' : ''}`} />
            </button>
          </div>
        </div>
      </div>

      {/* Tracklist Table */}
      {isLoading ? (
        <div className="p-8 text-center text-zinc-500 font-medium animate-pulse">
          Loading tracklist...
        </div>
      ) : (
        <div className="w-full">
          <div className="grid grid-cols-12 text-xs font-semibold text-zinc-500 uppercase tracking-wider pb-3 border-b border-zinc-800/40 px-3">
            <span className="col-span-1 text-center">#</span>
            <span className="col-span-8">Title</span>
            <span className="col-span-2 text-center">Favorite</span>
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
                  className={`grid grid-cols-12 items-center py-3 px-3 rounded-lg cursor-pointer transition-colors group ${
                    isCurrent
                      ? 'bg-zinc-800/80 text-rose-400'
                      : 'hover:bg-zinc-900/60 text-zinc-300'
                  }`}
                >
                  <div className="col-span-1 text-center text-xs font-mono text-zinc-500 group-hover:text-white">
                    {isCurrent && isPlaying ? (
                      <Music className="w-3.5 h-3.5 mx-auto animate-pulse text-rose-400" />
                    ) : (
                      <span>{track.trackNumber || idx + 1}</span>
                    )}
                  </div>

                  <div className="col-span-8 min-w-0 pr-4">
                    <p className={`text-sm font-medium truncate ${isCurrent ? 'text-rose-400 font-semibold' : 'text-white'}`}>
                      {track.title}
                    </p>
                    <p className="text-xs text-zinc-500 truncate mt-0.5">
                      {track.rawArtist || album.albumArtist}
                    </p>
                  </div>

                  <div className="col-span-2 text-center">
                    <button
                      onClick={(e) => handleToggleStarTrack(e, track)}
                      className={`p-1 text-zinc-600 hover:text-rose-500 transition-colors ${
                        track.isStarred ? 'text-rose-500' : 'opacity-0 group-hover:opacity-100'
                      }`}
                    >
                      <Heart className={`w-3.5 h-3.5 ${track.isStarred ? 'fill-current' : ''}`} />
                    </button>
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
