import React, { useEffect, useState } from 'react'
import { Disc, Play } from 'lucide-react'
import type { Album } from '../types'
import { api } from '../lib/api'
import { usePlayerStore } from '../store/playerStore'

interface AlbumsViewProps {
  onSelectAlbum: (album: Album) => void
  searchQuery: string
}

export const AlbumsView: React.FC<AlbumsViewProps> = ({ onSelectAlbum, searchQuery }) => {
  const [albums, setAlbums] = useState<Album[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const { playTrack } = usePlayerStore()

  useEffect(() => {
    loadAlbums()
  }, [])

  const loadAlbums = async () => {
    try {
      setIsLoading(true)
      const res = await api.getAlbums(100, 0)
      setAlbums(res.items)
    } catch (err) {
      console.error('Error fetching albums:', err)
    } finally {
      setIsLoading(false)
    }
  }

  const filteredAlbums = albums.filter((a) => {
    const q = searchQuery.toLowerCase()
    return a.title.toLowerCase().includes(q) || a.albumArtist.toLowerCase().includes(q)
  })

  const handlePlayAlbum = async (e: React.MouseEvent, album: Album) => {
    e.stopPropagation()
    try {
      const fullAlbum = await api.getAlbum(album.id)
      if (fullAlbum.tracks && fullAlbum.tracks.length > 0) {
        playTrack(fullAlbum.tracks[0], fullAlbum.tracks)
      }
    } catch (err) {
      console.error('Error playing album:', err)
    }
  }

  if (isLoading) {
    return (
      <div className="p-8 flex items-center justify-center h-64 text-zinc-500">
        <p className="animate-pulse font-medium">Loading albums...</p>
      </div>
    )
  }

  if (filteredAlbums.length === 0) {
    return (
      <div className="p-12 text-center text-zinc-500">
        <Disc className="w-12 h-12 mx-auto mb-3 text-zinc-700 stroke-1" />
        <h3 className="text-lg font-semibold text-zinc-400">No albums found</h3>
        <p className="text-sm text-zinc-600 mt-1">Configure a music library in Settings and run a scan</p>
      </div>
    )
  }

  return (
    <div className="p-8">
      <div className="flex items-center justify-between mb-6">
        <h2 className="text-2xl font-bold text-white tracking-tight">Albums</h2>
        <span className="text-xs text-zinc-500 font-mono">{filteredAlbums.length} albums</span>
      </div>

      <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 xl:grid-cols-6 gap-6">
        {filteredAlbums.map((album) => (
          <div
            key={album.id}
            onClick={() => onSelectAlbum(album)}
            className="group bg-zinc-900/40 hover:bg-zinc-900 border border-zinc-800/40 hover:border-zinc-700/60 rounded-xl p-3.5 transition-all duration-200 cursor-pointer flex flex-col shadow-sm hover:shadow-xl hover:-translate-y-0.5"
          >
            {/* Real Artwork / Cover Container */}
            <div className="relative aspect-square rounded-lg bg-zinc-900 overflow-hidden mb-3 flex items-center justify-center border border-zinc-800 shadow-inner">
              <img
                src={api.getArtworkUrl('album', album.id, 300)}
                alt={album.title}
                loading="lazy"
                onError={(e) => {
                  (e.target as HTMLElement).style.display = 'none'
                }}
                className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
              />
              <Disc className="w-12 h-12 text-zinc-700 stroke-1 absolute -z-10" />

              {/* Play Button Overlay */}
              <button
                onClick={(e) => handlePlayAlbum(e, album)}
                className="absolute bottom-2.5 right-2.5 w-10 h-10 rounded-full bg-rose-500 hover:bg-rose-600 text-white flex items-center justify-center shadow-xl opacity-0 translate-y-2 group-hover:opacity-100 group-hover:translate-y-0 transition-all duration-200"
              >
                <Play className="w-4 h-4 fill-current ml-0.5" />
              </button>
            </div>

            {/* Album Metadata */}
            <div className="min-w-0 flex-1">
              <h4 className="font-semibold text-sm text-zinc-100 truncate group-hover:text-rose-400 transition-colors">
                {album.title}
              </h4>
              <p className="text-xs text-zinc-400 truncate mt-0.5">{album.albumArtist}</p>
              <div className="flex items-center gap-2 mt-2 text-[11px] text-zinc-500 font-mono">
                {album.year > 0 && <span>{album.year}</span>}
                {album.year > 0 && <span>•</span>}
                <span>{album.trackCount} tracks</span>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
