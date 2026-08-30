import React, { useEffect, useState } from 'react'
import { Disc } from 'lucide-react'
import type { Album, Track } from '../types'
import { api } from '../lib/api'
import { HeroAlbumCard } from '../components/HeroAlbumCard'
import { TracklistTable } from '../components/TracklistTable'

interface AlbumsViewProps {
  onSelectAlbum: (album: Album) => void
  searchQuery: string
}

export const AlbumsView: React.FC<AlbumsViewProps> = ({ onSelectAlbum, searchQuery }) => {
  const [albums, setAlbums] = useState<Album[]>([])
  const [featuredTracks, setFeaturedTracks] = useState<Track[]>([])
  const [isLoading, setIsLoading] = useState(true)

  useEffect(() => {
    loadData()
  }, [])

  const loadData = async () => {
    try {
      setIsLoading(true)
      const [albumRes, trackRes] = await Promise.all([
        api.getAlbums(50, 0),
        api.getTracks(50, 0),
      ])
      setAlbums(albumRes.items || [])
      setFeaturedTracks(trackRes.items || [])
    } catch (err) {
      console.error('Error fetching albums/tracks:', err)
    } finally {
      setIsLoading(false)
    }
  }

  const filteredAlbums = albums.filter((a) => {
    const q = searchQuery.toLowerCase()
    return a.title.toLowerCase().includes(q) || a.albumArtist.toLowerCase().includes(q)
  })

  const featuredAlbum = albums.length > 0 ? albums[0] : null

  if (isLoading) {
    return (
      <div className="p-8 flex items-center justify-center h-64 text-zinc-500 font-hud">
        <p className="animate-pulse text-sm">INITIALIZING CATALOG TELEMETRY...</p>
      </div>
    )
  }

  return (
    <div className="p-6 md:p-8 space-y-8">
      {/* 1. Hero Showcase Card matching Image 2 */}
      {featuredAlbum && (
        <HeroAlbumCard
          album={featuredAlbum}
          tracks={featuredTracks}
          onSelectAlbum={onSelectAlbum}
        />
      )}

      {/* 2. Tracklist Section matching Image 2 */}
      <div>
        <div className="flex items-center justify-between mb-3 px-1">
          <h3 className="text-lg font-bold text-white tracking-tight">Songs</h3>
          <span className="text-xs text-zinc-400 font-hud">{featuredTracks.length} tracks</span>
        </div>
        <TracklistTable tracks={featuredTracks} />
      </div>

      {/* 3. Albums Grid */}
      {filteredAlbums.length > 1 && (
        <div className="pt-6 border-t border-white/5">
          <div className="flex items-center justify-between mb-4 px-1">
            <h3 className="text-lg font-bold text-white tracking-tight">Albums</h3>
            <span className="text-xs text-zinc-400 font-hud">{filteredAlbums.length} albums</span>
          </div>

          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-5">
            {filteredAlbums.map((album) => (
              <div
                key={album.id}
                onClick={() => onSelectAlbum(album)}
                className="group bg-[#211e1c] hover:bg-[#2a2623] border border-white/5 hover:border-white/15 rounded-2xl p-3.5 transition-all cursor-pointer shadow-lg hover:shadow-2xl hover:-translate-y-1"
              >
                <div className="relative aspect-square rounded-xl bg-[#171514] overflow-hidden mb-3 flex items-center justify-center border border-white/5">
                  <img
                    src={api.getArtworkUrl('album', album.id, 300)}
                    alt={album.title}
                    loading="lazy"
                    onError={(e) => {
                      ;(e.target as HTMLElement).style.display = 'none'
                    }}
                    className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-300"
                  />
                  <Disc className="w-12 h-12 text-zinc-700 stroke-1 absolute -z-10" />
                </div>

                <h4 className="font-bold text-sm text-zinc-100 truncate group-hover:text-white transition-colors">
                  {album.title}
                </h4>
                <p className="text-xs text-zinc-400 truncate mt-0.5">{album.albumArtist}</p>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
